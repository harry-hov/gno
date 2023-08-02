package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrGnoModNotFound is returned by [FindRootDir] when, even after traversing
// up to the root directory, a gno.mod file could not be found.
var ErrGnoModNotFound = errors.New("gno.mod file not found in current or any parent directory")

// FindRootDir determines the root directory of the project which contains the
// gno.mod file. If no gno.mod file is found, [ErrGnoModNotFound] is returned.
// The given path must be absolute.
func FindRootDir(absPath string) (string, error) {
	if !filepath.IsAbs(absPath) {
		return "", errors.New("requires absolute path")
	}

	root := filepath.VolumeName(absPath) + string(filepath.Separator)
	for absPath != root {
		modPath := filepath.Join(absPath, "gno.mod")
		_, err := os.Stat(modPath)
		if errors.Is(err, os.ErrNotExist) {
			absPath = filepath.Dir(absPath)
			continue
		}
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	return "", ErrGnoModNotFound
}

func extractPackageDir(dir string) string {
	parts := strings.Split(dir, string(filepath.Separator))
	var pkgDirParts []string

	for _, part := range parts {
		if part == "filetests" {
			break
		}
		pkgDirParts = append(pkgDirParts, part)
	}

	return filepath.Join(pkgDirParts...)
}

func isFiletestsDir(dir string) bool {
	absDirPath, err := filepath.Abs(dir)
	if err != nil {
		// TODO: don't use panic
		panic(fmt.Errorf("isFiletestsDir(): %w", err))
	}

	parts := strings.Split(absDirPath, string(filepath.Separator))
	for _, part := range parts {
		if part == "filetests" {
			return true
		}
	}
	return false
}
