package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func mirrorGitRepo(repoDir, cloneUrl, mainBranch string) {
	out, err := exec.Command("rm", "-rf", repoDir).CombinedOutput()
	if err != nil {
		fmt.Println("  Error during delete directory: " + string(out))
	}
	out, err = exec.Command("git", "clone", "--mirror", cloneUrl, repoDir).CombinedOutput()
	if err != nil {
		fmt.Printf("  Error mirroring %s:\n%s\n ", cloneUrl, string(out))
	} else {
		fmt.Printf("  Bare mirror success for %s\n", cloneUrl)
	}
}

func syncGitRepo(repoDir, gitUrl, mainBranch string) {
	mirror := os.Getenv("BITSYNC_MIRROR")
	if mirror == "true" {
		mirrorGitRepo(repoDir, gitUrl, mainBranch)
		return
	}
	// fmt.Printf("  Processing git repo %s from %s\n", repoDir, gitUrl)
	if _, err := os.Stat(repoDir); err != nil {
		out, err := exec.Command("git", "clone", gitUrl, repoDir).CombinedOutput()
		if err != nil {
			fmt.Printf("  Error during clone of %s:\n%s\n", gitUrl, string(out))
			return
		} else {
			fmt.Printf("  Clone successful for %s\n", gitUrl)
			return
		}
	}
	out, err := exec.Command("git", "-C", repoDir, "reset", "--hard").CombinedOutput()
	if err != nil {
		fmt.Printf("  Error during hard reset in %s:\n%s\n", repoDir, string(out))
	}
	out, err = exec.Command("git", "-C", repoDir, "fetch", "--all", "--prune").CombinedOutput()
	if err != nil {
		fmt.Printf("  Error during fetch %s:\n%s\n ", gitUrl, string(out))
	} else {
		fmt.Printf("  Fetch successful for %s\n", gitUrl)
	}
	out, err = exec.Command("git", "-C", repoDir, "checkout", mainBranch).CombinedOutput()
	if err != nil {
		fmt.Printf("  Error during checkout %s in %s:\n%s\n", mainBranch, repoDir, string(out))
	}
	out, err = exec.Command("git", "-C", repoDir, "pull").CombinedOutput()
	if err != nil {
		fmt.Printf("  Error during pull in %s:\n%s\n", repoDir, string(out))
	}
}

func processGitHubRepo(repo GitHubRepository) {
	homeDir, err := os.UserHomeDir()
	checkErr(err)
	var gitHubRepoPath, gitHubOrgPath string
	mirror := os.Getenv("BITSYNC_MIRROR")
	if mirror == "true" {
		gitHubOrgPath = filepath.Join(homeDir, "mirrors", "github", repo.Owner.Login)
		gitHubRepoPath = filepath.Join(homeDir, "mirrors", "github", repo.Owner.Login, repo.Name)

	} else {
		gitHubOrgPath = filepath.Join(homeDir, "repos", "github", repo.Owner.Login)
		gitHubRepoPath = filepath.Join(homeDir, "repos", "github", repo.Owner.Login, repo.Name)
	}
	err = os.MkdirAll(gitHubOrgPath, 0750)
	checkErr(err)
	syncGitRepo(gitHubRepoPath, repo.SSHUrl, repo.DefaultBranch)
}

func getWorkerCount() int {
	workersStr := os.Getenv("BITSYNC_WORKERS")
	if workersStr == "" {
		return 6
	}
	workers, err := strconv.Atoi(workersStr)
	if err != nil || workers < 1 {
		return 4
	}
	return workers
}

func processGitHubOrg(token, org string) {
	fmt.Printf("Processing repos in GitHub Org %v\n", org)
	repos := FetchGitHubRepos(token, org)
	workerCount := getWorkerCount()
	repoCh := make(chan GitHubRepository)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range repoCh {
				processGitHubRepo(repo)
			}
		}()
	}
	for _, repo := range repos {
		repoCh <- repo
	}
	close(repoCh)
	wg.Wait()
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
	var BitBucketRepoPath, BitBucketProjectPath string
	mirror := os.Getenv("BITSYNC_MIRROR")
	if mirror == "true" {
		BitBucketProjectPath = filepath.Join(homeDir, "mirrors", "bitbucket", workspace, repo.Project.Key)
		BitBucketRepoPath = filepath.Join(homeDir, "mirrors", "bitbucket", workspace, repo.Project.Key, repo.Slug)

	} else {
		BitBucketProjectPath = filepath.Join(homeDir, "repos", "bitbucket", workspace, repo.Project.Key)
		BitBucketRepoPath = filepath.Join(homeDir, "repos", "bitbucket", workspace, repo.Project.Key, repo.Slug)
	}
	BitBucketCloneUrl := "git@bitbucket.org:" + workspace + "/" + repo.Slug
	err = os.MkdirAll(BitBucketProjectPath, 0750)
	checkErr(err)
	syncGitRepo(BitBucketRepoPath, BitBucketCloneUrl, repo.Mainbranch.Name)
}

func processBitBucketWorkspace(bbUser, bbAppPass, bbWorkspace string) {
	fmt.Printf("Processing BitBucket workspace %v\n", bbWorkspace)
	repos := FetchBitbucketRepos(bbUser, bbAppPass, bbWorkspace)
	workerCount := getWorkerCount()
	repoCh := make(chan BitbucketRepository)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range repoCh {
				processBitBucketRepo(bbWorkspace, repo)
			}
		}()
	}
	for _, repo := range repos {
		repoCh <- repo
	}
	close(repoCh)
	wg.Wait()
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
