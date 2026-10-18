package main

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

type node struct {
	Number int `json:"number"`
}

type response struct {
	Repository struct {
		Discussions struct {
			Nodes []node `json:"nodes"`
		} `json:"discussions"`
		Issues struct {
			Nodes []node `json:"nodes"`
		} `json:"issues"`
		PullRequests struct {
			Nodes []node `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

const nextNumberQuery = `query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    discussions(orderBy: {field: CREATED_AT, direction: DESC}, first: 1) { nodes { number } }
    issues(orderBy: {field: CREATED_AT, direction: DESC}, first: 1) { nodes { number } }
    pullRequests(orderBy: {field: CREATED_AT, direction: DESC}, first: 1) { nodes { number } }
  }
}`

func lastNumber(nodes []node) int {
	if len(nodes) > 0 {
		return nodes[0].Number
	}
	return 0
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func getNextNumber(repo repository.Repository) (int, error) {
	client, err := api.NewGraphQLClient(api.ClientOptions{Host: repo.Host})
	if err != nil {
		return 0, err
	}

	vars := map[string]interface{}{"owner": repo.Owner, "name": repo.Name}

	resp := &response{}
	err = client.Do(nextNumberQuery, vars, resp)
	if err != nil {
		return 0, err
	}

	return max(
		lastNumber(resp.Repository.Discussions.Nodes),
		lastNumber(resp.Repository.Issues.Nodes),
		lastNumber(resp.Repository.PullRequests.Nodes),
	) + 1, nil
}

func main() {
	var repo repository.Repository
	var err error

	if len(os.Args) > 1 {
		repo, err = repository.Parse(os.Args[1])
		if err != nil {
			fatal(err)
		}
	} else {
		repo, err = repository.Current()
		if err != nil {
			fatal(err)
		}
	}
	next, err := getNextNumber(repo)
	if err != nil {
		fatal(err)
	}
	fmt.Println(next)
}
