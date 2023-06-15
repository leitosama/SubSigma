package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

var (
	VerboseLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	USESTATE      bool   = true
	VERBOSE       bool   = false
	STATEFILE     string = "./state.json"
	CONFIGFILE    string = "./config.json"
	configfile    string
)

// TODO: in config or to env or to cli options

func init() {
	var err error
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	VerboseLogger = log.New(os.Stdout, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	USESTATE, err = strconv.ParseBool(getEnv("USESTATE", "true"))
	checkerr(err)
	VERBOSE, err = strconv.ParseBool(getEnv("VERBOSE", "false"))
	checkerr(err)
	STATEFILE = getEnv("STATEFILE", "./state.json")
	CONFIGFILE = getEnv("CONFIGFILE", "./config.json")
	if _, err := os.Stat(STATEFILE); err == nil && USESTATE {
		configfile = STATEFILE
	} else {
		configfile = CONFIGFILE
	}

}

func checkerr(err error) {
	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}
}

func main() {
	cfg, err := GetConfig(configfile)
	if cfg.Verbose || VERBOSE {
		VERBOSE = true
	}
	checkerr(err)
	if VERBOSE {
		VerboseLogger.Println("config loaded.", cfg)
	}
	for i, repo := range cfg.Repos {
		if VERBOSE {
			VerboseLogger.Printf("processing repos[%d] - %s", i, repo.String())
		}
		fs := memfs.New()
		objrepo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
			URL: repo.Addr,
		})
		checkerr(err)
		ref, err := objrepo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", repo.Branch)), false)
		checkerr(err)
		if repo.LastCommit != "" && repo.LastCommit != ref.Hash().String() {
			if VERBOSE {
				VerboseLogger.Println("Analyzing...")
			}
			hCommit, err := objrepo.CommitObject(ref.Hash())
			checkerr(err)
			sCommit, err := objrepo.CommitObject(plumbing.NewHash(repo.LastCommit))
			checkerr(err)
			comparelink := fmt.Sprintf("%s/compare/%s...%s", repo.Addr, sCommit.Hash.String()[0:7], hCommit.Hash.String()[0:7])
			fmt.Printf("------\n[%s %s <- %s](%s)\n", GetRepoAuthor(repo), sCommit.Hash.String()[0:7], hCommit.Hash.String()[0:7], comparelink)
			filechanges, err := Compare(hCommit, sCommit, &repo)
			for i, filechange := range filechanges {
				file, err := fs.Open(filechange.Path)
				checkerr(err)
				fileContent, err := ioutil.ReadAll(file)
				checkerr(err)
				filechanges[i] = EnrichFileChange(filechange, fileContent)
			}
			if VERBOSE {
				VerboseLogger.Println(filechanges)
			}
			checkerr(err)
			compareresult := ""
			for _, filechange := range filechanges {
				compareresult += fmt.Sprintf("[%s](%s)\n", filechange.BaseName, filechange.RemoteUrl)
			}
			if compareresult != "" {
				if VERBOSE {
					VerboseLogger.Println("Here is results:")
				}
				fmt.Print(compareresult)
			}
			fmt.Println()
		}
		if repo.LastCommit != ref.Hash().String() {
			if VERBOSE {
				VerboseLogger.Printf("[+] %s lastcommit - %s\n", repo.Addr, repo.LastCommit)
			}
			repo.LastCommit = ref.Hash().String()
			cfg.Repos[i] = repo
		}
		file, _ := json.MarshalIndent(cfg, "", " ")
		_ = ioutil.WriteFile(STATEFILE, file, 0644)
	}
}
