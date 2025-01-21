package main

import (
    "encoding/json"
    "io"
    "net/http"
)

func FetchGitHubOrganisations(apiToken string) []string {
    var organisations GitHubOrganisationsResponse
    url := "https://api.github.com/user/orgs?per_page=100"
    organisations = FetchGitHubPage(apiToken, url)
    organisationNames := make([]string, len(organisations))
    for i, organisation := range organisations {
        organisationNames[i] = organisation.Login
    }
    return organisationNames
}

func FetchGitHubPage(apiToken, url string) GitHubOrganisationsResponse {
    var Organisations GitHubOrganisationsResponse
    client := http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    checkErr(err)
    req.Header.Set("Authorization", "bearer "+apiToken)
    req.Header.Set("Accept", "application/vnd.github+json")
    resp, err := client.Do(req)
    checkErr(err)
    defer resp.Body.Close()
    bodyText, err := io.ReadAll(resp.Body)
    checkErr(err)
    json.Unmarshal(bodyText, &Organisations)
    return Organisations
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
