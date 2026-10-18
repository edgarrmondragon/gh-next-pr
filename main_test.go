package main

import (
	"encoding/json"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockGraphQL(t *testing.T, repo repository.Repository, responseJSON string) {
	t.Setenv("GH_TOKEN", "abc123")
	t.Setenv("GH_CONFIG_DIR", t.TempDir())
	t.Cleanup(gock.Off)

	expectedBody, err := json.Marshal(map[string]interface{}{
		"query":     nextNumberQuery,
		"variables": map[string]interface{}{"owner": repo.Owner, "name": repo.Name},
	})
	require.NoError(t, err)

	gock.New("https://api.github.com").
		Post("/graphql").
		MatchHeader("Authorization", "token abc123").
		BodyString(string(expectedBody)).
		Reply(200).
		JSON(responseJSON)
}

func TestNextPullRequest(t *testing.T) {
	repo := repository.Repository{
		Host:  "github.com",
		Owner: "edgarrmondragon",
		Name:  "gh-next-pr",
	}

	mockGraphQL(t, repo, `{
  "data": {
    "repository": {
      "discussions": {
        "nodes": [
          {
            "number": 100
          }
        ]
      },
      "issues": {
        "nodes": [
          {
            "number": 10
          }
        ]
      },
      "pullRequests": {
        "nodes": [
          {
            "number": 1000
          }
        ]
      }
    }
  }
}`)

	next, err := getNextNumber(repo)
	assert.NoError(t, err)
	assert.Equal(t, 1001, next)
	assert.True(t, gock.IsDone(), "unmatched mocks: %v", gock.Pending())
}

func TestNextPullRequestFreshRepo(t *testing.T) {
	repo := repository.Repository{
		Host:  "github.com",
		Owner: "edgarrmondragon",
		Name:  "gh-next-pr",
	}
	mockGraphQL(t, repo, `{
  "data": {
    "repository": {
      "discussions": { "nodes": [] },
      "issues": { "nodes": [] },
      "pullRequests": { "nodes": [] }
    }
  }
}`)

	next, err := getNextNumber(repo)
	assert.NoError(t, err)
	assert.Equal(t, 1, next)
	assert.True(t, gock.IsDone(), "unmatched mocks: %v", gock.Pending())
}

var lastNumberTests = []struct {
	name  string
	nodes []node
	want  int
}{
	{"empty", []node{}, 0},
	{"first-wins", []node{{Number: 33}, {Number: 44}}, 33},
	{"non-empty", []node{{Number: 10}}, 10},
}

func TestLastNumber(t *testing.T) {
	for _, tt := range lastNumberTests {
		t.Run(tt.name, func(t *testing.T) {
			got := lastNumber(tt.nodes)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
