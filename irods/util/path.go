package util

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// JoinPath makes the path from dir and file paths
func JoinPath(dirPath string, filePath string) string {
	if strings.HasSuffix(dirPath, "/") {
		return fmt.Sprintf("%s/%s", dirPath[0:len(dirPath)-1], filePath)
	}
	return fmt.Sprintf("%s/%s", dirPath, filePath)
}

// SplitPath splits the path into dir and file
func SplitPath(p string) (string, string) {
	return filepath.Split(p)
}

// GetDirname returns the dir of the path
func GetDirname(p string) string {
	return filepath.Dir(p)
}

// GetFileName returns the filename of the path
func GetFileName(p string) string {
	return filepath.Base(p)
}

// IsAbsolutePath returns true if the path is absolute
func IsAbsolutePath(p string) bool {
	return strings.HasPrefix(p, "/")
}

// GetPathDepth returns depth of the path
// "/" returns 0
// "abc" returns -1
// "/abc" returns 0
// "/a/b" returns 1
// "/a/b/c" returns 2
func GetPathDepth(p string) int {
	if !strings.HasPrefix(p, "/") {
		return -1
	}

	cleanPath := path.Clean(p)

	if cleanPath == "/" {
		return 0
	}

	pArr := strings.Split(p[1:], "/")
	return len(pArr) - 1
}

// GetParentDirs returns all parent dirs
func GetParentDirs(p string) []string {
	parents := []string{}

	if p == "/" {
		return parents
	}

	curPath := p
	for len(curPath) > 0 && curPath != "/" {
		curDir := GetDirname(curPath)
		if len(curDir) > 0 {
			parents = append(parents, curDir)
		}

		curPath = curDir
	}

	// sort
	sort.Slice(parents, func(i int, j int) bool {
		return len(parents[i]) < len(parents[j])
	})

	return parents
}

// GetRelativePath returns relative path
func GetRelativePath(base string, target string) (string, error) {
	return filepath.Rel(base, target)
}

// GetCleanPath returns clean path
func GetCleanPath(path string) string {
	return filepath.Clean(path)
}
