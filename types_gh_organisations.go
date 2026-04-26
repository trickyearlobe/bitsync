package main

import (
    "encoding/json"
    "net/http"
)

func FetchGitHubOrganisations(apiToken string) ([]string, error) {
    var names []string
    url := "https://api.github.com/user/orgs?per_page=100"
    for url != "" {
        page, next, err := fetchGitHubOrgsPage(apiToken, url)
        if err != nil {
            return nil, err
        }
        for _, org := range page {
            names = append(names, org.Login)
        }
        url = next
    }
    return names, nil
}

func fetchGitHubOrgsPage(apiToken, url string) (GitHubOrganisationsResponse, string, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, "", err
    }
    req.Header.Set("Authorization", "bearer "+apiToken)
    req.Header.Set("Accept", "application/vnd.github+json")
    body, next, err := fetchAPI(req)
    if err != nil {
        return nil, "", err
    }
    var page GitHubOrganisationsResponse
    if err := json.Unmarshal(body, &page); err != nil {
        return nil, "", err
    }
    return page, next, nil
}

type GitHubOrganisationsResponse []GitHubOrganisation

type GitHubOrganisation struct {
    Login            string `json:"login"`
    Id               int    `json:"id"`
    NodeId           string `json:"node_id"`
    Url              string `json:"url"`
    ReposUrl         string `json:"repos_url"`
    EventsUrl        string `json:"events_url"`
    HooksUrl         string `json:"hooks_url"`
    IssuesUrl        string `json:"issues_url"`
    MembersUrl       string `json:"members_url"`
    PublicMembersUrl string `json:"public_members_url"`
    AvatarUrl        string `json:"avatar_url"`
    Description      string `json:"description"`
}
