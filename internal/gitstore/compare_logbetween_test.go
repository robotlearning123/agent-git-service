package gitstore_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ngaut/agent-git-service/internal/gitstore"
)

// seedRepoWithTags creates a real git repo at <root>/<fullName>.git with two
// commits tagged v1 and v2, so LogBetweenTags can be exercised end to end.
func seedRepoWithTags(t *testing.T, root, fullName string) {
	t.Helper()
	dir := filepath.Join(root, fullName+".git")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "a.txt")
	run("commit", "-q", "-m", "first commit")
	run("tag", "v1")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "a.txt")
	run("commit", "-q", "-m", "second commit")
	run("tag", "v2")
}

func TestLogBetweenTags_ValidRevsReturnLog(t *testing.T) {
	root := t.TempDir()
	store, err := gitstore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	seedRepoWithTags(t, root, "user/repo")

	got, err := store.LogBetweenTags(context.Background(), "user/repo", "v1", "v2")
	if err != nil {
		t.Fatalf("LogBetweenTags valid revs: %v", err)
	}
	if !strings.Contains(got, "second commit") {
		t.Fatalf("expected log to include the second commit, got %q", got)
	}
	if strings.Contains(got, "first commit") {
		t.Fatalf("v1..v2 log should exclude the first commit, got %q", got)
	}
}

// TestLogBetweenTags_RejectsFlagInjection is the regression guard: a revision
// that looks like a git flag must be rejected before git runs, so it cannot be
// used to smuggle options such as --output into the subprocess.
func TestLogBetweenTags_RejectsFlagInjection(t *testing.T) {
	root := t.TempDir()
	store, err := gitstore.New(root)
	if err != nil {
		t.Fatal(err)
	}
	seedRepoWithTags(t, root, "user/repo")

	sentinel := filepath.Join(t.TempDir(), "pwned")
	cases := []struct{ from, to string }{
		{from: "", to: "--output=" + sentinel},
		{from: "--output=" + sentinel, to: "v2"},
		{from: "-n1", to: "v2"},
	}
	for _, c := range cases {
		out, err := store.LogBetweenTags(context.Background(), "user/repo", c.from, c.to)
		if err == nil {
			t.Fatalf("from=%q to=%q: expected error, got output %q", c.from, c.to, out)
		}
		if !strings.Contains(err.Error(), "invalid tag revision") {
			t.Fatalf("from=%q to=%q: expected validation error, got %v", c.from, c.to, err)
		}
		if _, statErr := os.Stat(sentinel); statErr == nil {
			t.Fatalf("from=%q to=%q: git ran with injected flag and wrote %s", c.from, c.to, sentinel)
		}
	}
}
