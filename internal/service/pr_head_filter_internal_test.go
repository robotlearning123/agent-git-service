package service

import (
	"testing"

	"github.com/ngaut/agent-git-service/internal/db"
)

func TestApplyPRHeadFilterEscapesOwnerLikeWildcards(t *testing.T) {
	gdb, captured := newQueryCaptureDryRunMySQLDB(t)

	var prs []db.PullRequest
	q := applyPRHeadFilter(gdb.Model(&db.PullRequest{}), "own%er_x:main")
	if err := q.Find(&prs).Error; err != nil {
		t.Fatalf("dry-run find: %v", err)
	}

	want := `own\%er\_x/%`
	found := false
	for _, v := range captured.Vars {
		if s, ok := v.(string); ok && s == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected escaped LIKE owner arg %q in vars, got %#v (sql %q)", want, captured.Vars, captured.SQL)
	}
	for _, v := range captured.Vars {
		if s, ok := v.(string); ok && s == "main" {
			return
		}
	}
	t.Fatalf("expected head_ref arg %q in vars, got %#v", "main", captured.Vars)
}

func TestApplyPRHeadFilterBareBranchAndEmpty(t *testing.T) {
	gdb, captured := newQueryCaptureDryRunMySQLDB(t)

	var prs []db.PullRequest
	if err := applyPRHeadFilter(gdb.Model(&db.PullRequest{}), "feature-1").Find(&prs).Error; err != nil {
		t.Fatalf("dry-run find: %v", err)
	}
	found := false
	for _, v := range captured.Vars {
		if s, ok := v.(string); ok && s == "feature-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected bare head_ref arg %q in vars, got %#v", "feature-1", captured.Vars)
	}

	gdb2, captured2 := newQueryCaptureDryRunMySQLDB(t)
	if err := applyPRHeadFilter(gdb2.Model(&db.PullRequest{}), "  ").Find(&prs).Error; err != nil {
		t.Fatalf("dry-run find: %v", err)
	}
	for _, v := range captured2.Vars {
		if s, ok := v.(string); ok && s == "  " {
			t.Fatalf("blank head must not add a filter, got vars %#v", captured2.Vars)
		}
	}
}
