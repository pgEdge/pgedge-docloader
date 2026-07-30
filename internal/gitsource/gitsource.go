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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

// gitRefPattern is the conservative subset of git refname syntax accepted for
// --git-branch and --git-tag. It is deliberately narrower than
// git-check-ref-format(1): input outside it is rejected outright rather than
// escaped, so nothing a user supplies can reach git as anything but a ref.
var gitRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// validateRef rejects a branch or tag name that git could mistake for an
// option, or that uses refname syntax with side effects. kind names the
// option being validated so the error can point at it.
func validateRef(kind, ref string) error {
	if !gitRefPattern.MatchString(ref) {
		return fmt.Errorf("invalid %s %q: must start with a letter or digit and "+
			"contain only letters, digits and the characters '.', '_', '/' and '-'", kind, ref)
	}

	if strings.Contains(ref, "..") || strings.HasSuffix(ref, "/") || strings.HasSuffix(ref, ".lock") {
		return fmt.Errorf("invalid %s %q: must not contain '..' or end with '/' or '.lock'", kind, ref)
	}

	return nil
}

// validateURL rejects a repository URL that git would parse as an option
// rather than as a location to clone from.
func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("git URL must not be empty")
	}

	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("invalid git URL %q: must not start with '-'", url)
	}

	return nil
}

// GitSource represents a git repository source
type GitSource struct {
	config   *types.Config
	repoPath string
	cleanup  func() error
}

// New creates a new GitSource from configuration
func New(cfg *types.Config) (*GitSource, error) {
	// Check git is available
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git command not found: please install git to use git sources")
	}

	// Validate everything that ends up on a git command line before any
	// subprocess is started
	if err := validateURL(cfg.GitURL); err != nil {
		return nil, err
	}
	if cfg.GitBranch != "" {
		if err := validateRef("branch", cfg.GitBranch); err != nil {
			return nil, err
		}
	}
	if cfg.GitTag != "" {
		if err := validateRef("tag", cfg.GitTag); err != nil {
			return nil, err
		}
	}

	gs := &GitSource{
		config: cfg,
	}

	if err := gs.setup(); err != nil {
		return nil, err
	}

	return gs, nil
}

// setup prepares the git repository
func (gs *GitSource) setup() error {
	// Determine clone directory
	cloneDir := gs.config.GitCloneDir
	if cloneDir == "" {
		// Use temp directory
		tmpDir, err := os.MkdirTemp("", "docloader-git-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		cloneDir = tmpDir

		// Set up cleanup for temp directory
		if !gs.config.GitKeepClone {
			gs.cleanup = func() error {
				return os.RemoveAll(tmpDir)
			}
		}
	} else {
		// Using specified directory - ensure it exists
		if err := os.MkdirAll(cloneDir, 0750); err != nil {
			return fmt.Errorf("failed to create clone directory: %w", err)
		}
	}

	// Extract repo name from URL for subdirectory. Resolve to an absolute
	// path: a relative path beginning with '-' would be parsed by git as an
	// option rather than as a directory.
	repoName := extractRepoName(gs.config.GitURL)
	repoPath, err := filepath.Abs(filepath.Join(cloneDir, repoName))
	if err != nil {
		return fmt.Errorf("failed to resolve clone path: %w", err)
	}
	gs.repoPath = repoPath

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(gs.repoPath, ".git")); err == nil {
		// Repo exists
		if gs.config.GitSkipFetch {
			fmt.Printf("Using existing clone: %s\n", gs.repoPath)
		} else {
			fmt.Printf("Repository exists, fetching updates: %s\n", gs.repoPath)
			if err := gs.fetch(); err != nil {
				return err
			}
		}
	} else {
		// Clone repository
		if err := gs.clone(); err != nil {
			return err
		}
	}

	// Checkout specific branch/tag if specified
	if err := gs.checkout(); err != nil {
		return err
	}

	return nil
}

