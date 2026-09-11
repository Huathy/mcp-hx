package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourname/mcp-dbx/internal/driver"
	"github.com/yourname/mcp-dbx/internal/safety"
)

func (s *Server) registerRedisTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_get",
		Description: "Get value of a Redis key. Returns value and type.",
	}, s.handleRedisGet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_set",
		Description: "Set a Redis key-value pair. Requires write mode.",
	}, s.handleRedisSet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_del",
		Description: "Delete a Redis key. Requires write mode.",
	}, s.handleRedisDel)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_keys",
		Description: "Scan Redis keys matching a pattern. Uses SCAN (non-blocking). Limited to 100 keys by default.",
	}, s.handleRedisKeys)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_type",
		Description: "Get the data type of a Redis key (string/hash/list/set/zset/stream).",
	}, s.handleRedisType)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "redis_ttl",
		Description: "Get remaining TTL of a Redis key in seconds. Returns -1 if no expiry, -2 if key doesn't exist.",
	}, s.handleRedisTTL)
}

type redisGetInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Key        string `json:"key" jsonschema:"Redis key"`
}

type redisGetOutput struct {
	Text string `json:"text" jsonschema:"value and type as text"`
}

func (s *Server) handleRedisGet(ctx context.Context, req *mcp.CallToolRequest, in redisGetInput) (*mcp.CallToolResult, redisGetOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisGetOutput{}, err
	}

	val, err := d.Get(ctx, in.Key)
	if err != nil {
		return nil, redisGetOutput{}, fmt.Errorf("redis get: %w", err)
	}
	typ, _ := d.KeyType(ctx, in.Key)

	text := fmt.Sprintf("## Redis Get\n\n**Key**: %s\n**Type**: %s\n**Value**: %s", in.Key, typ, val)
	return nil, redisGetOutput{Text: text}, nil
}

type redisSetInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Key        string `json:"key" jsonschema:"Redis key"`
	Value      string `json:"value" jsonschema:"Redis value"`
	TTL        int    `json:"ttl,omitempty" jsonschema:"TTL in seconds. 0 = no expiry."`
}

type redisSetOutput struct {
	Text string `json:"text" jsonschema:"set result as text"`
}

func (s *Server) handleRedisSet(ctx context.Context, req *mcp.CallToolRequest, in redisSetInput) (*mcp.CallToolResult, redisSetOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisSetOutput{}, err
	}

	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, redisSetOutput{}, err
	}

	if err := d.Set(ctx, in.Key, in.Value, in.TTL); err != nil {
		return nil, redisSetOutput{}, fmt.Errorf("redis set: %w", err)
	}

	ttlStr := "no expiry"
	if in.TTL > 0 {
		ttlStr = fmt.Sprintf("%ds", in.TTL)
	}
	text := fmt.Sprintf("## Redis Set\n\n**Key**: %s\n**TTL**: %s\n**Status**: OK", in.Key, ttlStr)
	return nil, redisSetOutput{Text: text}, nil
}

type redisDelInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Key        string `json:"key" jsonschema:"Redis key"`
}

type redisDelOutput struct {
	Text string `json:"text" jsonschema:"delete result as text"`
}

func (s *Server) handleRedisDel(ctx context.Context, req *mcp.CallToolRequest, in redisDelInput) (*mcp.CallToolResult, redisDelOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisDelOutput{}, err
	}

	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, redisDelOutput{}, err
	}

	n, err := d.Del(ctx, in.Key)
	if err != nil {
		return nil, redisDelOutput{}, fmt.Errorf("redis del: %w", err)
	}

	text := fmt.Sprintf("## Redis Del\n\n**Key**: %s\n**Deleted**: %d", in.Key, n)
	return nil, redisDelOutput{Text: text}, nil
}

type redisKeysInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Pattern    string `json:"pattern,omitempty" jsonschema:"Glob pattern, e.g. user:*"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max keys to return"`
}

type redisKeysOutput struct {
	Text string `json:"text" jsonschema:"key list as markdown"`
}

func (s *Server) handleRedisKeys(ctx context.Context, req *mcp.CallToolRequest, in redisKeysInput) (*mcp.CallToolResult, redisKeysOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisKeysOutput{}, err
	}

	pattern := in.Pattern
	if pattern == "" {
		pattern = "*"
	}
	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	keys, err := d.Keys(ctx, pattern, limit)
	if err != nil {
		return nil, redisKeysOutput{}, fmt.Errorf("redis keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, redisKeysOutput{Text: "## Redis Keys\n\nNo keys matched."}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Redis Keys (%d found, limit %d)\n\n", len(keys), limit))
	sb.WriteString("| # | Key |\n|---|-----|\n")
	for i, k := range keys {
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", i+1, k))
	}
	return nil, redisKeysOutput{Text: sb.String()}, nil
}

type redisTypeInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Key        string `json:"key" jsonschema:"Redis key"`
}

type redisTypeOutput struct {
	Text string `json:"text" jsonschema:"key type as text"`
}

func (s *Server) handleRedisType(ctx context.Context, req *mcp.CallToolRequest, in redisTypeInput) (*mcp.CallToolResult, redisTypeOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisTypeOutput{}, err
	}

	typ, err := d.KeyType(ctx, in.Key)
	if err != nil {
		return nil, redisTypeOutput{}, fmt.Errorf("redis type: %w", err)
	}

	text := fmt.Sprintf("## Redis Type\n\n**Key**: %s\n**Type**: %s", in.Key, typ)
	return nil, redisTypeOutput{Text: text}, nil
}

type redisTTLInput struct {
	DataSource string `json:"datasource" jsonschema:"Redis data source name"`
	Key        string `json:"key" jsonschema:"Redis key"`
}

type redisTTLOutput struct {
	Text string `json:"text" jsonschema:"TTL in seconds as text"`
}

func (s *Server) handleRedisTTL(ctx context.Context, req *mcp.CallToolRequest, in redisTTLInput) (*mcp.CallToolResult, redisTTLOutput, error) {
	d, err := s.getRedisDriver(in.DataSource)
	if err != nil {
		return nil, redisTTLOutput{}, err
	}

	ttl, err := d.TTL(ctx, in.Key)
	if err != nil {
		return nil, redisTTLOutput{}, fmt.Errorf("redis ttl: %w", err)
	}

	text := fmt.Sprintf("## Redis TTL\n\n**Key**: %s\n**TTL**: %d seconds", in.Key, ttl)
	return nil, redisTTLOutput{Text: text}, nil
}

func (s *Server) getRedisDriver(dsName string) (driver.NoSQLDriver, error) {
	d, err := s.mgr.Get(dsName)
	if err != nil {
		return nil, err
	}
	redisDrv, ok := d.(driver.NoSQLDriver)
	if !ok {
		return nil, fmt.Errorf("data source %s is not a NoSQL driver (type: %s)", dsName, d.Type())
	}
	return redisDrv, nil
}
