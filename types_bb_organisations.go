package main

import (
    "encoding/json"
    "net/http"
    "time"
)

func FetchBitBucketOrganisations(user, password string) ([]string, error) {
    var slugs []string
    url := "https://api.bitbucket.org/2.0/workspaces?pagelen=100"
    for url != "" {
        page, err := fetchBitbucketOrgsPage(user, password, url)
        if err != nil {
            return nil, err
        }
        for _, org := range page.Values {
            slugs = append(slugs, org.Slug)
        }
        url = page.Next
    }
    return slugs, nil
}

func fetchBitbucketOrgsPage(user, password, url string) (BitBucketOrganisationsResponse, error) {
    var page BitBucketOrganisationsResponse
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return page, err
    }
    req.SetBasicAuth(user, password)
    req.Header.Set("Accept", "application/json")
    body, _, err := fetchAPI(req)
    if err != nil {
        return page, err
    }
    if err := json.Unmarshal(body, &page); err != nil {
        return page, err
    }
    return page, nil
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
