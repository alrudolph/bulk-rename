# Bulk Rename CLI

A simple CLI tool that helps you rename and reorganize files by editing a list in VSCode.

## Features

* Lists all files in a given directory (recursively)
* Lets you edit file names and paths in your editor
* Automatically:
  * Creates any new directories needed
  * Moves/renames files to the new paths
  * Deletes old files and cleans up empty directories
* Respects `.gitignore` files (including those in subdirectories)

## Installation

```bash
go install github.com/alrudolph/bulk-rename@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

## Usage

```bash
bulk-rename ./your-directory
```

This will:

1. Create a temporary file listing all the files under `./your-directory`
2. Open the file in VS Code
3. Wait for you to make changes and save
4. Process the changes by moving and renaming the files as needed

## Example

Original structure:

```
a.py
b/c.py
b/d.py
```

You edit it to:

```
new-dir/a.py
new-dir/b/c.py
new-dir/b/d.py
```

The CLI will:

* Create the `new-dir/` and subfolders
* Move the files into their new locations
* Clean up the `b/` if it becomes empty

## .gitignore Support

Files ignored by `.gitignore` files (even in subdirectories) are excluded from the list. This helps avoid unintentional renaming of build artifacts or temporary files.

## Development

Run the CLI from source:

```bash
go run main.go ./test-data
```

To build:

```bash
go build -o bulk-rename
```

## Caveats

* This tool performs file operations; it's recommended to use it with version control.
* It currently assumes you are using VS Code or that `$EDITOR` points to a compatible CLI-based editor.
