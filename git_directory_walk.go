package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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

func WalkDir(cwd, root string, fn fs.WalkDirFunc) error {
	ignores, err := buildInitialIgnores(cwd, root)

	if err != nil {
		return err
	}

	return walk(root, root, fn, ignores)
}

func buildInitialIgnores(start, end string) (*ignores, error) {
	ignores := &ignores{}

	startParts := strings.Split(strings.Trim(start, "/"), "/")
	endParts := strings.Split(strings.Trim(end, "/"), "/")

	// Find where start ends in end
	i := 0
	for ; i < len(startParts) && i < len(endParts); i++ {
		if startParts[i] != endParts[i] {
			fmt.Println("Start is not a prefix of end")
			return ignores, fmt.Errorf("start path %s is not a prefix of end path %s", start, end)
		}
	}

	// Start from startParts and add one segment at a time
	for j := len(startParts); j <= len(endParts); j++ {
		p := "/" + path.Join(endParts[:j]...)

		rule, err := getGitIgnore(p)

		if err != nil {
			return nil, fmt.Errorf("failed to get gitignore for path %s: %w", p, err)
		}

		if rule != nil {
			ignores.Rules = append(ignores.Rules, rule)
		}
	}

	return ignores, nil
}

func getGitIgnore(path string) (*gitignore.IgnoreMatcher, error) {
	gitignorePath := filepath.Join(path, ".gitignore")

	if _, err := os.Stat(gitignorePath); err == nil {
		rule, err := gitignore.NewGitIgnore(gitignorePath)

		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", gitignorePath, err)
		}

		return &rule, nil
	}

	return nil, nil
}

func walk(root, current string, fn fs.WalkDirFunc, parentIgnores *ignores) error {
	ignores := &ignores{Rules: append([]*gitignore.IgnoreMatcher{}, parentIgnores.Rules...)}

	rule, err := getGitIgnore(current)

	if err != nil {
		return err
	}

	if rule != nil {
		ignores.Rules = append(ignores.Rules, rule)
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
