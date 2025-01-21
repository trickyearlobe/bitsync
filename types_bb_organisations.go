package main

import (
    "encoding/json"
    "io"
    "net/http"
    "time"
)

func FetchBitBucketOrganisations(user, password string) []string {
    var organisations []string
    url := "https://api.bitbucket.org/2.0/workspaces?pagelen=100"
    for url != "" {
        orgPage := FetchBitbucketPage(user, password, url)
        for _, organisation := range orgPage.Values {
            organisations = append(organisations, organisation.Slug)
        }
        url = orgPage.Next
    }
    return organisations
}

func FetchBitbucketPage(user, password, url string) BitBucketOrganisationsResponse {
    var bitbucketOrganisations BitBucketOrganisationsResponse
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
    json.Unmarshal(bodyText, &bitbucketOrganisations)
    return bitbucketOrganisations
}

type BitBucketOrganisationsResponse struct {
    Values  []BitBucketOrganisation `json:"values"`
    Pagelen int                     `json:"pagelen"`
    Size    int                     `json:"size"`
    Page    int                     `json:"page"`
    Next    string                  `json:"next"`
}

type BitBucketOrganisation struct {
    Type      string `json:"type"`
    Uuid      string `json:"uuid"`
    Name      string `json:"name"`
    Slug      string `json:"slug"`
    IsPrivate bool   `json:"is_private"`
    Links     struct {
        Avatar       BitBucketOrganisationHref `json:"avatar"`
        Hooks        BitBucketOrganisationHref `json:"hooks"`
        Html         BitBucketOrganisationHref `json:"html"`
        HtmlOverview BitBucketOrganisationHref `json:"html_overview"`
        Members      BitBucketOrganisationHref `json:"members"`
        Owners       BitBucketOrganisationHref `json:"owners"`
        Projects     BitBucketOrganisationHref `json:"projects"`
        Repositories BitBucketOrganisationHref `json:"repositories"`
        Snippets     BitBucketOrganisationHref `json:"snippets"`
        Self         BitBucketOrganisationHref `json:"self"`
    } `json:"links"`
    CreatedOn time.Time `json:"created_on"`
}

type BitBucketOrganisationHref struct {
    Href string `json:"href"`
}
