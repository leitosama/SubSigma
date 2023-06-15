package main

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	got, _ := GetConfig("./test/data/newfile_state.json")
	repos := []Repository{{
		Addr:       "https://github.com/leitosama/SubSigma",
		Branch:     "dev",
		LastCommit: "7f81f8f99065340abcdb7ed5f8dc825113ad739c",
		RulesPath:  "test/data/rules/",
	}}
	want := Config{Verbose: false, Repos: repos}
	if got.Verbose != want.Verbose {
		t.Errorf("GetConfig Verbose:\nwant: %v\ngot: %v", want.Verbose, got.Verbose)
	}
	if got.Repos[0] != want.Repos[0] {
		t.Errorf("GetConfig Repos:\nwant: %v\ngot: %v", want, got)
	}
}
