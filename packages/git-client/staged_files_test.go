package gitclient

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	helpers "github.com/LMaxence/gookme/packages/test-helpers"

	"github.com/stretchr/testify/assert"
)

func TestGetStagedFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	// Create a file
	_, err = execCommandAtPath(&tmpDir, "touch", "file1")
	assert.NoError(t, err)

	// Stage the file
	_, err = execCommandAtPath(&tmpDir, "git", "add", ".")
	assert.NoError(t, err)

	// Call the function
	files, err := GetStagedFiles(&tmpDir)

	// Assert the results
	assert.NoError(t, err)

	assert.Contains(t, files, filepath.Join(tmpDir, "file1"))
}

func TestGetStagedFilesWithNoStagedFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	// Create a file
	_, err = execCommandAtPath(&tmpDir, "touch", "file1")
	assert.NoError(t, err)

	// Call the function
	files, err := GetStagedFiles(&tmpDir)

	// Assert the results
	assert.NoError(t, err)
	assert.NotContains(t, files, filepath.Join(tmpDir, "file1"))
}

func TestGetNotStagedFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	// Create a file
	_, err = execCommandAtPath(&tmpDir, "touch", "file1")
	assert.NoError(t, err)

	// Stage the file and then modify it
	_, err = execCommandAtPath(&tmpDir, "git", "add", ".")
	assert.NoError(t, err)

	// Write "test" to the file
	file, err := os.OpenFile(tmpDir+"/file1", os.O_WRONLY, fs.ModePerm)
	assert.NoError(t, err)

	_, err = file.WriteString("test")
	assert.NoError(t, err)

	// Call the function
	files, err := GetNotStagedFiles(&tmpDir)

	// Assert the results
	assert.NoError(t, err)
	assert.Contains(t, files, filepath.Join(tmpDir, "file1"))

	assert.NoError(t, file.Close())
}

func TestGetNotStagedFilesWithNoNotStagedFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	// Create a file
	_, err = execCommandAtPath(&tmpDir, "touch", "file1")
	assert.NoError(t, err)

	// Stage the file
	_, err = execCommandAtPath(&tmpDir, "git", "add", ".")
	assert.NoError(t, err)

	// Call the function
	files, err := GetNotStagedFiles(&tmpDir)

	// Assert the results
	assert.NoError(t, err)
	assert.NotContains(t, files, filepath.Join(tmpDir, "file1"))
}

func TestGetStagedAndUnstagedFiles(t *testing.T) {
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	trackedPath := filepath.Join(tmpDir, "tracked.txt")
	err = os.WriteFile(trackedPath, []byte("staged"), 0644)
	assert.NoError(t, err)

	_, err = execCommandAtPath(&tmpDir, "git", "add", ".")
	assert.NoError(t, err)

	err = os.WriteFile(trackedPath, []byte("unstaged"), 0644)
	assert.NoError(t, err)

	untrackedPath := filepath.Join(tmpDir, "untracked.txt")
	err = os.WriteFile(untrackedPath, []byte("untracked"), 0644)
	assert.NoError(t, err)

	files, err := GetStagedAndUnstagedFiles(&tmpDir)

	assert.NoError(t, err)
	assert.Contains(t, files, trackedPath)
	assert.Contains(t, files, untrackedPath)
	assert.Len(t, files, 2)
}

func TestGetFilesInDirectoryIncludesIgnoredFiles(t *testing.T) {
	tmpDir, err := helpers.SetupTmpGit()
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("build/\n"), 0644)
	assert.NoError(t, err)

	servicePath := filepath.Join(tmpDir, "build", "service")
	err = os.MkdirAll(filepath.Join(servicePath, "nested"), 0755)
	assert.NoError(t, err)

	mainPath := filepath.Join(servicePath, "main.go")
	configPath := filepath.Join(servicePath, "nested", "config.json")
	otherPath := filepath.Join(tmpDir, "build", "other", "main.go")

	err = os.WriteFile(mainPath, []byte("package main"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(configPath, []byte("{}"), 0644)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Dir(otherPath), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(otherPath, []byte("package main"), 0644)
	assert.NoError(t, err)

	files, err := GetFilesInDirectory(&tmpDir, "build/service")

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{mainPath, configPath}, files)
}
