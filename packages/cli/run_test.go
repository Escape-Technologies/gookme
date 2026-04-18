package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LMaxence/gookme/packages/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAllFilesWithTargetDirIncludesIgnoredFiles(t *testing.T) {
	repoPath := setupGitRepo(t)
	chdir(t, repoPath)

	require.NoError(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("build/\n"), 0644))

	servicePath := filepath.Join(repoPath, "build", "service")
	require.NoError(t, os.MkdirAll(filepath.Join(servicePath, "hooks"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(servicePath, "main.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(servicePath, "hooks", "pre-commit.json"), []byte(`{
  "steps": [
    {
      "name": "mark ignored service",
      "command": "touch ran.txt"
    }
  ]
}`), 0644))

	err := run(RunCommandArguments{
		HookType:  configuration.PreCommitHookType,
		AllFiles:  true,
		TargetDir: "build/service",
	})

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(servicePath, "ran.txt"))
}
