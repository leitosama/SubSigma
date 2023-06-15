package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bradleyjkemp/sigma-go"
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
	RuleDesc   string     `json:"ruledesc"`
	RuleTitle  string     `json:"ruletitle"`
	Path       string     `json:"path"`
}

func (f FileChange) String() string {
	return fmt.Sprintf("%#v", f)
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
			result = append(result, FileChange{ChangeType: New, Path: path, BaseName: filepath.Base(path), RemoteUrl: remoteurl})
		}
	}
	return result, nil
}

func EnrichFileChange(filechange FileChange, d []byte) FileChange {
	rule, err := sigma.ParseRule(d)
	checkerr(err)
	filechange.RuleDesc = rule.Description
	filechange.RuleTitle = rule.Title
	return filechange
}
