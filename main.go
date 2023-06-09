package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func main() {
	// Open the repository
	repo, err := git.PlainOpen("Sigma-Rules")

	if err != nil {
		fmt.Println("Error opening repository:", err)
		os.Exit(1)
	}

	// Open HEAD commit
	ref, err := repo.Head()

	hcommit, err := repo.CommitObject(ref.Hash())

	if err != nil {
		fmt.Println("Error getting commit1:", err)
		os.Exit(1)
	}

	fmt.Println(hcommit.Message)
}
