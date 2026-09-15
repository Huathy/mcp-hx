package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourname/mcp-x/internal/driver"
	"github.com/yourname/mcp-x/internal/safety"
)

func (s *Server) checkBucketKey(dsName, bucket, key string, requireBucket bool) error {
	for _, dc := range s.cfg.DataSources {
		if dc.Name != dsName {
			continue
		}
		if dc.Bucket == "" {
			if requireBucket {
				return fmt.Errorf("data source %s has no configured bucket; set 'bucket' in config before write operations", dsName)
			}
		} else if bucket != dc.Bucket {
			return fmt.Errorf("bucket %q not allowed (configured: %q)", bucket, dc.Bucket)
		}
		break
	}
	if strings.Contains(key, "..") {
		return fmt.Errorf("object key must not contain '..'")
	}
	return nil
}

func (s *Server) registerObjectTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "obj_list_buckets",
		Description: "List all buckets in an object store (MinIO/S3). Use this first to know available buckets.",
	}, s.handleObjListBuckets)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "obj_list",
		Description: "List objects in a bucket with optional prefix filter. Returns object keys, sizes, and content types.",
	}, s.handleObjList)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "obj_get",
		Description: "Download an object's content as text. Returns the raw content.",
	}, s.handleObjGet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "obj_put",
		Description: "Upload text content to an object. Requires write mode.",
	}, s.handleObjPut)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "obj_delete",
		Description: "Delete an object from a bucket. Requires write mode.",
	}, s.handleObjDelete)
}

func (s *Server) getObjectDriver(dsName string) (driver.ObjectStoreDriver, error) {
	d, err := s.mgr.Get(dsName)
	if err != nil {
		return nil, err
	}
	objDrv, ok := d.(driver.ObjectStoreDriver)
	if !ok {
		return nil, fmt.Errorf("data source %s is not an object store driver (type: %s)", dsName, d.Type())
	}
	return objDrv, nil
}

type objListBucketsInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
}

type objListBucketsOutput struct {
	Text string `json:"text" jsonschema:"bucket list as markdown"`
}

func (s *Server) handleObjListBuckets(ctx context.Context, req *mcp.CallToolRequest, in objListBucketsInput) (*mcp.CallToolResult, objListBucketsOutput, error) {
	d, err := s.getObjectDriver(in.DataSource)
	if err != nil {
		return nil, objListBucketsOutput{}, err
	}
	buckets, err := d.ListBuckets(ctx)
	if err != nil {
		return nil, objListBucketsOutput{}, fmt.Errorf("list buckets: %w", err)
	}
	if len(buckets) == 0 {
		return nil, objListBucketsOutput{Text: "## Buckets\n\nNo buckets found."}, nil
	}
	var sb strings.Builder
	sb.WriteString("## Buckets\n\n| # | Name |\n|---|------|\n")
	for i, b := range buckets {
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", i+1, b.Name))
	}
	return nil, objListBucketsOutput{Text: sb.String()}, nil
}

type objListInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Bucket     string `json:"bucket" jsonschema:"bucket name"`
	Prefix     string `json:"prefix,omitempty" jsonschema:"object key prefix filter"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max objects to return (default 100)"`
}

type objListOutput struct {
	Text string `json:"text" jsonschema:"object list as markdown"`
}

func (s *Server) handleObjList(ctx context.Context, req *mcp.CallToolRequest, in objListInput) (*mcp.CallToolResult, objListOutput, error) {
	d, err := s.getObjectDriver(in.DataSource)
	if err != nil {
		return nil, objListOutput{}, err
	}
	limit := in.Limit
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	objects, err := d.ListObjects(ctx, in.Bucket, in.Prefix, limit)
	if err != nil {
		return nil, objListOutput{}, fmt.Errorf("list objects: %w", err)
	}
	if len(objects) == 0 {
		return nil, objListOutput{Text: "## Objects\n\nNo objects found."}, nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Objects (%d found, limit %d)\n\n", len(objects), limit))
	sb.WriteString("| # | Key | Size | Type | Last Modified |\n|---|-----|------|------|----------------|\n")
	for i, o := range objects {
		sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s | %s |\n", i+1, o.Key, o.Size, o.ContentType, o.LastModified))
	}
	return nil, objListOutput{Text: sb.String()}, nil
}

type objGetInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Bucket     string `json:"bucket" jsonschema:"bucket name"`
	Key        string `json:"key" jsonschema:"object key"`
}

