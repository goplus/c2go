package pathutil

import (
	"path/filepath"
)

func Canonical(baseDir string, uri string) string {
	if filepath.IsAbs(uri) {
		return uri
	}
	return filepath.Join(baseDir, uri)
}
