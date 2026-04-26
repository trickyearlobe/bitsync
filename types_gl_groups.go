package main

import (
	"encoding/json"
	"net/http"
)

func FetchGitLabGroups(baseURL, apiToken string) ([]string, error) {
	var paths []string
	url := baseURL + "/api/v4/groups?membership=true&per_page=100&top_level_only=true"
	for url != "" {
		page, next, err := fetchGitLabGroupsPage(apiToken, url)
		if err != nil {
			return nil, err
		}
		for _, g := range page {
			paths = append(paths, g.FullPath)
		}
		url = next
	}
	return paths, nil
}

func fetchGitLabGroupsPage(apiToken, url string) (GitLabGroups, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("PRIVATE-TOKEN", apiToken)
	req.Header.Set("Accept", "application/json")
	body, next, err := fetchAPI(req)
	if err != nil {
		return nil, "", err
	}
	var page GitLabGroups
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, "", err
	}
	return page, next, nil
}

type GitLabGroups []GitLabGroup

type GitLabGroup struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	FullPath   string `json:"full_path"`
	WebURL     string `json:"web_url"`
	Visibility string `json:"visibility"`
	ParentId   int64  `json:"parent_id"`
}
