package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	Addr       string `json:"addr"`
	Branch     string `json:"branch"`
	LastCommit string `json:"lastcommit"`
	RulesPath  string `json:"rulespath"`
}

func (r Repostitory) String() string {
	return fmt.Sprintf("{Addr: %s, Branch: %s, Lasthash: %s, Rulespath: %s}", r.Addr, r.Branch, r.LastCommit, r.RulesPath)
}

func (c Config) String() string {
	return fmt.Sprintf("{Repos:%v}", c.Repos)
}

func normalize_path(dirtyPath string) string {
	if dirtyPath == "." || dirtyPath == "" {
		return ""
	}
	p := filepath.Clean(dirtyPath)
	return strings.TrimPrefix(string(p), "/") + "/"
}

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

func compare(hCommit *object.Commit, sCommit *object.Commit, repo Repostitory) []FileChange {
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

var (
	VerboseLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

// TODO: in config or to env or to cli options
const USESTATE bool = true
const VERBOSE bool = true
const STATEFILE string = "./state.json"
const CONFIGFILE string = "./config.json"

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	VerboseLogger = log.New(os.Stdout, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func checkerr(err error) {
	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}
}

func main() {
	var configfile string
	if _, err := os.Stat(STATEFILE); err == nil && USESTATE {
		configfile = STATEFILE
	} else {
		configfile = CONFIGFILE
	}
	config, err := ioutil.ReadFile(configfile)
	checkerr(err)
	cfg := &Config{}
	err = json.Unmarshal([]byte(config), cfg)
	checkerr(err)
	for i, repo := range cfg.Repos {
		cfg.Repos[i].RulesPath = normalize_path(repo.RulesPath)
	}
	if VERBOSE {
		VerboseLogger.Println("config loaded.", cfg)
	}
	for i, repo := range cfg.Repos {
		if VERBOSE {
			VerboseLogger.Printf("processing repos[%d] - %s", i, repo.String())
		}
		objrepo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: repo.Addr,
		})
		checkerr(err)

		ref, err := objrepo.Head()
		checkerr(err)
		if repo.LastCommit != "" {
			hcommit, err := objrepo.CommitObject(ref.Hash())
			checkerr(err)
			sCommit, err := objrepo.CommitObject(plumbing.NewHash(repo.LastCommit))
			checkerr(err)
			filechanges := compare(hcommit, sCommit, repo)
			for _, filechange := range filechanges {
				fmt.Println("[+]", filechange.String())
			}
		}
		repo.LastCommit = ref.Hash().String()
		cfg.Repos[i] = repo
		file, _ := json.MarshalIndent(cfg, "", " ")
		_ = ioutil.WriteFile(STATEFILE, file, 0644)
	}
}
