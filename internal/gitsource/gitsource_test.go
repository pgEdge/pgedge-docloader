//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package gitsource

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL with .git",
			url:      "https://github.com/org/myrepo.git",
			expected: "myrepo",
		},
		{
			name:     "HTTPS URL without .git",
			url:      "https://github.com/org/myrepo",
			expected: "myrepo",
		},
		{
			name:     "SSH URL",
			url:      "git@github.com:org/myrepo.git",
			expected: "myrepo",
		},
		{
			name:     "SSH URL without .git",
			url:      "git@github.com:org/myrepo",
			expected: "myrepo",
		},
		{
			name:     "Git protocol URL",
			url:      "git://github.com/org/repo.git",
			expected: "repo",
		},
		{
			name:     "Trailing slash falls back rather than cloning over the clone dir",
			url:      "https://github.com/org/",
			expected: "repo",
		},
		{
			name:     "Dot-dot falls back rather than cloning over the parent",
			url:      "https://github.com/org/..",
			expected: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoName(tt.url)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"HTTPS with .git", "https://github.com/org/repo.git", true},
		{"SSH URL", "git@github.com:org/repo.git", true},
		{"Git protocol", "git://github.com/org/repo.git", true},
		{"SSH protocol", "ssh://git@github.com/org/repo.git", true},
		{"Local path", "/path/to/files", false},
		{"Relative path", "./docs", false},
		{"Glob pattern", "/path/**/*.md", false},
		{"Windows path", "C:\\docs\\files", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitURL(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Branch and tag names are rejected unless they are plainly refs, so nothing
// a user supplies can reach git as an option.
func TestValidateRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{"Simple branch", "main", false},
		{"Namespaced branch", "release/18-stable", false},
		{"Version tag", "v1.2.3", false},
		{"Underscores and dots", "feature_x.2", false},
		{"Empty", "", true},
		{"Leading dash reads as an option", "-oProxyCommand=id", true},
		{"Long option", "--upload-pack=touch /tmp/pwned", true},
		{"Shell metacharacters", "main; rm -rf /", true},
		{"Command substitution", "$(id)", true},
		{"Leading dot", ".hidden", true},
		{"Double dot range", "main..other", true},
		{"Trailing slash", "main/", true},
		{"Lock suffix", "main.lock", true},
		{"Reflog syntax", "main@{1}", true},
		{"Whitespace", "my branch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRef("branch", tt.ref)
			if tt.wantErr && err == nil {
				t.Errorf("expected %q to be rejected, got nil error", tt.ref)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected %q to be accepted, got: %v", tt.ref, err)
			}
		})
	}
}

// A repository URL that git would read as an option is rejected.
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"HTTPS URL", "https://github.com/org/repo.git", false},
		{"SSH URL", "git@github.com:org/repo.git", false},
		{"Local path", "/srv/repos/docs.git", false},
		{"Empty", "", true},
		{"Leading dash reads as an option", "--upload-pack=id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("expected %q to be rejected, got nil error", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected %q to be accepted, got: %v", tt.url, err)
			}
		})
	}
}

// New must refuse a hostile branch name before any git process is started.
func TestNewRejectsHostileRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	cfg := &types.Config{
		GitURL:      "https://github.com/org/repo.git",
		GitBranch:   "--upload-pack=touch /tmp/pwned",
		GitCloneDir: t.TempDir(),
	}

	if _, err := New(cfg); err == nil {
		t.Fatal("expected a hostile branch name to be rejected")
	}
}

func TestGitSourceCloneAndCleanup(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a local bare repository for testing
	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-repo.git")

	// Initialize bare repo
	if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
		t.Fatalf("failed to create bare repo: %v", err)
	}

	// Create a working copy, add a file, and push
	workDir := filepath.Join(tmpDir, "work")
	if err := exec.Command("git", "clone", bareRepo, workDir).Run(); err != nil {
		t.Fatalf("failed to clone: %v", err)
	}

	testFile := filepath.Join(workDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n\nContent"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := exec.Command("git", "-C", workDir, "add", ".").Run(); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	exec.Command("git", "-C", workDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", workDir, "config", "user.name", "Test").Run()

	if err := exec.Command("git", "-C", workDir, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	if err := exec.Command("git", "-C", workDir, "push").Run(); err != nil {
		t.Fatalf("failed to push: %v", err)
	}

	// Test GitSource with temp directory (cleanup enabled)
	cfg := &types.Config{
		GitURL:       bareRepo,
		GitKeepClone: false,
	}

	gs, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create GitSource: %v", err)
	}

	sourcePaths := gs.GetSourcePaths()
	if len(sourcePaths) == 0 {
		t.Fatal("expected at least one source path")
	}
	sourcePath := sourcePaths[0]
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Error("source path should exist after clone")
	}

	// Verify test.md exists
	clonedFile := filepath.Join(sourcePath, "test.md")
	if _, err := os.Stat(clonedFile); os.IsNotExist(err) {
		t.Error("test.md should exist in cloned repo")
	}

	// Cleanup
	if err := gs.Cleanup(); err != nil {
		t.Errorf("cleanup failed: %v", err)
	}

	// Verify cleanup occurred
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Error("source path should be removed after cleanup")
	}
}

