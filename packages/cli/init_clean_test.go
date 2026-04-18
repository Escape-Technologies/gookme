package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LMaxence/gookme/packages/configuration"
	helpers "github.com/LMaxence/gookme/packages/test-helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chdir(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	return strings.TrimSpace(string(out))
}

func setupGitRepo(t *testing.T) string {
	t.Helper()

	repoPath, err := helpers.SetupTmpGit()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(repoPath))
	})

	return repoPath
}

func setupCommittedGitRepo(t *testing.T) string {
	t.Helper()

	repoPath := setupGitRepo(t)

	require.NoError(t, helpers.WriteFile(repoPath, "README.md", "test"))
	runGit(t, repoPath, "add", ".")
	runGit(
		t,
		repoPath,
		"-c",
		"user.email=gookme@example.com",
		"-c",
		"user.name=Gookme",
		"-c",
		"commit.gpgsign=false",
		"commit",
		"-m",
		"initial",
	)

	return repoPath
}

func setupLinkedWorktree(t *testing.T) string {
	t.Helper()

	repoPath := setupCommittedGitRepo(t)
	worktreePath := filepath.Join(filepath.Dir(repoPath), filepath.Base(repoPath)+"-worktree")
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(worktreePath))
	})

	runGit(t, repoPath, "worktree", "add", "-b", "gookme-test-worktree", worktreePath)

	return worktreePath
}

func resolvedHookPath(t *testing.T, repoPath string, hookType configuration.HookType) string {
	t.Helper()

	hookPath := filepath.FromSlash(runGit(t, repoPath, "rev-parse", "--git-path", filepath.Join("hooks", string(hookType))))
	if filepath.IsAbs(hookPath) {
		return filepath.Clean(hookPath)
	}

	return filepath.Clean(filepath.Join(repoPath, hookPath))
}

func TestInitWritesHookToResolvedGitPath(t *testing.T) {
	repoPath := setupGitRepo(t)
	hookPath := resolvedHookPath(t, repoPath, configuration.PreCommitHookType)
	chdir(t, repoPath)

	err := initFn(InitCommandArguments{
		HookTypes: []configuration.HookType{configuration.PreCommitHookType},
	})

	require.NoError(t, err)
	assert.FileExists(t, hookPath)

	content, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "gookme run -t pre-commit $1")
}

func TestInitFromLinkedWorktreeWritesHookToResolvedGitPath(t *testing.T) {
	worktreePath := setupLinkedWorktree(t)
	hookPath := resolvedHookPath(t, worktreePath, configuration.PreCommitHookType)
	chdir(t, worktreePath)

	err := initFn(InitCommandArguments{
		HookTypes: []configuration.HookType{configuration.PreCommitHookType},
	})

	require.NoError(t, err)
	assert.FileExists(t, hookPath)
	assert.NoDirExists(t, filepath.Join(worktreePath, ".git", "hooks"))
}

func TestCleanFromLinkedWorktreeRemovesGookmeScriptFromResolvedGitPath(t *testing.T) {
	worktreePath := setupLinkedWorktree(t)
	hookPath := resolvedHookPath(t, worktreePath, configuration.PreCommitHookType)
	chdir(t, worktreePath)

	require.NoError(t, initFn(InitCommandArguments{
		HookTypes: []configuration.HookType{configuration.PreCommitHookType},
	}))
	require.NoError(t, clean())

	_, err := os.Stat(hookPath)
	assert.True(t, os.IsNotExist(err))
}

func TestInitUsesConfiguredHooksPath(t *testing.T) {
	repoPath := setupGitRepo(t)
	runGit(t, repoPath, "config", "core.hooksPath", ".githooks")
	hookPath := filepath.Join(repoPath, ".githooks", string(configuration.PreCommitHookType))
	chdir(t, repoPath)

	err := initFn(InitCommandArguments{
		HookTypes: []configuration.HookType{configuration.PreCommitHookType},
	})

	require.NoError(t, err)
	assert.FileExists(t, hookPath)
}
