package gitclient

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/LMaxence/gookme/packages/logging"
)

var logger = logging.NewLogger("git-client")

type GitRefDelimiter struct {
	From string
	To   string
}

func outputToAbsolutePaths(root string, out string) []string {
	paths := make([]string, 0)

	for _, rawPath := range strings.Split(out, "\n") {
		rawPath = strings.TrimSpace(rawPath)
		if rawPath == "" {
			continue
		}

		path := filepath.FromSlash(rawPath)
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}

		paths = append(paths, filepath.Clean(path))
	}

	return paths
}

func dedupePaths(paths ...[]string) []string {
	seen := make(map[string]bool)
	deduped := make([]string, 0)

	for _, group := range paths {
		for _, path := range group {
			if seen[path] {
				continue
			}

			seen[path] = true
			deduped = append(deduped, path)
		}
	}

	return deduped
}

func GetChangedFilesBetweenRefs(
	dirPath *string,
	from string,
	to string,
) ([]string, error) {
	root, err := GetRepoPath(dirPath)
	if err != nil {
		return nil, err
	}

	out, err := execCommandAtPath(
		dirPath,
		"git",
		"diff",
		"--name-only",
		"--diff-filter=d",
		fmt.Sprintf("--line-prefix=%s", root+"/"),
		fmt.Sprintf("%s...%s", from, to),
	)

	if err != nil {
		return nil, err
	}

	return outputToAbsolutePaths(root, out), nil
}

func getCommitsToBePushed(dirPath *string) ([]string, error) {
	out, err := execCommandAtPath(
		dirPath,
		"git",
		"rev-list",
		"@{push}^..",
	)

	if err != nil {
		return nil, err
	}

	return strings.Split(string(out), "\n"), nil
}

func GetFilesToBePushed(dirPath *string) ([]string, error) {
	commits, err := getCommitsToBePushed(dirPath)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		logger.Warn("Commits to push list is empty")
		return []string{}, nil
	}

	start, end := commits[0], commits[len(commits)-1]

	return GetChangedFilesBetweenRefs(dirPath, start, end)
}

func GetStagedFiles(dirPath *string) ([]string, error) {
	root, err := GetRepoPath(dirPath)
	if err != nil {
		return nil, err
	}

	out, err := execCommandAtPath(
		dirPath,
		"git",
		"diff",
		"--cached",
		"--name-only",
		"--diff-filter=d",
		fmt.Sprintf("--line-prefix=%s", root+"/"),
	)

	if err != nil {
		return nil, err
	}

	return outputToAbsolutePaths(root, out), nil
}

func GetNotStagedFiles(dirPath *string) ([]string, error) {
	root, err := GetRepoPath(dirPath)
	if err != nil {
		return nil, err
	}

	out, err := execCommandAtPath(
		dirPath,
		"git",
		"diff",
		"--name-only",
		"--diff-filter=d",
		fmt.Sprintf("--line-prefix=%s", root+"/"),
	)

	if err != nil {
		return nil, err
	}

	return outputToAbsolutePaths(root, out), nil
}

func GetFilesChangedNCommitsBefore(dirPath *string, n int) ([]string, error) {
	out, err := execCommandAtPath(
		dirPath,
		"git",
		"diff-tree",
		"--no-commit-id",
		"--name-only", "-r",
		fmt.Sprintf("HEAD~%d", n),
	)

	if err != nil {
		return nil, err
	}

	return strings.Split(string(out), "\n"), nil
}

func GetUntrackedFiles(dirPath *string) ([]string, error) {
	root, err := GetRepoPath(dirPath)
	if err != nil {
		return nil, err
	}

	out, err := execCommandAtPath(
		dirPath,
		"git",
		"ls-files",
		"--others",
		"--exclude-standard",
	)

	if err != nil {
		return nil, err
	}

	return outputToAbsolutePaths(root, out), nil
}

func GetStagedAndUnstagedFiles(dirPath *string) ([]string, error) {
	staged, err := GetStagedFiles(dirPath)
	if err != nil {
		return nil, err
	}

	unstaged, err := GetNotStagedFiles(dirPath)
	if err != nil {
		return nil, err
	}

	untracked, err := GetUntrackedFiles(dirPath)
	if err != nil {
		return nil, err
	}

	return dedupePaths(staged, unstaged, untracked), nil
}

func GetFilesInDirectory(dirPath *string, targetDir string) ([]string, error) {
	root, err := GetRepoPath(dirPath)
	if err != nil {
		return nil, err
	}

	targetPath := filepath.FromSlash(targetDir)
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}
	targetPath = filepath.Clean(targetPath)

	paths := make([]string, 0)
	err = filepath.WalkDir(targetPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			paths = append(paths, filepath.Clean(path))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}
