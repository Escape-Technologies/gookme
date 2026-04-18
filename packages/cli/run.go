package cli

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/LMaxence/gookme/packages/configuration"
	"github.com/LMaxence/gookme/packages/executor"
	"github.com/LMaxence/gookme/packages/filters"
	gitclient "github.com/LMaxence/gookme/packages/git-client"
	"github.com/urfave/cli/v2"
)

const (
	RunCommandName CommandName = "run"
)

type RunCommandArguments struct {
	HookType       configuration.HookType
	GitCommandArgs []string
	From           string
	To             string
	AllFiles       bool
	TargetDir      string
}

func parseRunCommandArguments(cContext *cli.Context) (*RunCommandArguments, error) {
	hookType, err := validateHookType(cContext.String("type"))
	if err != nil {
		return nil, err
	}

	args := &RunCommandArguments{
		HookType:       hookType,
		GitCommandArgs: cContext.Args().Slice(),
		From:           cContext.String("from"),
		To:             cContext.String("to"),
		AllFiles:       cContext.Bool("all-files"),
		TargetDir:      cContext.String("dir"),
	}
	return args, nil
}

func run(args RunCommandArguments) error {
	dir, err := gitclient.GetRepoPath(nil)
	if err != nil {
		logger.Errorf("Error while getting current working directory: %s", err)
		return err
	}

	logger.Debugf("Loading configurations")
	conf, err := configuration.LoadConfiguration(dir, args.HookType)
	if err != nil {
		logger.Errorf("Error while loading configuration: %s", err)
		return err
	}

	strategy := filters.SelectResolvingStrategy(dir, &filters.StrategySelectionParameters{
		HookType:  args.HookType,
		From:      args.From,
		To:        args.To,
		AllFiles:  args.AllFiles,
		TargetDir: args.TargetDir,
	})
	changedPaths, err := strategy.Resolve()
	logger.Tracef("Resolved changeset: %v", changedPaths)
	if err != nil {
		logger.Errorf("Error while getting staged files: %s", err)
		return err
	}

	conf.Hooks = filters.FilterHooksWithChangeset(changedPaths, conf.Hooks)
	conf.Hooks = filters.FilterStepsWithOnlyOn(changedPaths, conf.Hooks)

	nSteps := 0
	for _, hook := range conf.Hooks {
		nSteps += len(hook.Steps)
	}

	logger.Infof("Running %d hooks, %d steps", len(conf.Hooks), nSteps)
	executors := make([]*executor.HookExecutor, 0, len(conf.Hooks))

	customEnv := map[string]string{
		"PATH": filepath.Join(dir, "hooks", "partials") + ":" + os.Getenv("PATH"),
	}
	for _, hook := range conf.Hooks {
		exec := executor.NewHookExecutor(&hook, args.GitCommandArgs, customEnv)
		exec = exec.WithExitOnStepError()
		executors = append(executors, exec)
	}

	hooksWg := sync.WaitGroup{}
	for _, exec := range executors {
		hooksWg.Add(1)
		go func() {
			exec.Run()
			hooksWg.Done()
		}()
	}

	hooksWg.Wait()
	return nil
}

var RunCommand *cli.Command = &cli.Command{
	Name:    string(RunCommandName),
	Aliases: []string{"r"},
	Usage:   "load and run git hooks based on Git changes",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t"},
			Value:   "pre-commit",
			Usage:   "The type of Git hook to run. Default is pre-commit, but accepted values are: pre-commit, prepare-commit-msg, commit-msg,  post-commit, post-merge, post-rewrite, pre-rebase, post-checkout, pre-push",
		},
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "An optional commit SHA-1 hash to compare to generate the staged changes from.",
		},
		&cli.StringFlag{
			Name:    "to",
			Aliases: []string{"o"},
			Usage:   "An optional commit SHA-1 hash to compare to generate the staged changes to.",
		},
		&cli.BoolFlag{
			Name:  "all-files",
			Usage: "Include staged, unstaged, and untracked files. When --dir is set, include all files in that directory.",
		},
		&cli.StringFlag{
			Name:  "dir",
			Usage: "Limit the files used to select hooks and steps to a target directory.",
		},
	},
	Action: func(cContext *cli.Context) error {
		args, err := parseRunCommandArguments(cContext)

		if err != nil {
			return err
		}
		return run(*args)
	},
}
