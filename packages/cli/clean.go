package cli

import (
	"path/filepath"
	"strings"

	"github.com/LMaxence/gookme/packages/configuration"
	gitclient "github.com/LMaxence/gookme/packages/git-client"
	hooksscripts "github.com/LMaxence/gookme/packages/hooks-scripts"
	"github.com/urfave/cli/v2"
)

const (
	CleanCommandName CommandName = "clean"
)

func cleanHookScriptFile(
	hookType string,
) error {
	hookPath, err := gitclient.GetGitPath(filepath.Join("hooks", hookType), nil)
	if err != nil {
		return err
	}

	logger.Infof("Cleaning %s hook script", hookType)

	var content string
	logger.Debugf("Checking if script %s file exists at path %s", string(hookType), hookPath)
	exists, err := hooksscripts.ScriptFileExists(hookPath)
	if err != nil {
		return err
	}

	if exists {
		logger.Debugf("Script file exists for %s hook, loading content", hookType)
		content, err = hooksscripts.LoadScriptFileContent(hookPath)
		logger.Tracef("Loaded content of %s hook script:", hookType)
		logger.Trace(content)

		if err != nil {
			return err
		}

		// Remove the existing gookme script from the content if it exists
		content = hooksscripts.RemoveExistingGookmeScript(content)
		logger.Tracef("Content of %s hook script after deletion:", hookType)
		logger.Trace(content)
	} else {
		logger.Infof("Script file %s does not exist", hookType)
		return nil
	}

	logger.Debugf("Writing script to %s hook script file", hookType)

	if strings.Trim(content, " \n") == "#!/bin/sh" {
		logger.Infof("Script file %s is empty, removing it", hookType)
		err = hooksscripts.DeleteScriptFile(hookPath)
		if err != nil {
			return err
		}
	} else {
		err = hooksscripts.WriteScriptFileContent(hookPath, content)

		if err != nil {
			return err
		}
	}

	logger.Infof("Successfully cleaned %s hook script", hookType)
	return nil
}

func clean() error {
	for _, hookType := range configuration.ALL_HOOKS {
		err := cleanHookScriptFile(string(hookType))
		if err != nil {
			return err
		}
		logger.Infof("Successfully cleaned %s hook", hookType)
	}

	return nil
}

var CleanCommand *cli.Command = &cli.Command{
	Name:    string(CleanCommandName),
	Aliases: []string{"c"},
	Usage:   "Clean Git hooks scripts configured by Gookme",
	Action: func(cContext *cli.Context) error {
		return clean()
	},
}
