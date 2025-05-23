package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monochromegane/go-gitignore"
)

type ignores struct {
	Rules []*gitignore.IgnoreMatcher
}

func (ig *ignores) Match(path string, isDir bool) bool {
	for _, rule := range ig.Rules {
		if (*rule).Match(path, isDir) {
			return true
		}
	}

	return false
}

func WalkDir(root string, fn fs.WalkDirFunc) error {
	return walk(root, root, fn, &ignores{})
}

func walk(root, current string, fn fs.WalkDirFunc, parentIgnores *ignores) error {
	ignores := &ignores{Rules: append([]*gitignore.IgnoreMatcher{}, parentIgnores.Rules...)}

	gitignorePath := filepath.Join(current, ".gitignore")

	if _, err := os.Stat(gitignorePath); err == nil {
		rule, err := gitignore.NewGitIgnore(gitignorePath)

		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", gitignorePath, err)
		}

		ignores.Rules = append(ignores.Rules, &rule)
	}

	entries, err := os.ReadDir(current)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		if name == ".git" {
			continue
		}

		absPath := filepath.Join(current, name)
		relPath := strings.TrimPrefix(absPath, root+string(os.PathSeparator))

		if ignores.Match(absPath, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			if err := walk(root, absPath, fn, ignores); err != nil {
				return err
			}
		} else {
			// TODO: what should err be?
			if err := fn(relPath, entry, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
