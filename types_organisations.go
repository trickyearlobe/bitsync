package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "time"
)

func fetchOrganisations(user, password string) []string {
    var organisations []string
    url := "https://api.bitbucket.org/2.0/workspaces?pagelen=100"
    for url != "" {
        orgPage := fetchBitbucketPage(user, password, url)
        for _, organisation := range orgPage.Values {
            organisations = append(organisations, organisation.Slug)
        }
        url = orgPage.Next
    }
    return organisations
}

func fetchBitbucketPage(user, password, url string) BitbucketOrganisationsResponse {
    var bitbucketOrganisations BitbucketOrganisationsResponse
    client := http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    checkErr(err)
    req.SetBasicAuth(user, password)
    req.Header.Set("Accept", "application/json")
    resp, err := client.Do(req)
    checkErr(err)
    defer resp.Body.Close()
    bodyText, err := ioutil.ReadAll(resp.Body)
    checkErr(err)
    json.Unmarshal(bodyText, &bitbucketOrganisations)
    return bitbucketOrganisations
}

type BitbucketOrganisationsResponse struct {
    Values  []bitbucketOrganisation `json:"values"`
    Pagelen int                     `json:"pagelen"`
    Size    int                     `json:"size"`
    Page    int                     `json:"page"`
    Next    string                  `json:"next"`
}

type bitbucketOrganisation struct {
    Type      string `json:"type"`
    Uuid      string `json:"uuid"`
    Name      string `json:"name"`
    Slug      string `json:"slug"`
    IsPrivate bool   `json:"is_private"`
    Links     struct {
        Avatar       bitbucketOrganisationHref `json:"avatar"`
        Hooks        bitbucketOrganisationHref `json:"hooks"`
        Html         bitbucketOrganisationHref `json:"html"`
        HtmlOverview bitbucketOrganisationHref `json:"html_overview"`
        Members      bitbucketOrganisationHref `json:"members"`
        Owners       bitbucketOrganisationHref `json:"owners"`
        Projects     bitbucketOrganisationHref `json:"projects"`
        Repositories bitbucketOrganisationHref `json:"repositories"`
        Snippets     bitbucketOrganisationHref `json:"snippets"`
        Self         bitbucketOrganisationHref `json:"self"`
    } `json:"links"`
    CreatedOn time.Time `json:"created_on"`
}

type bitbucketOrganisationHref struct {
    Href string `json:"href"`
}
