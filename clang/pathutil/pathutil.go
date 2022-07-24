package pathutil

import (
	"path/filepath"
)

func Canonical(baseDir string, uri string) string {
	if filepath.IsAbs(uri) {
		return filepath.Clean(uri)
	}
	return filepath.Join(baseDir, uri)
}
