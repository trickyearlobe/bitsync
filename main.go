package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

func cloneOrSyncGitRepo(repoDir, cloneUrl, mainBranch string) {
    fmt.Printf("  Processing git repo %s from %s\n", repoDir, cloneUrl)
    if _, err := os.Stat(repoDir); err != nil {
        out, err := exec.Command("git", "clone", cloneUrl, repoDir).CombinedOutput()
        if err != nil {
            fmt.Println("Error during clone: " + string(out))
        }
    }
    out, err := exec.Command("git", "-C", repoDir, "reset", "--hard").CombinedOutput()
    if err != nil {
        fmt.Println("Error during hard reset: " + string(out))
    }
    out, err = exec.Command("git", "-C", repoDir, "fetch", "--all", "--prune").CombinedOutput()
    if err != nil {
        fmt.Println("Error during fetch all: " + string(out))
    }
    out, err = exec.Command("git", "-C", repoDir, "checkout", mainBranch).CombinedOutput()
    if err != nil {
        fmt.Println("Error during checkout: " + string(out))
    }
    out, err = exec.Command("git", "-C", repoDir, "pull").CombinedOutput()
    if err != nil {
        fmt.Println("Error during pull: " + string(out))
    }
}

func processGitHubRepo(repo GitHubRepository) {
    homeDir, err := os.UserHomeDir()
    checkErr(err)
    gitHubOrgPath := filepath.Join(homeDir, "repos", "github", repo.Owner.Login)
    gitHubRepoPath := filepath.Join(homeDir, "repos", "github", repo.Owner.Login, repo.Name)
    err = os.MkdirAll(gitHubOrgPath, 0750)
    checkErr(err)
    cloneOrSyncGitRepo(gitHubRepoPath, repo.SSHUrl, repo.DefaultBranch)
}

func processGitHubOrg(token, org string) {
    fmt.Printf("Processing repos in GitHub Org %v\n", org)
    repos := FetchGitHubRepos(token, org)
    for _, repo := range repos {
        processGitHubRepo(repo)
    }
}

func processGitHubOrgs() {
    ghtoken := os.Getenv("GHTOKEN")
    if ghtoken == "" {
        fmt.Println("GHTOKEN is not set, skipping GitHub repositories")
    } else {
        fmt.Println("GHTOKEN is set, processing GitHub repositories")
        var ghorgs []string
        ghorg := os.Getenv("GHORG")
        if ghorg != "" {
            fmt.Printf("GHORG is set, processing selected GitHub orgs\n")
            ghorgs = strings.Split(ghorg, ",")
        } else {
            fmt.Printf("GHORG is not set, processing all GitHub orgs\n")
            ghorgs = FetchGitHubOrganisations(ghtoken)
        }
        for _, ghorg := range ghorgs {
            processGitHubOrg(ghtoken, ghorg)
        }
    }
}

func processBitBucketRepo(workspace string, repo BitbucketRepository) {
    homeDir, err := os.UserHomeDir()
    checkErr(err)
    BitBucketProjectPath := filepath.Join(homeDir, "repos", "bitbucket", workspace, repo.Project.Key)
    BitBucketRepoPath := filepath.Join(homeDir, "repos", "bitbucket", workspace, repo.Project.Key, repo.Slug)
    BitBucketCloneUrl := "git@bitbucket.org:" + workspace + "/" + repo.Slug
    err = os.MkdirAll(BitBucketProjectPath, 0750)
    checkErr(err)
    cloneOrSyncGitRepo(BitBucketRepoPath, BitBucketCloneUrl, repo.Mainbranch.Name)
}

func processBitBucketWorkspace(bbUser, bbAppPass, bbWorkspace string) {
    fmt.Printf("Processing BitBucket workspace %v\n", bbWorkspace)
    repos := FetchBitbucketRepos(bbUser, bbAppPass, bbWorkspace)
    for _, repo := range repos {
        processBitBucketRepo(bbWorkspace, repo)
    }
}

func processBitBucketWorkspaces() {
    bbuser := os.Getenv("BBUSER")
    bbapppass := os.Getenv("BBAPPPASS")
    bborg := os.Getenv("BBORG")

    if bbuser == "" || bbapppass == "" {
        fmt.Println("BBUSER and/or BBAPPPASS not set, skipping BitBucket repositories")
    } else {
        fmt.Println("BBUSER and BBAPPPASS are set, processing BitBucket repositories")
        var bbWorkspaces []string
        if bborg == "" {
            fmt.Println("BBORG is not set, processing all BitBucket Workspaces")
            bbWorkspaces = FetchBitBucketOrganisations(bbuser, bbapppass)
        } else {
            fmt.Println("BBORG is set, processing selected BitBucket Workspaces")
            bbWorkspaces = strings.Split(bborg, ",")
        }
        for _, bbWorkspace := range bbWorkspaces {
            processBitBucketWorkspace(bbuser, bbapppass, bbWorkspace)
        }
    }
}

func checkErr(err error) {
    if err != nil {
        fmt.Printf("An error occured. %v\n", err)
        os.Exit(1)
    }
}

func main() {
    startTime := time.Now()
    fmt.Printf("Starting BitSync process at %v\n", startTime)
    processGitHubOrgs()
    processBitBucketWorkspaces()
    endTime := time.Now()
    fmt.Printf("Finished BitSync process at %v\n", endTime)
    fmt.Printf("Elapsed time %v\n", endTime.Sub(startTime))
}
