package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/yourname/mcp-x/internal/driver"
)

type ESDriver struct {
	client *elasticsearch.Client
	index  string
}

func (d *ESDriver) Name() string            { return "elasticsearch" }
func (d *ESDriver) Type() driver.DriverType { return driver.DriverTypeNoSQL }

func (d *ESDriver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	addrs := cfg.Addrs
	if len(addrs) == 0 && cfg.Endpoint != "" {
		addrs = []string{cfg.Endpoint}
	}
	cli, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: addrs,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return fmt.Errorf("es new client: %w", err)
	}
	d.client = cli
	d.index = cfg.IndexName
	return d.Ping(ctx)
}

func (d *ESDriver) ListIndices(ctx context.Context) ([]string, error) {
	res, err := d.client.Indices.Get([]string{"*"},
		d.client.Indices.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("es list indices: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("es list indices: %s", res.String())
	}
	body, _ := io.ReadAll(res.Body)
	var indices map[string]json.RawMessage
	if err := json.Unmarshal(body, &indices); err != nil {
		return nil, fmt.Errorf("es parse indices: %w", err)
	}
	names := make([]string, 0, len(indices))
	for name := range indices {
		names = append(names, name)
	}
	return names, nil
}

func (d *ESDriver) Search(ctx context.Context, index, query string, limit int) ([]driver.DocInfo, error) {
	if limit <= 0 {
		limit = 100
	}
	if index == "" {
		index = d.index
	}
	var queryObj any
	if query == "" {
		queryObj = map[string]any{"match_all": map[string]any{}}
	} else if err := json.Unmarshal([]byte(query), &queryObj); err != nil {
		return nil, fmt.Errorf("es query not valid JSON: %w", err)
	}
	bodyObj := map[string]any{"size": limit, "query": queryObj}
	bodyBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return nil, fmt.Errorf("es marshal body: %w", err)
	}
	res, err := d.client.Search(
		d.client.Search.WithContext(ctx),
		d.client.Search.WithIndex(index),
		d.client.Search.WithBody(bytes.NewReader(bodyBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("es search: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("es search: %s", res.String())
	}

	raw, _ := io.ReadAll(res.Body)
	var sr struct {
		Hits struct {
			Hits []struct {
				ID     string         `json:"_id"`
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, fmt.Errorf("es parse search: %w", err)
	}
	docs := make([]driver.DocInfo, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		docs = append(docs, driver.DocInfo{ID: h.ID, Source: h.Source})
	}
	return docs, nil
}

func (d *ESDriver) GetDoc(ctx context.Context, index, id string) (*driver.DocInfo, error) {
	if index == "" {
		index = d.index
	}
	res, err := d.client.Get(index, id, d.client.Get.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("es get: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("es get: %s", res.String())
	}
	raw, _ := io.ReadAll(res.Body)
	var gr struct {
		ID     string         `json:"_id"`
		Found  bool           `json:"found"`
		Source map[string]any `json:"_source"`
	}
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, fmt.Errorf("es parse get: %w", err)
	}
	if !gr.Found {
		return nil, nil
	}
	return &driver.DocInfo{ID: gr.ID, Source: gr.Source}, nil
}

func (d *ESDriver) IndexDoc(ctx context.Context, index, id, body string) error {
	if index == "" {
		index = d.index
	}
	var doc any
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		return fmt.Errorf("es doc body not valid JSON: %w", err)
	}
	cleanBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("es remarshal body: %w", err)
	}
	res, err := d.client.Index(index, bytes.NewReader(cleanBytes),
		d.client.Index.WithContext(ctx),
		d.client.Index.WithDocumentID(id),
		d.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("es index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index: %s", res.String())
	}
	return nil
}

func (d *ESDriver) DeleteDoc(ctx context.Context, index, id string) error {
	if index == "" {
		index = d.index
	}
	res, err := d.client.Delete(index, id, d.client.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("es delete: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es delete: %s", res.String())
	}
	return nil
}

func (d *ESDriver) Ping(ctx context.Context) error {
	res, err := d.client.Ping(d.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es ping: %s", res.String())
	}
	return nil
}

func (d *ESDriver) Close() error {
	return d.client.Close(context.Background())
}

func init() {
	driver.Register("elasticsearch", func() driver.AnyDriver { return &ESDriver{} })
}