func TestGitSourceWithDocPath(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-repo.git")
	exec.Command("git", "init", "--bare", bareRepo).Run()

	workDir := filepath.Join(tmpDir, "work")
	exec.Command("git", "clone", bareRepo, workDir).Run()

	// Create docs subdirectory
	docsDir := filepath.Join(workDir, "docs", "api")
	os.MkdirAll(docsDir, 0755)
	os.WriteFile(filepath.Join(docsDir, "api.md"), []byte("# API\n\nDocs"), 0644)

	exec.Command("git", "-C", workDir, "add", ".").Run()
	exec.Command("git", "-C", workDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", workDir, "config", "user.name", "Test").Run()
	exec.Command("git", "-C", workDir, "commit", "-m", "add docs").Run()
	exec.Command("git", "-C", workDir, "push").Run()

	cfg := &types.Config{
		GitURL:       bareRepo,
		GitDocPath:   []string{"docs/api"},
		GitKeepClone: false,
	}

	gs, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create GitSource: %v", err)
	}
	defer gs.Cleanup()

	sourcePaths := gs.GetSourcePaths()
	if len(sourcePaths) == 0 {
		t.Fatal("expected at least one source path")
	}
	sourcePath := sourcePaths[0]
	if !filepath.IsAbs(sourcePath) {
		t.Error("source path should be absolute")
	}

	expectedSuffix := filepath.Join("docs", "api")
	if !strings.HasSuffix(sourcePath, expectedSuffix) {
		t.Errorf("source path should end with %s, got %s", expectedSuffix, sourcePath)
	}

	// Verify api.md exists in the doc path
	apiFile := filepath.Join(sourcePath, "api.md")
	if _, err := os.Stat(apiFile); os.IsNotExist(err) {
		t.Error("api.md should exist in doc path")
	}
}

func TestGitSourceKeepClone(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-repo.git")
	exec.Command("git", "init", "--bare", bareRepo).Run()

	workDir := filepath.Join(tmpDir, "work")
	exec.Command("git", "clone", bareRepo, workDir).Run()
	os.WriteFile(filepath.Join(workDir, "test.md"), []byte("# Test"), 0644)
	exec.Command("git", "-C", workDir, "add", ".").Run()
	exec.Command("git", "-C", workDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", workDir, "config", "user.name", "Test").Run()
	exec.Command("git", "-C", workDir, "commit", "-m", "initial").Run()
	exec.Command("git", "-C", workDir, "push").Run()

	cloneDir := filepath.Join(tmpDir, "clones")
	cfg := &types.Config{
		GitURL:       bareRepo,
		GitCloneDir:  cloneDir,
		GitKeepClone: true,
	}

	gs, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create GitSource: %v", err)
	}

	sourcePaths := gs.GetSourcePaths()
	if len(sourcePaths) == 0 {
		t.Fatal("expected at least one source path")
	}
	sourcePath := sourcePaths[0]

	// Cleanup should not remove the repo when GitKeepClone is true
	gs.Cleanup()

	// Verify repo still exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Error("source path should still exist when GitKeepClone is true")
	}
}

func TestGitSourceSkipFetch(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	bareRepo := filepath.Join(tmpDir, "test-repo.git")
	exec.Command("git", "init", "--bare", bareRepo).Run()

	workDir := filepath.Join(tmpDir, "work")
	exec.Command("git", "clone", bareRepo, workDir).Run()
	os.WriteFile(filepath.Join(workDir, "test.md"), []byte("# Test"), 0644)
	exec.Command("git", "-C", workDir, "add", ".").Run()
	exec.Command("git", "-C", workDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", workDir, "config", "user.name", "Test").Run()
	exec.Command("git", "-C", workDir, "commit", "-m", "initial").Run()
	exec.Command("git", "-C", workDir, "push").Run()

	cloneDir := filepath.Join(tmpDir, "clones")

	// First clone
	cfg := &types.Config{
		GitURL:       bareRepo,
		GitCloneDir:  cloneDir,
		GitKeepClone: true,
	}
	gs1, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create first GitSource: %v", err)
	}
	gs1.Cleanup()

	// Second use with skip-fetch should reuse existing clone
	cfg.GitSkipFetch = true
	gs2, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create second GitSource: %v", err)
	}
	defer gs2.Cleanup()

	// Should succeed without network call
	sourcePaths := gs2.GetSourcePaths()
	if len(sourcePaths) == 0 {
		t.Fatal("expected at least one source path")
	}
	if _, err := os.Stat(sourcePaths[0]); os.IsNotExist(err) {
		t.Error("source path should exist")
	}
}
