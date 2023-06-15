package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestCompare(t *testing.T) {
	repo := Repository{
		Addr:       "https://github.com/leitosama/SubSigma",
		Branch:     "dev",
		LastCommit: "7f81f8f99065340abcdb7ed5f8dc825113ad739c",
		RulesPath:  "test/data/rules/",
	}
	rulefile := "test/data/rules/example.yml"
	objrepo, _ := git.PlainOpen(".")
	beforenewfile, _ := objrepo.CommitObject(plumbing.NewHash(repo.LastCommit))
	ref, _ := objrepo.Head()
	hCommit, _ := objrepo.CommitObject(ref.Hash())
	got, _ := Compare(hCommit, beforenewfile, &repo)
	rurl := fmt.Sprintf("%s/blob/%s/%s", repo.Addr, repo.Branch, rulefile)
	want := []FileChange{{ChangeType: 0, BaseName: filepath.Base(rulefile), RemoteUrl: rurl}}
	if got[0] != want[0] {
		t.Errorf("got - %v, want - %v", got, want)
	}
}
