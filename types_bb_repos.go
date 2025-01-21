package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Iterates to fetch all the repos
func FetchBitbucketRepos(user string, password string, workspace string) []BitbucketRepository {
	var repositories []BitbucketRepository
	url := "https://api.bitbucket.org/2.0/repositories/" + workspace + "?pagelen=100"
	for url != "" {
		repoPage := fetchBitbucketReposPage(user, password, url)
		repositories = append(repositories, repoPage.Values...)
		url = repoPage.Next
	}
	return repositories
}

// Fetches an individual page of repos
func fetchBitbucketReposPage(user, password, url string) BitbucketRepositoriesResponse {
	var bitbucketResponse BitbucketRepositoriesResponse
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err)
	req.SetBasicAuth(user, password)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	checkErr(err)
	err = json.Unmarshal(bodyText, &bitbucketResponse)
	checkErr(err)
	return bitbucketResponse
}

type BitbucketRepositoriesResponse struct {
	Values  []BitbucketRepository `json:"values"`
	Pagelen int                   `json:"pagelen"`
	Size    int                   `json:"size"`
	Page    int                   `json:"page"`
	Next    string                `json:"next"`
}

type BitbucketRepository struct {
	Type     string `json:"type"`
	FullName string `json:"full_name"`
	Links    struct {
		Self         bitbucketRepoHref `json:"self"`
		Html         bitbucketRepoHref `json:"html"`
		Avatar       bitbucketRepoHref `json:"avatar"`
		Pullrequests bitbucketRepoHref `json:"pullrequests"`
		Commits      bitbucketRepoHref `json:"commits"`
		Forks        bitbucketRepoHref `json:"forks"`
		Watchers     bitbucketRepoHref `json:"watchers"`
		Branches     bitbucketRepoHref `json:"branches"`
		Tags         bitbucketRepoHref `json:"tags"`
		Downloads    bitbucketRepoHref `json:"downloads"`
		Source       bitbucketRepoHref `json:"source"`
		Clone        []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
		Hooks bitbucketRepoHref `json:"hooks"`
	} `json:"links"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Description string      `json:"description"`
	Scm         string      `json:"scm"`
	Website     interface{} `json:"website"`
	Owner       struct {
		DisplayName string `json:"display_name"`
		Links       struct {
			Self   bitbucketRepoHref `json:"self"`
			Avatar bitbucketRepoHref `json:"avatar"`
			Html   bitbucketRepoHref `json:"html"`
		} `json:"links"`
		Type     string `json:"type"`
		Uuid     string `json:"uuid"`
		Username string `json:"username"`
	} `json:"owner"`
	Workspace struct {
		Type  string `json:"type"`
		Uuid  string `json:"uuid"`
		Name  string `json:"name"`
		Slug  string `json:"slug"`
		Links struct {
			Avatar bitbucketRepoHref `json:"avatar"`
			Html   bitbucketRepoHref `json:"html"`
			Self   bitbucketRepoHref `json:"self"`
		} `json:"links"`
	} `json:"workspace"`
	IsPrivate bool `json:"is_private"`
	Project   struct {
		Type  string `json:"type"`
		Key   string `json:"key"`
		Uuid  string `json:"uuid"`
		Name  string `json:"name"`
		Links struct {
			Self   bitbucketRepoHref `json:"self"`
			Html   bitbucketRepoHref `json:"html"`
			Avatar bitbucketRepoHref `json:"avatar"`
		} `json:"links"`
	} `json:"project"`
	ForkPolicy string    `json:"fork_policy"`
	CreatedOn  time.Time `json:"created_on"`
	UpdatedOn  time.Time `json:"updated_on"`
	Size       int       `json:"size"`
	Language   string    `json:"language"`
	HasIssues  bool      `json:"has_issues"`
	HasWiki    bool      `json:"has_wiki"`
	Uuid       string    `json:"uuid"`
	Mainbranch struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"mainbranch"`
	OverrideSettings struct {
		DefaultMergeStrategy bool `json:"default_merge_strategy"`
		BranchingModel       bool `json:"branching_model"`
	} `json:"override_settings"`
}

type bitbucketRepoHref struct {
	Href string `json:"href"`
}
