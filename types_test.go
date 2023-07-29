package main

import (
    "os"
    "testing"
)

var organisations []string

func TestFetchOrganisations(t *testing.T) {
    organisations := fetchOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    if len(organisations) < 1 {
        t.Fatalf(`Length of workspaces is %v`, len(organisations))
    }
}

func TestFetchRepos(t *testing.T) {
    organisations := fetchOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    repos := FetchBitbucketRepos(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"), organisations[0])
    if len(repos) < 1 {
        t.Fatalf(`Length of repositories is %v`, len(organisations))
    }
}
