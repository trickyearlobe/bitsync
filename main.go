package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

func main() {
    // Check we have env vars set
    if os.Getenv("BBUSER") == "" || os.Getenv("BBAPPPASS") == "" {
        fmt.Println("Please set credential in environment vars BBUSER and BBAPPPASS")
        fmt.Println("  export BBUSER=ebeneezer")
        fmt.Println("  export BBAPPPASS=ijfewIIejdiiowIOEIJEDiojoewfiJIOEF")
        return
    }

    // Iterate the workspaces we have access to
    workspaces := fetchOrganisations(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
    for _, workspace := range workspaces {
        processWorkspace(workspace)
    }
}

func processWorkspace(workspace string) {
    fmt.Printf("Processing workspace %v\n", workspace)
    repos := FetchBitbucketRepos(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"), workspace)

    for _, repo := range repos {
        processRepo(workspace, repo.Project.Key, repo.Slug, repo.Mainbranch.Name)
    }
    fmt.Println("DONE")
}

func processRepo(workspace, project, repo, mainBranch string) {
    //fmt.Printf("  Repo: %v/%v/%v\n", workspace, project, repo)
    home, err := os.UserHomeDir()
    checkErr(err)

    projectDir := filepath.Join(home, "repos", workspace, project)
    repoDir := filepath.Join(home, "repos", workspace, project, repo)

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
