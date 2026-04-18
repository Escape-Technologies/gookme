package hooksscripts

/* ====================================================================

This file contains utility functions to read and write hook scripts.

===================================================================== */

import (
	"os"
	"path/filepath"
)

// LoadScriptFileContent loads the content of a hook script.
func LoadScriptFileContent(hookPath string) (string, error) {
	content, err := os.ReadFile(hookPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteScriptFileContent writes a hook script with the provided content.
func WriteScriptFileContent(hookPath string, content string) error {
	err := os.MkdirAll(filepath.Dir(hookPath), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(hookPath, []byte(content), 0755)
	if err != nil {
		return err
	}

	return nil
}

func DeleteScriptFile(hookPath string) error {
	err := os.Remove(hookPath)
	if err != nil {
		return err
	}

	return nil
}
