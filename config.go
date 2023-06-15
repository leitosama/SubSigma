package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Config struct {
	Repos   []Repository `json:"repos"`
	Verbose bool         `json:"verbose"`
}

func (c Config) String() string {
	return fmt.Sprintf("{Repos:%v}", c.Repos)
}

type Repository struct {
	Addr       string `json:"addr"`
	Branch     string `json:"branch"`
	LastCommit string `json:"lastcommit"`
	RulesPath  string `json:"rulespath"`
}

func (r Repository) String() string {
	return fmt.Sprintf("{Addr: %s, Branch: %s, Lasthash: %s, Rulespath: %s}", r.Addr, r.Branch, r.LastCommit, r.RulesPath)
}

func normalize_path(dirtyPath string) string {
	if dirtyPath == "." || dirtyPath == "" {
		return ""
	}
	p := filepath.Clean(dirtyPath)
	return strings.TrimPrefix(string(p), "/") + "/"
}

func GetRepoAuthor(r Repository) string {
	s := strings.Split(r.Addr, "/")
	return s[len(s)-2]
}

func GetConfig(configfile string) (*Config, error) {
	config, err := ioutil.ReadFile(configfile)
	checkerr(err)
	cfg := &Config{}
	err = json.Unmarshal([]byte(config), cfg)
	checkerr(err)
	for i, repo := range cfg.Repos {
		cfg.Repos[i].RulesPath = normalize_path(repo.RulesPath)
	}
	return cfg, err
}