type objGetOutput struct {
	Text string `json:"text" jsonschema:"object content as text"`
}

func (s *Server) handleObjGet(ctx context.Context, req *mcp.CallToolRequest, in objGetInput) (*mcp.CallToolResult, objGetOutput, error) {
	d, err := s.getObjectDriver(in.DataSource)
	if err != nil {
		return nil, objGetOutput{}, err
	}
	if err := s.checkBucketKey(in.DataSource, in.Bucket, in.Key, false); err != nil {
		return nil, objGetOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	getCtx, cancel := chk.WrapContext(ctx)
	defer cancel()
	content, err := d.GetObject(getCtx, in.Bucket, in.Key)
	if err != nil {
		return nil, objGetOutput{}, fmt.Errorf("get object: %w", err)
	}
	text := fmt.Sprintf("## Object Content\n\n**Bucket**: %s\n**Key**: %s\n**Size**: %d bytes\n\n```\n%s\n```",
		in.Bucket, in.Key, len(content), content)
	return nil, objGetOutput{Text: text}, nil
}

type objPutInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Bucket     string `json:"bucket" jsonschema:"bucket name"`
	Key        string `json:"key" jsonschema:"object key"`
	Body       string `json:"body" jsonschema:"text content to upload"`
	ContType   string `json:"content_type,omitempty" jsonschema:"MIME content type (default application/octet-stream)"`
}

type objPutOutput struct {
	Text string `json:"text" jsonschema:"put result as text"`
}

func (s *Server) handleObjPut(ctx context.Context, req *mcp.CallToolRequest, in objPutInput) (*mcp.CallToolResult, objPutOutput, error) {
	d, err := s.getObjectDriver(in.DataSource)
	if err != nil {
		return nil, objPutOutput{}, err
	}
	if err := s.checkBucketKey(in.DataSource, in.Bucket, in.Key, true); err != nil {
		return nil, objPutOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, objPutOutput{}, err
	}
	if err := d.PutObject(ctx, in.Bucket, in.Key, in.ContType, []byte(in.Body)); err != nil {
		return nil, objPutOutput{}, fmt.Errorf("put object: %w", err)
	}
	text := fmt.Sprintf("## Put Result\n\n**Bucket**: %s\n**Key**: %s\n**Size**: %d bytes\n**Status**: OK",
		in.Bucket, in.Key, len(in.Body))
	return nil, objPutOutput{Text: text}, nil
}

type objDeleteInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Bucket     string `json:"bucket" jsonschema:"bucket name"`
	Key        string `json:"key" jsonschema:"object key"`
}

type objDeleteOutput struct {
	Text string `json:"text" jsonschema:"delete result as text"`
}

func (s *Server) handleObjDelete(ctx context.Context, req *mcp.CallToolRequest, in objDeleteInput) (*mcp.CallToolResult, objDeleteOutput, error) {
	d, err := s.getObjectDriver(in.DataSource)
	if err != nil {
		return nil, objDeleteOutput{}, err
	}
	if err := s.checkBucketKey(in.DataSource, in.Bucket, in.Key, true); err != nil {
		return nil, objDeleteOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, objDeleteOutput{}, err
	}
	if err := d.DeleteObject(ctx, in.Bucket, in.Key); err != nil {
		return nil, objDeleteOutput{}, fmt.Errorf("delete object: %w", err)
	}
	text := fmt.Sprintf("## Delete Result\n\n**Bucket**: %s\n**Key**: %s\n**Status**: OK", in.Bucket, in.Key)
	return nil, objDeleteOutput{Text: text}, nil
}
