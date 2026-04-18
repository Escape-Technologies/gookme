package filters

/* ====================== Strategies ======================

Strategies denote how Gookme will select files to account for
in the changeset. This changeset is therefore used to determine
whether a hook should be executed or not.

=========================================================== */

import (
	"path/filepath"
	"strings"

	configuration "github.com/LMaxence/gookme/packages/configuration"
	gitclient "github.com/LMaxence/gookme/packages/git-client"
)

// StrategySelectionParameters is a struct that holds the parameters for selecting a strategy.
type StrategySelectionParameters struct {
	// HookType is the type of the hook that is being executed.
	HookType configuration.HookType
	// From is the Git ref from which the changeset should be calculated.
	From string
	// To is the Git ref to which the changeset should be calculated.
	To string
	// AllFiles includes staged, unstaged, and untracked files in the changeset.
	AllFiles bool
	// TargetDir limits the changeset to a specific directory.
	TargetDir string
}

// SelectResolvingStrategy selects the appropriate changeset resolving strategy base on the parameters.
//
// When the parameters `from` and `to` are set to a non-empty string, the `FromToChangesResolvingStrategy` is used
// regardless of the hook type.
//
// When the hook type is `PrePushHookType`, the `PrePushChangesResolvingStrategy` is used.
//
// When the hook type is `PostCommitHookType`, the `StagedChangesResolvingStrategy` is used.
//
// Otherwise, the `StagedChangesResolvingStrategy` is used.
func SelectResolvingStrategy(dir string, parameters *StrategySelectionParameters) ChangesetResolvingStrategy {
	var changesetResolvingStrategy ChangesetResolvingStrategy

	if parameters.AllFiles {
		logger.Debugf("Using AllFilesChangesResolvingStrategy")
		changesetResolvingStrategy = NewAllFilesChangesResolvingStrategy(dir, parameters.TargetDir)
	} else if parameters.From != "" && parameters.To != "" {
		logger.Debugf("Using FromToChangesResolvingStrategy")
		changesetResolvingStrategy = NewFromToChangesResolvingStrategy(dir, parameters.From, parameters.To)
	} else if parameters.HookType == configuration.PrePushHookType {
		logger.Debugf("Using PrePushChangesResolvingStrategy")
		changesetResolvingStrategy = NewStagedChangesResolvingStrategy(dir)
	} else if parameters.HookType == configuration.PostCommitHookType {
		logger.Debugf("Using StagedChangesResolvingStrategy")
		changesetResolvingStrategy = NewStagedChangesResolvingStrategy(dir)
	} else {
		logger.Debugf("Using StagedChangesResolvingStrategy")
		changesetResolvingStrategy = NewStagedChangesResolvingStrategy(dir)
	}

	if parameters.TargetDir != "" && !parameters.AllFiles {
		changesetResolvingStrategy = NewTargetDirChangesResolvingStrategy(dir, parameters.TargetDir, changesetResolvingStrategy)
	}

	return changesetResolvingStrategy
}

// ChangesetResolvingStrategy is an interface that defines a strategy to resolve a changeset.
type ChangesetResolvingStrategy interface {
	Resolve() ([]string, error)
}

// StagedChangesResolvingStrategy is a changeset resolving strategy that
// resolves the changeset by looking at the staged files in the repository.
type StagedChangesResolvingStrategy struct {
	repositoryPath string
}

func NewStagedChangesResolvingStrategy(repositoryPath string) *StagedChangesResolvingStrategy {
	return &StagedChangesResolvingStrategy{
		repositoryPath: repositoryPath,
	}
}

func (s *StagedChangesResolvingStrategy) Resolve() ([]string, error) {
	return gitclient.GetStagedFiles(&s.repositoryPath)
}

// AllFilesChangesResolvingStrategy resolves the changeset by looking at staged,
// unstaged, and untracked files. When a target directory is provided, it walks
// that directory on disk so ignored generated files can be included.
type AllFilesChangesResolvingStrategy struct {
	repositoryPath string
	targetDir      string
}

func NewAllFilesChangesResolvingStrategy(repositoryPath string, targetDir string) *AllFilesChangesResolvingStrategy {
	return &AllFilesChangesResolvingStrategy{
		repositoryPath: repositoryPath,
		targetDir:      targetDir,
	}
}

func (s *AllFilesChangesResolvingStrategy) Resolve() ([]string, error) {
	if s.targetDir != "" {
		return gitclient.GetFilesInDirectory(&s.repositoryPath, s.targetDir)
	}

	return gitclient.GetStagedAndUnstagedFiles(&s.repositoryPath)
}

// TargetDirChangesResolvingStrategy limits another changeset resolving strategy
// to a specific target directory.
type TargetDirChangesResolvingStrategy struct {
	repositoryPath string
	targetDir      string
	strategy       ChangesetResolvingStrategy
}

func NewTargetDirChangesResolvingStrategy(repositoryPath string, targetDir string, strategy ChangesetResolvingStrategy) *TargetDirChangesResolvingStrategy {
	return &TargetDirChangesResolvingStrategy{
		repositoryPath: repositoryPath,
		targetDir:      targetDir,
		strategy:       strategy,
	}
}

func (s *TargetDirChangesResolvingStrategy) Resolve() ([]string, error) {
	paths, err := s.strategy.Resolve()
	if err != nil {
		return nil, err
	}

	return filterChangesetWithTargetDir(paths, s.repositoryPath, s.targetDir), nil
}

// FromToChangesResolvingStrategy is a changeset resolving strategy that
// resolves the changeset by looking at the changes between two refs.
type FromToChangesResolvingStrategy struct {
	repositoryPath string
	from           string
	to             string
}

func NewFromToChangesResolvingStrategy(repositoryPath string, from string, to string) *FromToChangesResolvingStrategy {
	return &FromToChangesResolvingStrategy{
		repositoryPath: repositoryPath,
		from:           from,
		to:             to,
	}
}

func (s *FromToChangesResolvingStrategy) Resolve() ([]string, error) {
	return gitclient.GetChangedFilesBetweenRefs(&s.repositoryPath, s.from, s.to)
}

// PrePushChangesResolvingStrategy is a changeset resolving strategy that
// resolves the changeset by looking at the files about to be pushed in the repository.
type PrePushChangesResolvingStrategy struct {
	repositoryPath string
}

func NewPrePushChangesResolvingStrategy(repositoryPath string) *PrePushChangesResolvingStrategy {
	return &PrePushChangesResolvingStrategy{
		repositoryPath: repositoryPath,
	}
}

func (s *PrePushChangesResolvingStrategy) Resolve() ([]string, error) {
	return gitclient.GetFilesToBePushed(&s.repositoryPath)
}

func filterChangesetWithTargetDir(changedPaths []string, repositoryPath string, targetDir string) []string {
	targetPath := filepath.FromSlash(targetDir)
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(repositoryPath, targetPath)
	}
	targetPath = filepath.Clean(targetPath)

	filtered := make([]string, 0)
	for _, path := range changedPaths {
		path = filepath.Clean(path)

		if path == targetPath || strings.HasPrefix(path, targetPath+string(filepath.Separator)) {
			filtered = append(filtered, path)
		}
	}

	return filtered
}
