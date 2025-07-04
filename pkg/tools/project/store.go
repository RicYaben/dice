package project

import (
	"os"
	"path/filepath"
	"strings"
)

type FileStore interface {
	Save(path string, content []byte) error
}

type LocalStore struct {
	root string
}

func newLocalStore(root string) *LocalStore {
	return &LocalStore{root: root}
}

func (l *LocalStore) Save(fpath string, content []byte) error {
	if !strings.HasPrefix(fpath, l.root) {
		fpath = filepath.Join(l.root, fpath)
	}

	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fpath, content, 0644)
}
