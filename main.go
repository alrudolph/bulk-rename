package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

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
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCommand(root string) {
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

	if err := launchEditor(tmpFile.Name()); err != nil {
		fmt.Fprintf(os.Stderr, "Editor exited with error: %v\n", err)
		return
	}

	data, err := os.ReadFile(tmpFile.Name())

	if err != nil {
		panic(err)
	}

	newFiles := strings.Split(string(data), "\n")

	if err = handleDiffNew(root, files, newFiles); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to handle diff: %v\n", err)
		return
	}
}

func launchEditor(fileName string) error {
	editors := []string{}

	if editorEnv := os.Getenv("EDITOR"); editorEnv != "" {
		editors = append(editors, editorEnv)
	}

	editors = append(editors, "code", "nano", "vim", "vi")
	for _, editor := range editors {
		cmd := exec.Command(editor, fileName)

		if editor == "code" {
			cmd = exec.Command("code", "--wait", fileName)
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if editor == "code" {
			fmt.Println("CTRL-C to cancel")

			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			go func() {
				<-sigChan
				if cmd.Process != nil {
					syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
				}
			}()
		}

		err := cmd.Run()

		if err == nil {
			fmt.Println("Editor exited normally")
			return nil
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			status, ok := exitErr.Sys().(syscall.WaitStatus)
			if ok && status.Signaled() {
				return fmt.Errorf("editor was killed by signal: %s", status.Signal())
			}
		}
	}

	return fmt.Errorf("no suitable editor found. please set $EDITOR or install nano/vim/code")
}
