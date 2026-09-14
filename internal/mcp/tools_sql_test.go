package mcp

import (
	"strings"
	"testing"

	"github.com/yourname/mcp-x/internal/config"
	"github.com/yourname/mcp-x/internal/driver"
	"github.com/yourname/mcp-x/internal/safety"
)

func newTestSafety(mode string) *safety.Checker {
	return safety.New(config.SafetyConfig{
		Mode:            mode,
		MaxRows:         1000,
		BlockedKeywords: []string{"DROP", "TRUNCATE"},
	})
}

func TestSafetyReadOnly(t *testing.T) {
	chk := newTestSafety("read-only")
	if err := chk.CheckWrite(); err == nil {
		t.Error("read-only should block writes")
	}
}

func TestSafetyReadWrite(t *testing.T) {
	chk := newTestSafety("read-write")
	if err := chk.CheckWrite(); err != nil {
		t.Error("read-write should allow writes")
	}
}

func TestSafetyBlockedKeyword(t *testing.T) {
	chk := newTestSafety("read-write")
	if err := chk.CheckSQL("DROP TABLE users"); err == nil {
		t.Error("DROP should be blocked")
	}
	if err := chk.CheckSQL("SELECT * FROM users"); err != nil {
		t.Error("SELECT should pass")
	}
}

func TestSafetyDeleteWithoutWhere(t *testing.T) {
	chk := newTestSafety("read-write")
	if err := chk.CheckSQL("DELETE FROM users"); err == nil {
		t.Error("DELETE without WHERE should be blocked")
	}
	if err := chk.CheckSQL("DELETE FROM users WHERE id=1"); err != nil {
		t.Error("DELETE with WHERE should pass")
	}
}

func TestSafetyUpdateWithoutWhere(t *testing.T) {
	chk := newTestSafety("read-write")
	if err := chk.CheckSQL("UPDATE users SET name='a'"); err == nil {
		t.Error("UPDATE without WHERE should be blocked")
	}
}

func TestFormatQueryResult(t *testing.T) {
	r := &driver.QueryResult{
		Columns: []string{"id", "name"},
		Rows:    [][]any{{1, "Alice"}, {2, "Bob"}},
	}
	text := formatQueryResult(r)
	if !strings.Contains(text, "Query Result (2 rows)") {
		t.Error("missing row count")
	}
	if !strings.Contains(text, "Alice") {
		t.Error("missing data")
	}
}
