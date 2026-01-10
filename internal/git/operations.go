package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tapiaw38/tracehub-server/internal/config"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// GitOps handles Git operations
type GitOps struct {
	repoPath string
	repo     *git.Repository
	config   *config.GitConfig
}

// NewGitOps creates a new GitOps instance
func NewGitOps(repoPath string, cfg *config.GitConfig) (*GitOps, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository at %s: %w", repoPath, err)
	}

	return &GitOps{
		repoPath: repoPath,
		repo:     repo,
		config:   cfg,
	}, nil
}

// CreateFixBranch creates a new branch for the fix
func (g *GitOps) CreateFixBranch(errorID string) (string, error) {
	// Generate branch name
	branchName := fmt.Sprintf("%s%s-%d", g.config.BranchPrefix, errorID, time.Now().Unix())

	// Get the current HEAD
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	err = g.repo.Storer.SetReference(ref)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	// Checkout the new branch
	w, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: refName,
		Create: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to checkout branch: %w", err)
	}

	log.Printf("Created and checked out branch: %s", branchName)
	return branchName, nil
}

// ApplyFix applies a fix to the repository
func (g *GitOps) ApplyFix(fix *models.Fix) error {
	if fix.FilePath == "" {
		return fmt.Errorf("fix file path is empty")
	}

	// Construct full file path
	fullPath := filepath.Join(g.repoPath, fix.FilePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Write the fix to the file
	err := ioutil.WriteFile(fullPath, []byte(fix.Code), 0644)
	if err != nil {
		return fmt.Errorf("failed to write fix to file: %w", err)
	}

	log.Printf("Applied fix to file: %s", fix.FilePath)
	return nil
}

// Commit creates a commit with the fix
func (g *GitOps) Commit(proposal *models.FixProposal) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Stage the changed file
	_, err = w.Add(proposal.Fix.FilePath)
	if err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	// Create commit message
	commitMsg := fmt.Sprintf("%s Fix %s in %s\n\n%s\n\nRoot cause: %s",
		g.config.CommitPrefix,
		proposal.ErrorID,
		proposal.ServiceName,
		proposal.Fix.Changes,
		proposal.RootCause,
	)

	// Get author info (use default for now)
	author := &object.Signature{
		Name:  "TraceHub AutoFix",
		Email: "autofix@tracehub.dev",
		When:  time.Now(),
	}

	// Create commit
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: author,
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Printf("Created commit for fix %s", proposal.ID)
	return nil
}

// Push pushes the branch to remote
func (g *GitOps) Push(branchName string) error {
	// Get remote
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Push to remote
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName))
	err = remote.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{refSpec},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	log.Printf("Pushed branch %s to remote", branchName)
	return nil
}

// GetCurrentBranch returns the current branch name
func (g *GitOps) GetCurrentBranch() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", err
	}

	return head.Name().Short(), nil
}

// CheckoutBranch checks out a specific branch
func (g *GitOps) CheckoutBranch(branchName string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	refName := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: refName,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}

	return nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *GitOps) HasUncommittedChanges() (bool, error) {
	w, err := g.repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := w.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}
