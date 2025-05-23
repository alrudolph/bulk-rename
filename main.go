package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "bulk-rename",
	Short: "Bulk rename files in a directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		root := "."

		if len(args) >= 1 {
			root = args[0]
		}

		root = strings.TrimPrefix(root, "./")

		rootCommand(root)
	},
}

func main() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}

func rootCommand(root string) {
	// root := "/home/alex/projs/merced/src/spec"
	// root := "test copy"
	files := []string{}

	WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return err
	})

	tmpFile, err := os.CreateTemp(".", "cli-input-*.txt")

	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(strings.Join(files, "\n")); err != nil {
		panic(err)
	}
	tmpFile.Close()

	fmt.Println("Close the editor to rename files")

	if err := launchVscode(tmpFile.Name()); err != nil {
		fmt.Fprintf(os.Stderr, "Editor exited with error: %v\n", err)
		return
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	newFiles := strings.Split(string(data), "\n")

	// TODO: check result is exactly the same length
	if err = handleDiff(root, files, newFiles); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to handle diff: %v\n", err)
		return
	}
}

func launchVscode(fileName string) error {
	cmd := exec.Command("code", "--wait", fileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleDiff(root string, oldFiles, newFiles []string) error {
	for i, oldFile := range oldFiles {
		newFile := newFiles[i]
		src := filepath.Join(root, oldFile)
		dst := filepath.Join(root, newFile)

		// maybe verbose mode
		// fmt.Printf("Copying %s -> %s\n", oldFile, newFile)
		err := copyAndRemoveNested(root, src, dst)

		if err != nil {
			return err
		}
	}

	return nil
}

func copyAndRemoveNested(root, src, dst string) error {
	if src == dst {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dst, err)
	}

	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
	}

	// maybe I should do all of the copying first and then all of the deleting
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to delete original %s: %w", src, err)
	}

	return removeEmptyDirsUp(filepath.Dir(src), root)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)

	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)

	return err
}

func removeEmptyDirsUp(path, stop string) error {
	for path != stop && path != "." {
		if empty, err := isDirEmpty(path); err != nil || !empty {
			return err
		}

		os.Remove(path)

		path = filepath.Dir(path)
	}

	return nil
}

func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)

	if err != nil {
		return false, err
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF {
		return true, nil
	}

	return false, err
}
