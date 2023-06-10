package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Config struct {
	Repos []Repostitory `json:"repos"`
}

type Repostitory struct {
	Addr      string `json:"addr"`
	Branch    string `json:"branch"`
	Lasthash  string `json:"lasthash"`
	Rulespath string `json:"rulespath"`
}

func normalize_path(dirtypath string) string {
	if dirtypath == "." || dirtypath == "" {
		return ""
	}
	p := filepath.Clean(dirtypath)
	return strings.TrimPrefix(string(p), "/") + "/"
}

type ChangeEnum int8

const (
	New ChangeEnum = iota
	Changed
)

type FileChange struct {
	ChangeType ChangeEnum `json:"changetype"`
	Filepath   string     `json:"filepath"`
}

func compare(hcommit *object.Commit, scommit *object.Commit, rulespath string) []FileChange {
	result := []FileChange{}
	patch, err := scommit.Patch(hcommit)
	if err != nil {
		os.Exit(1)
	}
	for _, el := range patch.FilePatches() {
		oldfile, newfile := el.Files()
		if oldfile == nil && strings.HasPrefix(newfile.Path(), rulespath) {
			result = append(result, FileChange{ChangeType: New, Filepath: newfile.Path()})
		}
	}
	return result
}

// TODO: in config or to env or to cli options
const USESTATE bool = true
const STATEFILE string = "./state.json"
const CONFIGFILE string = "./config.json"

func main() {
	var configfile string
	if _, err := os.Stat(STATEFILE); err == nil && USESTATE {
		configfile = STATEFILE
	} else {
		configfile = CONFIGFILE
	}

	config, err := ioutil.ReadFile(configfile)
	if err != nil {
		fmt.Println("Error open config.json", err)
		os.Exit(1)
	}
	cfg := &Config{}
	err = json.Unmarshal([]byte(config), cfg)
	if err != nil {
		fmt.Println(err)
	}
	for i, repo := range cfg.Repos {
		cfg.Repos[i].Rulespath = normalize_path(repo.Rulespath)
	}
	for i, repo := range cfg.Repos {
		objrepo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: repo.Addr,
		})
		if err != nil {
			fmt.Println("Error opening repository:", err)
			os.Exit(1)
		}

		ref, _ := objrepo.Head()

		if repo.Lasthash == "" {
			repo.Lasthash = ref.Hash().String()
		} else {
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
			filechanges := compare(hcommit, scommit, repo.Rulespath)
			fmt.Println(filechanges)
		}
		cfg.Repos[i] = repo
		file, _ := json.MarshalIndent(cfg, "", " ")
		_ = ioutil.WriteFile(STATEFILE, file, 0644)
	}

}
