package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

func cloneOrSyncGitRepo(repoDir, cloneUrl, mainBranch string) {
    fmt.Printf("  Processing git repo %s from %s\n", repoDir, cloneUrl)
    if _, err := os.Stat(repoDir); err != nil {
        fmt.Println("    Cloning...")
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

func processBitBucketOrgs() {
    bbuser := os.Getenv("BBUSER")
    bbapppass := os.Getenv("BBAPPPASS")
    bborg := os.Getenv("BBORG")

    if bbuser == "" || bbapppass == "" {
        fmt.Println("BBUSER and/or BBAPPPASS not set, skipping BitBucket repositories")
    } else {
        var bborgs []string
        if bborg == "" {
            fmt.Println("BBORG is not set, processing all BitBucket Organisations")
            bborgs = FetchBitBucketOrganisations(bbuser, bbapppass)
        } else {
            fmt.Println("BBORG is set, processing selected BitBucket repositories")
            bborgs = strings.Split(bborg, ",")
        }
        for _, bborg := range bborgs {
            processWorkspace(bborg)
        }
    }
}

func main() {
    processGitHubOrgs()
    processBitBucketOrgs()
}

func processWorkspace(workspace string) {
    fmt.Printf("Processing BitBucket workspace %v\n", workspace)
    repos := FetchBitbucketRepos(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"), workspace)

    for _, repo := range repos {
        processRepo(workspace, repo.Project.Key, repo.Slug, repo.Mainbranch.Name)
    }
    fmt.Println("DONE")
}

func processRepo(workspace, project, repo, mainBranch string) {
    home, err := os.UserHomeDir()
    checkErr(err)

    projectDir := filepath.Join(home, "repos", workspace, project)
    repoDir := filepath.Join(home, "repos", "bitbucket", workspace, project, repo)

    // Make the project directory
    err = os.MkdirAll(projectDir, 0750)
    checkErr(err)

    // Check if the repo is cloned
    if _, err = os.Stat(repoDir); err != nil {
        cloneRepo(repoDir, workspace, project, repo)
    } else {
        pullRepo(repoDir, mainBranch)
    }

}

func pullRepo(repoDir, mainBranch string) {
    fmt.Printf("  Pulling %v branch %v\n", repoDir, mainBranch)
    out, err := exec.Command("git", "-C", repoDir, "pull", "origin", mainBranch+":"+mainBranch).CombinedOutput()
    if err != nil {
        fmt.Printf("%v\n", string(out))
    }
}

func cloneRepo(repoDir, workspace, project, repo string) {
    fmt.Printf("  Cloning %v/%v/%v into %v\n", workspace, project, repo, repoDir)
    out, err := exec.Command("git", "clone", "git@bitbucket.org:"+workspace+"/"+repo, repoDir).CombinedOutput()
    if err != nil {
        fmt.Printf("%v\n", string(out))
    }
}

func checkErr(err error) {
    if err != nil {
        fmt.Printf("An error occured. %v\n", err)
        os.Exit(1)
    }
}
