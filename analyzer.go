package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type ChangeEnum int8

const (
	New ChangeEnum = iota
	Changed
)

type FileChange struct {
	ChangeType ChangeEnum `json:"changetype"`
	BaseName   string     `json:"basename"`
	RemoteUrl  string     `json:"remoteurl"`
}

func (f FileChange) String() string {
	return fmt.Sprintf("{ChangeType: %v, BaseName: %s, Url: %s}", f.ChangeType, f.BaseName, f.RemoteUrl)
}

func Compare(hCommit *object.Commit, sCommit *object.Commit, repo *Repository) ([]FileChange, error) {
	result := []FileChange{}
	patch, err := sCommit.Patch(hCommit)
	if err != nil {
		return nil, err
	}
	for _, el := range patch.FilePatches() {
		oldfile, newfile := el.Files()
		if oldfile == nil && strings.HasPrefix(newfile.Path(), repo.RulesPath) {
			path := newfile.Path()
			remoteurl := fmt.Sprintf("%s/blob/%s/%s", repo.Addr, repo.Branch, path)
			result = append(result, FileChange{ChangeType: New, BaseName: filepath.Base(path), RemoteUrl: remoteurl})
		}
	}
	return result, nil
}
