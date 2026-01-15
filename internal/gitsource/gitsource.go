//-------------------------------------------------------------------------
//
// pgEdge Docloader
//
// Portions copyright (c) 2025 - 2026, pgEdge, Inc.
// This software is released under The PostgreSQL License
//
//-------------------------------------------------------------------------

package gitsource

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pgedge/pgedge-docloader/internal/types"
)

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
		if err := os.MkdirAll(cloneDir, 0755); err != nil {
			return fmt.Errorf("failed to create clone directory: %w", err)
		}
	}

	// Extract repo name from URL for subdirectory
	repoName := extractRepoName(gs.config.GitURL)
	gs.repoPath = filepath.Join(cloneDir, repoName)

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

	args = append(args, gs.config.GitURL, gs.repoPath)

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

	fmt.Printf("Checking out: %s\n", ref)

	cmd := exec.Command("git", "-C", gs.repoPath, "checkout", ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Pull latest if on a branch (not a tag) and not skipping fetch
	if gs.config.GitBranch != "" && !gs.config.GitSkipFetch {
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
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "repo"
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
