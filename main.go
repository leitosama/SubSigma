package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Repostitory struct {
	addr      string
	lasthash  string
	rulespath string
}

func normalize_path(dirtypath string) string {
	p := filepath.Clean(dirtypath)
	return strings.TrimPrefix(string(p), "/") + "/"
}
func main() {
	// TODO: Parse config to repos
	repos := []Repostitory{
		{addr: "https://github.com/joesecurity/sigma-rules/", lasthash: "39c5f36034e12ca81a7a4d835889dfb07c0b3903", rulespath: normalize_path("rules")},
		{addr: "https://github.com/SigmaHQ/sigma", lasthash: "c5c61ac04052632889999f21f96ddbec9efa2219", rulespath: "rules"},
	}
	for _, repo := range repos {
		objrepo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: repo.addr,
		})
		if err != nil {
			fmt.Println("Error opening repository:", err)
			os.Exit(1)
		}

		ref, _ := objrepo.Head()
		hcommit, err := objrepo.CommitObject(ref.Hash())

		if err != nil {
			fmt.Println("Error getting head commit:", err)
			os.Exit(1)
		}

		scommit, err := objrepo.CommitObject(plumbing.NewHash(repo.lasthash))
		if err != nil {
			fmt.Println("Error getting scommit:", err)
			os.Exit(1)
		}
		patch, err := scommit.Patch(hcommit)
		if err != nil {
			os.Exit(1)
		}
		for _, el := range patch.FilePatches() {
			oldfile, newfile := el.Files()
			if oldfile == nil && strings.HasPrefix(newfile.Path(), repo.rulespath) {
				fmt.Println(newfile.Path())
			}
		}
		fmt.Println("new lasthash: ", hcommit.Hash)
	}

}
