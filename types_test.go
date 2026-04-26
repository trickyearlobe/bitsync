package main

import (
    "os"
    "testing"
)

func TestFetchGitHubOrganisations(t *testing.T) {
    organisations, err := FetchGitHubOrganisations(os.Getenv("GHTOKEN"))
    if err != nil {
        t.Fatalf("FetchGitHubOrganisations: %v", err)
    }
    if len(organisations) < 1 {
        t.Fatalf(`Length of workspaces is %v`, len(organisations))
    }
}

func TestFetchGitHubRepos(t *testing.T) {
    organisations, err := FetchGitHubOrganisations(os.Getenv("GHTOKEN"))
    if err != nil {
        t.Fatalf("FetchGitHubOrganisations: %v", err)
    }
    repos, err := FetchGitHubRepos(os.Getenv("GHTOKEN"), organisations[0])
    if err != nil {
        t.Fatalf("FetchGitHubRepos: %v", err)
    }
    if len(repos) < 1 {
        t.Fatalf(`Length of repositories is %v`, len(organisations))
    }
}

func TestFetchBitBucketOrganisations(t *testing.T) {
    organisations, err := FetchBitBucketOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    if err != nil {
        t.Fatalf("FetchBitBucketOrganisations: %v", err)
    }
    if len(organisations) < 1 {
        t.Fatalf(`Length of workspaces is %v`, len(organisations))
    }
}

func TestFetchRepos(t *testing.T) {
    organisations, err := FetchBitBucketOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    if err != nil {
        t.Fatalf("FetchBitBucketOrganisations: %v", err)
    }
    repos, err := FetchBitbucketRepos(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"), organisations[0])
    if err != nil {
        t.Fatalf("FetchBitbucketRepos: %v", err)
    }
    if len(repos) < 1 {
        t.Fatalf(`Length of repositories is %v`, len(organisations))
    }
}
