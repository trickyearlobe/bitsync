package main

import (
    "os"
    "testing"
)

func TestFetchGitHubOrganisations(t *testing.T) {
    organisations := FetchGitHubOrganisations(os.Getenv("GHTOKEN"))
    if len(organisations) < 1 {
        t.Fatalf(`Length of workspaces is %v`, len(organisations))
    }
}

func TestFetchGitHubRepos(t *testing.T) {
    organisations := FetchGitHubOrganisations(os.Getenv("GHTOKEN"))
    repos := FetchGitHubRepos(os.Getenv("GHTOKEN"), organisations[0])
    if len(repos) < 1 {
        t.Fatalf(`Length of repositories is %v`, len(organisations))
    }
}

func TestFetchBitBucketOrganisations(t *testing.T) {
    organisations := FetchBitBucketOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    if len(organisations) < 1 {
        t.Fatalf(`Length of workspaces is %v`, len(organisations))
    }
}

func TestFetchRepos(t *testing.T) {
    organisations := FetchBitBucketOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    repos := FetchBitbucketRepos(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"), organisations[0])
    if len(repos) < 1 {
        t.Fatalf(`Length of repositories is %v`, len(organisations))
    }
}
