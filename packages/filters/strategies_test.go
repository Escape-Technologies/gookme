package filters

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterChangesetWithTargetDir(t *testing.T) {
	repositoryPath := t.TempDir()
	buildPath := filepath.Join(repositoryPath, "build")

	files := []string{
		filepath.Join(buildPath, "service", "main.go"),
		filepath.Join(buildPath, "service", "hooks", "pre-commit.json"),
		filepath.Join(repositoryPath, "building", "main.go"),
		filepath.Join(repositoryPath, "README.md"),
	}

	filtered := filterChangesetWithTargetDir(files, repositoryPath, "build")

	assert.ElementsMatch(t, []string{
		filepath.Join(buildPath, "service", "main.go"),
		filepath.Join(buildPath, "service", "hooks", "pre-commit.json"),
	}, filtered)
}
