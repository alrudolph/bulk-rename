package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fileToCopy struct {
	src  string
	dst  string
	same bool
}

func NewFileToCopy(src, dst string) *fileToCopy {
	return &fileToCopy{
		src:  src,
		dst:  dst,
		same: src == dst,
	}
}

func (f *fileToCopy) Copy() error {
	if f.same {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(f.dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", f.dst, err)
	}

	if err := copyFile(f.src, f.dst); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", f.src, f.dst, err)
	}

	return nil
}

func (f *fileToCopy) DeleteSource(root string) error {
	if f.same {
		return nil
	}

	if err := os.Remove(f.src); err != nil {
		return fmt.Errorf("failed to delete original %s: %w", f.src, err)
	}

	return removeEmptyDirsUp(filepath.Dir(f.src), root)
}

func (f *fileToCopy) DeleteDestination(root string) error {
	if f.same {
		return nil
	}

	if err := os.Remove(f.dst); err != nil {
		return fmt.Errorf("failed to delete original %s: %w", f.dst, err)
	}

	return removeEmptyDirsUp(filepath.Dir(f.dst), root)
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

func handleDiffNew(root string, oldFiles, newFiles []string) error {
	files := []fileToCopy{}
	// TODO: check result is exactly the same length

	for i, oldFile := range oldFiles {
		newFile := newFiles[i]

		if newFile == "" {
			// TODO: warning message
			continue
		}

		files = append(files, *NewFileToCopy(filepath.Join(root, oldFile), filepath.Join(root, newFile)))
	}

	i := 0
	var err error

	for ; i < len(files); i++ {
		if err = files[i].Copy(); err != nil {
			break
		}
	}

	// if failure occurred while copying files,
	// delete all previously copied files to maintain consistency
	if i < len(files) {
		for j := 0; j < i; j++ {
			// todo: collect errors?
			err := files[i].DeleteDestination(root)

			if err != nil {
				return fmt.Errorf("failed to delete %s: %v", files[i].dst, err)
			}
		}

		return fmt.Errorf("failed to copy %s to %s: %v", files[i].src, files[i].dst, err)
	}

	for _, file := range files {
		// todo: collect errors?
		err := file.DeleteSource(root)

		if err != nil {
			return fmt.Errorf("failed to delete original %s: %v", file.src, err)
		}
	}

	return nil
}
