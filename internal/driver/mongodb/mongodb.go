package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yourname/mcp-x/internal/driver"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDriver struct {
	client     *mongo.Client
	database   string
	collection string
}

func (d *MongoDriver) Name() string            { return "mongodb" }
func (d *MongoDriver) Type() driver.DriverType { return driver.DriverTypeNoSQL }

func (d *MongoDriver) Connect(ctx context.Context, cfg driver.ConnConfig) error {
	uri := cfg.DSN
	if uri == "" && len(cfg.Addrs) > 0 {
		uri = "mongodb://"
		if cfg.Username != "" {
			uri += url.PathEscape(cfg.Username) + ":" + url.PathEscape(cfg.Password) + "@"
		}
		uri += joinAddrs(cfg.Addrs)
		uri += "/" + cfg.Database
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return fmt.Errorf("mongo ping: %w", err)
	}
	d.client = client
	d.database = cfg.Database
	d.collection = cfg.Bucket
	if d.collection == "" {
		d.collection = "mcp_x"
	}
	return nil
}

func (d *MongoDriver) coll() *mongo.Collection {
	return d.client.Database(d.database).Collection(d.collection)
}

func (d *MongoDriver) Get(ctx context.Context, key string) (string, error) {
	var doc bson.M
	err := d.coll().FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return "(nil)", nil
	}
	if err != nil {
		return "", fmt.Errorf("mongo get: %w", err)
	}
	bytes, err := bson.MarshalExtJSON(doc, false, false)
	if err != nil {
		return "", fmt.Errorf("mongo marshal: %w", err)
	}
	return string(bytes), nil
}

func (d *MongoDriver) Set(ctx context.Context, key, value string, ttlSeconds int) error {
	var doc bson.M
	if err := bson.UnmarshalExtJSON([]byte(value), true, &doc); err != nil {
		return fmt.Errorf("mongo unmarshal: %w", err)
	}
	doc["_id"] = key
	if ttlSeconds > 0 {
		doc["_expires_at"] = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}
	_, err := d.coll().UpdateOne(ctx, bson.M{"_id": key}, bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true))
	return err
}

func (d *MongoDriver) Del(ctx context.Context, key string) (int64, error) {
	res, err := d.coll().DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return 0, fmt.Errorf("mongo del: %w", err)
	}
	return res.DeletedCount, nil
}

func (d *MongoDriver) Keys(ctx context.Context, pattern string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	filter := bson.M{}
	if pattern != "" && pattern != "*" {
		filter["_id"] = bson.M{"$regex": globToRegex(pattern)}
	}
	cursor, err := d.coll().Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, fmt.Errorf("mongo keys: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []string
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		if id, ok := doc["_id"].(string); ok {
			keys = append(keys, id)
		}
	}
	return keys, cursor.Err()
}

func (d *MongoDriver) KeyType(ctx context.Context, key string) (string, error) {
	count, err := d.coll().CountDocuments(ctx, bson.M{"_id": key})
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "none", nil
	}
	return "document", nil
}

func (d *MongoDriver) TTL(ctx context.Context, key string) (int64, error) {
	var doc bson.M
	err := d.coll().FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return -2, nil
	}
	if err != nil {
		return 0, err
	}
	if _, ok := doc["_expires_at"]; !ok {
		return -1, nil
	}
	return -1, nil
}

func (d *MongoDriver) Execute(ctx context.Context, command string, args ...string) (any, error) {
	switch command {
	case "find":
		filter := bson.M{}
		if len(args) > 0 {
			_ = bson.UnmarshalExtJSON([]byte(args[0]), true, &filter)
		}
		cursor, err := d.coll().Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			return nil, err
		}
		return results, nil
	case "aggregate":
		pipeline := bson.A{}
		if len(args) > 0 {
			_ = bson.UnmarshalExtJSON([]byte(args[0]), true, &pipeline)
		}
		cursor, err := d.coll().Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			return nil, err
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported mongo command: %s (use find/aggregate)", command)
	}
}

func (d *MongoDriver) Ping(ctx context.Context) error {
	return d.client.Ping(ctx, nil)
}

func (d *MongoDriver) Close() error {
	if d.client == nil {
		return nil
	}
	return d.client.Disconnect(context.Background())
}

func joinAddrs(addrs []string) string {
	result := ""
	for i, a := range addrs {
		if i > 0 {
			result += ","
		}
		result += a
	}
	return result
}

func globToRegex(pattern string) string {
	var sb strings.Builder
	sb.WriteString("^")
	for _, r := range pattern {
		switch r {
		case '*':
			sb.WriteString(".*")
		case '?':
			sb.WriteString(".")
		default:
			sb.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	sb.WriteString("$")
	return sb.String()
}

func init() {
	driver.Register("mongodb", func() driver.AnyDriver { return &MongoDriver{} })
}