// clone clones the repository
func (gs *GitSource) clone() error {
	fmt.Printf("Cloning repository: %s\n", gs.config.GitURL)

	args := []string{"clone", "--depth", "1"}

	// Add branch/tag to clone command for efficiency
	if gs.config.GitBranch != "" {
		args = append(args, "--branch", gs.config.GitBranch)
	} else if gs.config.GitTag != "" {
		args = append(args, "--branch", gs.config.GitTag)
	}

	// "--" stops git parsing either positional argument as an option
	args = append(args, "--", gs.config.GitURL, gs.repoPath)

	// #nosec G204 -- the executable and every flag are constants; the URL is
	// checked by validateURL, the ref by validateRef, and repoPath is an
	// absolute path built in setup().
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// fetch fetches updates from the remote
func (gs *GitSource) fetch() error {
	// #nosec G204 -- the executable, subcommand and flags are constants;
	// repoPath is an absolute path built in setup().
	cmd := exec.Command("git", "-C", gs.repoPath, "fetch", "--all", "--prune", "--tags")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	return nil
}

// checkout checks out the specified branch or tag
func (gs *GitSource) checkout() error {
	var ref string
	if gs.config.GitBranch != "" {
		ref = gs.config.GitBranch
	} else if gs.config.GitTag != "" {
		ref = gs.config.GitTag
	} else {
		return nil // Use default branch from clone
	}

	// Re-check the ref immediately before it reaches a command line, so this
	// path stays safe even if a future caller builds a GitSource by hand
	if err := validateRef("ref", ref); err != nil {
		return err
	}

	fmt.Printf("Checking out: %s\n", ref)

	// #nosec G204 -- the executable and subcommand are constants; ref is
	// checked by validateRef just above and repoPath is an absolute path
	// built in setup().
	cmd := exec.Command("git", "-C", gs.repoPath, "checkout", ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Pull latest if on a branch (not a tag) and not skipping fetch
	if gs.config.GitBranch != "" && !gs.config.GitSkipFetch {
		// #nosec G204 -- the executable, subcommand and flags are constants;
		// repoPath is an absolute path built in setup().
		pullCmd := exec.Command("git", "-C", gs.repoPath, "pull", "--ff-only")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		// Pull may fail for various reasons (detached HEAD, conflicts, etc.)
		// This is not fatal - we already have the checkout
		if err := pullCmd.Run(); err != nil {
			fmt.Printf("Note: git pull skipped (%v)\n", err)
		}
	}

	return nil
}

// GetSourcePaths returns the paths to process files from
func (gs *GitSource) GetSourcePaths() []string {
	if len(gs.config.GitDocPath) > 0 {
		paths := make([]string, len(gs.config.GitDocPath))
		for i, docPath := range gs.config.GitDocPath {
			paths[i] = filepath.Join(gs.repoPath, docPath)
		}
		return paths
	}
	return []string{gs.repoPath}
}

// Cleanup removes the cloned repository if configured
func (gs *GitSource) Cleanup() error {
	if gs.cleanup != nil {
		fmt.Println("Cleaning up cloned repository...")
		return gs.cleanup()
	}
	return nil
}

// extractRepoName extracts repository name from URL
func extractRepoName(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH URLs like git@github.com:org/repo
	if strings.Contains(url, ":") && !strings.Contains(url, "://") {
		parts := strings.Split(url, ":")
		if len(parts) > 1 {
			url = parts[len(parts)-1]
		}
	}

	// Get last path component
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]

	// Fall back for anything that is not a plain subdirectory name: "" or "."
	// would clone over the clone directory itself, ".." over its parent
	if name == "" || name == "." || name == ".." {
		return "repo"
	}

	return name
}

// IsGitURL checks if a string looks like a git URL
func IsGitURL(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "git@") ||
		strings.HasPrefix(s, "git://") ||
		strings.HasPrefix(s, "ssh://") ||
		(strings.HasPrefix(s, "https://") && strings.Contains(s, ".git")) ||
		(strings.HasPrefix(s, "http://") && strings.Contains(s, ".git"))
}
