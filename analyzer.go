package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

func compare(hCommit *object.Commit, sCommit *object.Commit, repo *Repository) []FileChange {
	result := []FileChange{}
	patch, err := sCommit.Patch(hCommit)
	if err != nil {
		os.Exit(1)
	}
	for _, el := range patch.FilePatches() {
		oldfile, newfile := el.Files()
		if oldfile == nil && strings.HasPrefix(newfile.Path(), repo.RulesPath) {
			path := newfile.Path()
			remoteurl := fmt.Sprintf("%s/blob/%s/%s", repo.Addr, repo.Branch, path)
			result = append(result, FileChange{ChangeType: New, BaseName: filepath.Base(path), RemoteUrl: remoteurl})
		}
	}
	return result
}

func AnalyzeRepo(objrepo *git.Repository, repo *Repository) ([]FileChange, error) {
	var (
		err         error
		ref         *plumbing.Reference
		hCommit     *object.Commit
		sCommit     *object.Commit
		filechanges []FileChange
	)
	ref, err = objrepo.Head()
	checkerr(err)
	hCommit, err = objrepo.CommitObject(ref.Hash())
	checkerr(err)
	sCommit, err = objrepo.CommitObject(plumbing.NewHash(repo.LastCommit))
	checkerr(err)
	filechanges = compare(hCommit, sCommit, repo)
	return filechanges, err
}
