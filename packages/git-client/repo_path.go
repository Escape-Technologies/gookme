package gitclient

import (
	"os"
	"path/filepath"
	"strings"
)

// GetRepoPath returns the absolute path of the folder containing the git repository
// for the current working directory.
func GetRepoPath(start *string) (string, error) {
	// If the start path is not provided, use the current working directory

	if start == nil {
		path, err := os.Getwd()
		if err != nil {
			return "", err
		}

		start = &path
	}

	// Run the git-rev-parse command from the start path
	out, err := execCommandAtPath(start, "git", "rev-parse", "--show-toplevel")

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// GetGitPath returns the resolved filesystem path for a path inside Git's
// repository metadata.
func GetGitPath(relativePath string, start *string) (string, error) {
	repositoryPath, err := GetRepoPath(start)
	if err != nil {
		return "", err
	}

	out, err := execCommandAtPath(&repositoryPath, "git", "rev-parse", "--git-path", filepath.ToSlash(relativePath))
	if err != nil {
		return "", err
	}

	gitPath := filepath.FromSlash(strings.TrimSpace(out))
	if filepath.IsAbs(gitPath) {
		return filepath.Clean(gitPath), nil
	}

	return filepath.Clean(filepath.Join(repositoryPath, gitPath)), nil
}
