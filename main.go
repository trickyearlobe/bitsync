package main

import (
  "fmt"
  "github.com/ktrysmt/go-bitbucket"
  "os"
  "os/exec"
  "path/filepath"
)

var client *bitbucket.Client

func main() {
  // Check we have env vars set
  if os.Getenv("BBUSER") == "" || os.Getenv("BBAPPPASS") == "" {
    fmt.Println("Please set credential in environment vars BBUSER and BBAPPPASS")
    fmt.Println("  export BBUSER=ebeneezer")
    fmt.Println("  export BBAPPPASS=ijfewIIejdiiowIOEIJEDiojoewfiJIOEF")
    return
  }

  // Iterate the workspaces we have access to
  client = bitbucket.NewBasicAuth(os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"))
  workspaces, err := client.Workspaces.List()
  checkErr(err)
  for _, workspace := range workspaces.Workspaces {
    processWorkspace(workspace)
  }
}

func processWorkspace(workkspace bitbucket.Workspace) {
  fmt.Printf("Processing workspace %v\n", workkspace.Slug)
  options := bitbucket.RepositoriesOptions{
    Owner: workkspace.Slug,
  }
  repos, err := client.Repositories.ListForAccount(&options)
  checkErr(err)

  for _, repo := range repos.Items {
    processRepo(workkspace.Slug, repo.Project.Key, repo.Slug, repo.Mainbranch.Name)
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
  out, err := exec.Command("git", "-C", repoDir, "fetch", "origin", mainBranch).CombinedOutput()
  if err != nil {
    fmt.Printf("    Error: %v\n", string(out))
  }
}

func cloneRepo(repoDir, workspace, project, repo string) {
  fmt.Printf("  Cloning %v/%v/%v into %v\n", workspace, project, repo, repoDir)
  out, err := exec.Command("git", "clone", "git@bitbucket.org:"+workspace+"/"+repo, repoDir).CombinedOutput()
  if err != nil {
    fmt.Printf("    Error: %v\n", string(out))
  }
}

func checkErr(err error) {
  if err != nil {
    fmt.Println("An error occured. %v", err)
    os.Exit(1)
  }
}
