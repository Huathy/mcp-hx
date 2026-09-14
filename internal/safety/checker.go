package safety

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yourname/mcp-x/internal/config"
)

type Checker struct {
	cfg config.SafetyConfig
}

func New(cfg config.SafetyConfig) *Checker {
	return &Checker{cfg: cfg}
}

func (c *Checker) IsReadOnly() bool {
	return c.cfg.Mode == "read-only"
}

func (c *Checker) CheckWrite() error {
	if c.IsReadOnly() {
		return fmt.Errorf("data source is read-only, write operations blocked")
	}
	return nil
}

func (c *Checker) CheckSQL(sqlStr string) error {
	upper := strings.ToUpper(strings.TrimSpace(sqlStr))

	if !isReadOnlySQL(upper) {
		if err := c.CheckWrite(); err != nil {
			return err
		}
	}

	for _, kw := range c.cfg.BlockedKeywords {
		kwUpper := strings.ToUpper(kw)
		if containsWord(upper, kwUpper) && !containsInList(c.cfg.AllowBlockedKeywords, kwUpper) {
			return fmt.Errorf("blocked keyword '%s' in SQL. Add to allow_blocked_keywords to override", kw)
		}
	}

	if err := checkDeleteWithoutWhere(upper); err != nil {
		return err
	}
	if err := checkUpdateWithoutWhere(upper); err != nil {
		return err
	}

	return nil
}

func (c *Checker) CheckRedisCommand(cmd string) error {
	upper := strings.ToUpper(strings.TrimSpace(cmd))
	for _, blocked := range c.cfg.BlockedCommands {
		if upper == strings.ToUpper(blocked) && !containsInList(c.cfg.AllowBlockedCommands, upper) {
			return fmt.Errorf("blocked Redis command: %s", upper)
		}
	}
	return nil
}

func (c *Checker) MaxRows() int {
	if c.cfg.MaxRows <= 0 {
		return 1000
	}
	return c.cfg.MaxRows
}

func (c *Checker) WrapContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.cfg.QueryTimeout.Std() <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.cfg.QueryTimeout.Std())
}

func isReadOnlySQL(upper string) bool {
	return strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "WITH")
}

func checkDeleteWithoutWhere(upper string) error {
	if !strings.HasPrefix(upper, "DELETE") {
		return nil
	}
	if !strings.Contains(upper, "WHERE") {
		return fmt.Errorf("DELETE without WHERE clause is blocked")
	}
	return nil
}

func checkUpdateWithoutWhere(upper string) error {
	if !strings.HasPrefix(upper, "UPDATE") {
		return nil
	}
	if !strings.Contains(upper, "WHERE") {
		return fmt.Errorf("UPDATE without WHERE clause is blocked")
	}
	return nil
}

func containsWord(text, word string) bool {
	re := regexp.MustCompile(`\b` + word + `\b`)
	return re.MatchString(strings.ToUpper(text))
}

func containsInList(list []string, s string) bool {
	for _, item := range list {
		if strings.ToUpper(item) == s {
			return true
		}
	}
	return false
}
