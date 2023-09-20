package core

import (
	"path/filepath"
	"strings"
)

type Excludes map[string]struct{}

func (e Excludes) ParseExcludes(exc string) {
	exc, _ = e.isExt(exc)
	e[exc] = struct{}{}
}

func (e Excludes) isExt(exc string) (string, bool) {
	ext := filepath.Ext(exc)
	trim := strings.TrimRight(exc, ext)
	if trim == "*" {
		return ext, true
	}
	return exc, false
}

func (e Excludes) CheckIfExclude(excPath string) bool {
	if e.checkFilenameForExclude(filepath.Base(excPath)) {
		return true
	}
	return e.checkExtForExclude(filepath.Ext(excPath))
}

func (e Excludes) checkFilenameForExclude(name string) bool {
	_, ok := e[name]
	return ok
}

func (e Excludes) checkExtForExclude(ext string) bool {
	_, ok := e[ext]
	return ok
}
