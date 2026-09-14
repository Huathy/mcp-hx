package driver

import "context"

type DriverType string

const (
	DriverTypeSQL   DriverType = "sql"
	DriverTypeNoSQL DriverType = "nosql"
)

type ConnConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int64
	Addr            string
	Addrs           []string
	Password        string
	DB              int
	Mode            string
	PoolSize        int
	Username        string
	Endpoint        string
	Endpoints       []string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	Region          string
	Bucket          string
	IndexName       string
	Database        string
}

type QueryResult struct {
	Columns []string
	Rows    [][]any
}

type ExecResult struct {
	AffectedRows int64
	LastInsertID int64
	Duration     int64
}

type TableInfo struct {
	Name string
}

type TableSchema struct {
	Table   string
	Columns []ColumnInfo
	Indexes []IndexInfo
}

type ColumnInfo struct {
	Name     string
	Type     string
	Nullable string
	Key      string
	Default  string
}

type IndexInfo struct {
	Name    string
	Columns []string
	Unique  bool
}

type Driver interface {
	Name() string
	Type() DriverType
	Connect(ctx context.Context, cfg ConnConfig) error
	Query(ctx context.Context, sql string, args []any) (*QueryResult, error)
	Execute(ctx context.Context, sql string, args []any) (*ExecResult, error)
	ListTables(ctx context.Context) ([]TableInfo, error)
	DescribeTable(ctx context.Context, table string) (*TableSchema, error)
	Ping(ctx context.Context) error
	Close() error
}

type NoSQLDriver interface {
	Name() string
	Type() DriverType
	Connect(ctx context.Context, cfg ConnConfig) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttlSeconds int) error
	Del(ctx context.Context, key string) (int64, error)
	Keys(ctx context.Context, pattern string, limit int) ([]string, error)
	KeyType(ctx context.Context, key string) (string, error)
	TTL(ctx context.Context, key string) (int64, error)
	Execute(ctx context.Context, command string, args ...string) (any, error)
	Ping(ctx context.Context) error
	Close() error
}

type AnyDriver interface {
	Name() string
	Type() DriverType
	Connect(ctx context.Context, cfg ConnConfig) error
	Ping(ctx context.Context) error
	Close() error
}

type DocInfo struct {
	ID     string
	Source map[string]any
}

type DocStoreDriver interface {
	Name() string
	Type() DriverType
	Connect(ctx context.Context, cfg ConnConfig) error
	ListIndices(ctx context.Context) ([]string, error)
	Search(ctx context.Context, index, query string, limit int) ([]DocInfo, error)
	GetDoc(ctx context.Context, index, id string) (*DocInfo, error)
	IndexDoc(ctx context.Context, index, id, body string) error
	DeleteDoc(ctx context.Context, index, id string) error
	Ping(ctx context.Context) error
	Close() error
}

type BucketInfo struct {
	Name        string
	ObjectCount int64
}

type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified string
}

type ObjectStoreDriver interface {
	Name() string
	Type() DriverType
	Connect(ctx context.Context, cfg ConnConfig) error
	ListBuckets(ctx context.Context) ([]BucketInfo, error)
	ListObjects(ctx context.Context, bucket, prefix string, limit int) ([]ObjectInfo, error)
	GetObject(ctx context.Context, bucket, key string) (string, error)
	PutObject(ctx context.Context, bucket, key, contentType string, body []byte) error
	DeleteObject(ctx context.Context, bucket, key string) error
	Ping(ctx context.Context) error
	Close() error
}
