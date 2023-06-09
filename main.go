package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Config struct {
	Repos []Repostitory `json:"repos"`
}

type Repostitory struct {
	Addr      string `json:"addr"`
	Lasthash  string `json:"lasthash"`
	Rulespath string `json:"rulespath"`
}

func normalize_path(dirtypath string) string {
	p := filepath.Clean(dirtypath)
	return strings.TrimPrefix(string(p), "/") + "/"
}
func main() {
	config := `{
		"repos": [{
				"addr": "https://github.com/joesecurity/sigma-rules/",
				"lasthash": "39c5f36034e12ca81a7a4d835889dfb07c0b3903",
				"rulespath": "rules"
			},
			{
				"addr": "https://github.com/SigmaHQ/sigma",
				"lasthash": "c5c61ac04052632889999f21f96ddbec9efa2219",
				"rulespath": "rules"
			}
		]
	}
	`
	cfg := Config{}
	err := json.Unmarshal([]byte(config), &cfg)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfg)
	// TODO: Parse config to repos
	for _, repo := range cfg.Repos {
		objrepo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: repo.Addr,
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

		scommit, err := objrepo.CommitObject(plumbing.NewHash(repo.Lasthash))
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
			if oldfile == nil && strings.HasPrefix(newfile.Path(), repo.Rulespath) {
				fmt.Println(newfile.Path())
			}
		}
		fmt.Println("new lasthash: ", hcommit.Hash)
	}

}
