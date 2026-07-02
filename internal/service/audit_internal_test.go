package service

import (
	"context"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type capturedQuery struct {
	SQL  string
	Vars []any
}

func newAuditDryRunMySQLDB(t *testing.T) (*gorm.DB, *capturedQuery) {
	t.Helper()

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("open dry-run mysql db: %v", err)
	}
	captured := &capturedQuery{}
	err = gdb.Callback().Query().After("gorm:query").Register("test:capture_audit_query", func(tx *gorm.DB) {
		captured.SQL = tx.Statement.SQL.String()
		captured.Vars = append([]any(nil), tx.Statement.Vars...)
	})
	if err != nil {
		t.Fatalf("register capture callback: %v", err)
	}
	return gdb, captured
}

func TestListOrgAuditLogEscapesPhraseLikeWildcards(t *testing.T) {
	gdb, captured := newAuditDryRunMySQLDB(t)
	svc := &Service{DB: gdb}

	_, err := svc.ListOrgAuditLog(context.Background(), 42, AuditLogFilters{Phrase: `50%_off\`})
	if err != nil {
		t.Fatalf("ListOrgAuditLog dry run: %v", err)
	}

	sql := captured.SQL
	if sql == "" {
		t.Fatal("no query captured")
	}
	want := `%50\%\_off\\%`
	found := false
	for _, v := range captured.Vars {
		if s, ok := v.(string); ok && s == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected escaped LIKE arg %q in vars, got %#v (sql %q)", want, captured.Vars, sql)
	}
}

func TestListOrgAuditLogPlainPhraseStillSubstringMatch(t *testing.T) {
	gdb, captured := newAuditDryRunMySQLDB(t)
	svc := &Service{DB: gdb}

	_, err := svc.ListOrgAuditLog(context.Background(), 42, AuditLogFilters{Phrase: "add_member"})
	if err != nil {
		t.Fatalf("ListOrgAuditLog dry run: %v", err)
	}

	want := `%add\_member%`
	found := false
	for _, v := range captured.Vars {
		if s, ok := v.(string); ok && s == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected escaped LIKE arg %q in vars, got %#v", want, captured.Vars)
	}
}
