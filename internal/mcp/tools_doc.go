package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourname/mcp-x/internal/driver"
	"github.com/yourname/mcp-x/internal/safety"
)

func (s *Server) registerDocTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "doc_list_indices",
		Description: "List all indices/collections in a document store (Elasticsearch indices). Use first to know available indices.",
	}, s.handleDocListIndices)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "doc_search",
		Description: "Search documents in an Elasticsearch index using a JSON query body. Returns matching documents with IDs and sources.",
	}, s.handleDocSearch)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "doc_get",
		Description: "Get a single document by ID from an Elasticsearch index.",
	}, s.handleDocGet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "doc_index",
		Description: "Index (create or replace) a document in an Elasticsearch index. Requires write mode.",
	}, s.handleDocIndex)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "doc_delete",
		Description: "Delete a document by ID from an Elasticsearch index. Requires write mode.",
	}, s.handleDocDelete)
}

func (s *Server) getDocDriver(dsName string) (driver.DocStoreDriver, error) {
	d, err := s.mgr.Get(dsName)
	if err != nil {
		return nil, err
	}
	docDrv, ok := d.(driver.DocStoreDriver)
	if !ok {
		return nil, fmt.Errorf("data source %s is not a document store driver (type: %s)", dsName, d.Type())
	}
	return docDrv, nil
}

type docListIndicesInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
}

type docListIndicesOutput struct {
	Text string `json:"text" jsonschema:"index list as markdown"`
}

func (s *Server) handleDocListIndices(ctx context.Context, req *mcp.CallToolRequest, in docListIndicesInput) (*mcp.CallToolResult, docListIndicesOutput, error) {
	d, err := s.getDocDriver(in.DataSource)
	if err != nil {
		return nil, docListIndicesOutput{}, err
	}
	indices, err := d.ListIndices(ctx)
	if err != nil {
		return nil, docListIndicesOutput{}, fmt.Errorf("list indices: %w", err)
	}
	if len(indices) == 0 {
		return nil, docListIndicesOutput{Text: "## Indices\n\nNo indices found."}, nil
	}
	var sb strings.Builder
	sb.WriteString("## Indices\n\n| # | Name |\n|---|------|\n")
	for i, idx := range indices {
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", i+1, idx))
	}
	return nil, docListIndicesOutput{Text: sb.String()}, nil
}

type docSearchInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Index      string `json:"index,omitempty" jsonschema:"Elasticsearch index name"`
	Query      string `json:"query,omitempty" jsonschema:"JSON query body, e.g. {\"query\":{\"match_all\":{}}}. Empty = match_all."`
	Limit      int    `json:"limit,omitempty" jsonschema:"max results (default 100)"`
}

type docSearchOutput struct {
	Text string `json:"text" jsonschema:"search results as markdown"`
}

func (s *Server) handleDocSearch(ctx context.Context, req *mcp.CallToolRequest, in docSearchInput) (*mcp.CallToolResult, docSearchOutput, error) {
	d, err := s.getDocDriver(in.DataSource)
	if err != nil {
		return nil, docSearchOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	limit := in.Limit
	if limit <= 0 || limit > chk.MaxRows() {
		limit = chk.MaxRows()
	}
	searchCtx, cancel := chk.WrapContext(ctx)
	defer cancel()
	docs, err := d.Search(searchCtx, in.Index, in.Query, limit)
	if err != nil {
		return nil, docSearchOutput{}, fmt.Errorf("search failed: %w", err)
	}
	if len(docs) == 0 {
		return nil, docSearchOutput{Text: "## Search Result\n\nNo documents found."}, nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Search Result (%d docs)\n\n", len(docs)))
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("### %d. ID: %s\n\n", i+1, doc.ID))
		bytes, _ := json.MarshalIndent(doc.Source, "", "  ")
		sb.WriteString("```json\n")
		sb.Write(bytes)
		sb.WriteString("\n```\n\n")
	}
	return nil, docSearchOutput{Text: sb.String()}, nil
}

type docGetInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Index      string `json:"index,omitempty" jsonschema:"Elasticsearch index name"`
	ID         string `json:"id" jsonschema:"document ID"`
}

type docGetOutput struct {
	Text string `json:"text" jsonschema:"document as markdown"`
}

func (s *Server) handleDocGet(ctx context.Context, req *mcp.CallToolRequest, in docGetInput) (*mcp.CallToolResult, docGetOutput, error) {
	d, err := s.getDocDriver(in.DataSource)
	if err != nil {
		return nil, docGetOutput{}, err
	}
	doc, err := d.GetDoc(ctx, in.Index, in.ID)
	if err != nil {
		return nil, docGetOutput{}, fmt.Errorf("get doc: %w", err)
	}
	if doc == nil {
		return nil, docGetOutput{Text: fmt.Sprintf("## Document\n\n**ID**: %s\n**Status**: not found", in.ID)}, nil
	}
	bytes, _ := json.MarshalIndent(doc.Source, "", "  ")
	text := fmt.Sprintf("## Document\n\n**Index**: %s\n**ID**: %s\n\n```json\n%s\n```", in.Index, in.ID, string(bytes))
	return nil, docGetOutput{Text: text}, nil
}

type docIndexInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Index      string `json:"index,omitempty" jsonschema:"Elasticsearch index name"`
	ID         string `json:"id" jsonschema:"document ID"`
	Body       string `json:"body" jsonschema:"JSON document body"`
}

type docIndexOutput struct {
	Text string `json:"text" jsonschema:"index result as text"`
}

func (s *Server) handleDocIndex(ctx context.Context, req *mcp.CallToolRequest, in docIndexInput) (*mcp.CallToolResult, docIndexOutput, error) {
	d, err := s.getDocDriver(in.DataSource)
	if err != nil {
		return nil, docIndexOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, docIndexOutput{}, err
	}
	if err := d.IndexDoc(ctx, in.Index, in.ID, in.Body); err != nil {
		return nil, docIndexOutput{}, fmt.Errorf("index doc: %w", err)
	}
	text := fmt.Sprintf("## Index Result\n\n**Index**: %s\n**ID**: %s\n**Status**: OK", in.Index, in.ID)
	return nil, docIndexOutput{Text: text}, nil
}

type docDeleteInput struct {
	DataSource string `json:"datasource" jsonschema:"the name of the data source"`
	Index      string `json:"index,omitempty" jsonschema:"Elasticsearch index name"`
	ID         string `json:"id" jsonschema:"document ID"`
}

type docDeleteOutput struct {
	Text string `json:"text" jsonschema:"delete result as text"`
}

func (s *Server) handleDocDelete(ctx context.Context, req *mcp.CallToolRequest, in docDeleteInput) (*mcp.CallToolResult, docDeleteOutput, error) {
	d, err := s.getDocDriver(in.DataSource)
	if err != nil {
		return nil, docDeleteOutput{}, err
	}
	chk := safety.New(s.cfg.EffectiveSafety(in.DataSource))
	if err := chk.CheckWrite(); err != nil {
		return nil, docDeleteOutput{}, err
	}
	if err := d.DeleteDoc(ctx, in.Index, in.ID); err != nil {
		return nil, docDeleteOutput{}, fmt.Errorf("delete doc: %w", err)
	}
	text := fmt.Sprintf("## Delete Result\n\n**Index**: %s\n**ID**: %s\n**Status**: OK", in.Index, in.ID)
	return nil, docDeleteOutput{Text: text}, nil
}
