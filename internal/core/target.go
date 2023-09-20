package core

import (
	"io/fs"
	"os"
	"path/filepath"
)

type Target struct {
	Path    string `json:"path"`
	Exclude string `json:"exclude"`
}

type TargetBuilder struct {
	path               string
	excludes           Excludes
	skipDirAndExtCheck bool
}

func NewTargetBuilder(t Target) *TargetBuilder {
	var skipDirAndExtCheck bool

	e := make(Excludes)
	e.ParseExcludes(t.Exclude)

	basePath := filepath.Base(t.Path)
	ext := filepath.Ext(basePath)
	if ext == "" {
		skipDirAndExtCheck = true
	}

	return &TargetBuilder{
		path:               t.Path,
		excludes:           e,
		skipDirAndExtCheck: skipDirAndExtCheck,
	}
}

func (t *TargetBuilder) Build() ([]string, error) {
	if t.skipDirAndExtCheck {
		return t.parseAllFiles()
	}

	return t.parseFilesWithConstraints()
}

func (t *TargetBuilder) parseAllFiles() ([]string, error) {
	var suitable []string
	dir := filepath.Dir(t.path)

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if isExc := t.excludes.CheckIfExclude(path); isExc {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		suitable = append(suitable, abs)

		return nil
	})

	if err != nil {
		return suitable, err
	}

	return suitable, nil
}

func (t *TargetBuilder) parseFilesWithConstraints() ([]string, error) {
	var suitable []string
	dirPath := filepath.Dir(t.path)

	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return suitable, err
	}

	abs, err := filepath.Abs(dirPath)
	if err != nil {
		return suitable, err
	}
	suitable = append(suitable, abs)

	targetExt := filepath.Ext(t.path)
	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		if filepath.Ext(f.Name()) == targetExt {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(t.path), f.Name()))
			if err != nil {
				return suitable, err
			}
			suitable = append(suitable, abs)
		}
	}

	return suitable, nil
}
