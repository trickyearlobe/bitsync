package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Populated at build time via -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "unknown"
)

type apiError struct {
	Method     string
	URL        string
	Status     string
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s %s: %s: %s", e.Method, e.URL, e.Status, e.Body)
}

func is404(err error) bool {
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// fetchAPI performs the HTTP request, validates the status, and returns the
// body plus the URL extracted from a `Link: <...>; rel="next"` header (empty
// string if absent). Bitbucket carries pagination in the body, so callers there
// ignore the returned next-link.
func fetchAPI(req *http.Request) ([]byte, string, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", &apiError{
			Method:     req.Method,
			URL:        req.URL.String(),
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}
	return body, parseNextLink(resp.Header.Get("Link")), nil
}

func parseNextLink(header string) string {
	for _, link := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(parts[0])
		urlPart = strings.TrimPrefix(urlPart, "<")
		urlPart = strings.TrimSuffix(urlPart, ">")
		for _, param := range parts[1:] {
			if strings.TrimSpace(param) == `rel="next"` {
				return urlPart
			}
		}
	}
	return ""
}

func processConcurrently[T any](items []T, workers int, fn func(T)) {
	ch := make(chan T)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				fn(item)
			}
		}()
	}
	for _, item := range items {
		ch <- item
	}
	close(ch)
	wg.Wait()
}

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

func git(repoDir string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoDir}, args...)
	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	return string(out), err
}

// syncGitRepo updates a working tree's main branch to match origin without
// silently destroying the user's in-progress work. If the user is on a
// different branch, that branch is left untouched. If the working tree is
// dirty, changes are stashed for the duration of the sync and popped after.
func syncGitRepo(repoDir, gitUrl, mainBranch string) {
	if os.Getenv("BITSYNC_MIRROR") == "true" {
		mirrorGitRepo(repoDir, gitUrl, mainBranch)
		return
	}

	if _, err := os.Stat(repoDir); err != nil {
		out, err := exec.Command("git", "clone", gitUrl, repoDir).CombinedOutput()
		if err != nil {
			fmt.Printf("  Error during clone of %s:\n%s\n", gitUrl, string(out))
			return
		}
		fmt.Printf("  Clone successful for %s\n", gitUrl)
		return
	}

	if mainBranch == "" {
		fmt.Printf("  Skipping %s: no default branch reported\n", repoDir)
		return
	}

	branchOut, err := git(repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Printf("  Error reading current branch in %s:\n%s\n", repoDir, branchOut)
		return
	}
	currentBranch := strings.TrimSpace(branchOut)
	if currentBranch == "HEAD" {
		fmt.Printf("  Skipping %s: detached HEAD state\n", repoDir)
		return
	}

	statusOut, err := git(repoDir, "status", "--porcelain")
	if err != nil {
		fmt.Printf("  Error reading status in %s:\n%s\n", repoDir, statusOut)
		return
	}
	dirty := strings.TrimSpace(statusOut) != ""

	var stashed bool
	if dirty {
		out, err := git(repoDir, "stash", "push", "--include-untracked", "--message", "bitsync auto-stash")
		if err != nil {
			fmt.Printf("  Error stashing local changes in %s:\n%s\n", repoDir, out)
			return
		}
		stashed = !strings.Contains(out, "No local changes to save")
	}

	defer func() {
		if currentBranch != mainBranch {
			if out, err := git(repoDir, "checkout", currentBranch); err != nil {
				fmt.Printf("  Error returning to branch %s in %s:\n%s\n", currentBranch, repoDir, out)
				return
			}
		}
		if stashed {
			if out, err := git(repoDir, "stash", "pop"); err != nil {
				fmt.Printf("  Could not pop bitsync auto-stash in %s — your changes remain in `git stash list`:\n%s\n", repoDir, out)
			}
		}
	}()

	if out, err := git(repoDir, "fetch", "--all", "--prune"); err != nil {
		fmt.Printf("  Error during fetch %s:\n%s\n", gitUrl, out)
		return
	}
	fmt.Printf("  Fetch successful for %s\n", gitUrl)

	if currentBranch != mainBranch {
		if out, err := git(repoDir, "checkout", mainBranch); err != nil {
			fmt.Printf("  Error during checkout %s in %s:\n%s\n", mainBranch, repoDir, out)
			return
		}
	}

	if out, err := git(repoDir, "reset", "--hard", "origin/"+mainBranch); err != nil {
		fmt.Printf("  Error force-updating %s in %s:\n%s\n", mainBranch, repoDir, out)
		return
	}
}

func processGitHubRepo(repo GitHubRepository) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("  Error resolving home directory for %v: %v\n", repo.FullName, err)
		return
	}
	var gitHubRepoPath, gitHubOrgPath string
	mirror := os.Getenv("BITSYNC_MIRROR")
	if mirror == "true" {
		gitHubOrgPath = filepath.Join(homeDir, "mirrors", "github", repo.Owner.Login)
		gitHubRepoPath = filepath.Join(homeDir, "mirrors", "github", repo.Owner.Login, repo.Name)

	} else {
		gitHubOrgPath = filepath.Join(homeDir, "repos", "github", repo.Owner.Login)
		gitHubRepoPath = filepath.Join(homeDir, "repos", "github", repo.Owner.Login, repo.Name)
	}
	if err := os.MkdirAll(gitHubOrgPath, 0750); err != nil {
		fmt.Printf("  Error creating %s: %v\n", gitHubOrgPath, err)
		return
	}
	syncGitRepo(gitHubRepoPath, repo.SSHUrl, repo.DefaultBranch)
}

func getEnvWorkers(name string, defaultValue int) int {
	s := os.Getenv(name)
	if s == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultValue
	}
	return n
}

func getWorkerCount() int    { return getEnvWorkers("BITSYNC_WORKERS", 6) }
func getOrgWorkerCount() int { return getEnvWorkers("BITSYNC_ORG_WORKERS", 2) }

func processGitHubOrg(token, org string) {
	fmt.Printf("Processing repos in GitHub Org %v\n", org)
	repos, err := FetchGitHubRepos(token, org)
	if err != nil {
		fmt.Printf("  Error fetching repos for GitHub org %v: %v\n", org, err)
		return
	}
	processConcurrently(repos, getWorkerCount(), processGitHubRepo)
}

func processGitHubOrgs() {
	ghtoken := os.Getenv("GHTOKEN")
	if ghtoken == "" {
		fmt.Println("GHTOKEN is not set, skipping GitHub repositories")
		return
	}
	fmt.Println("GHTOKEN is set, processing GitHub repositories")
	var ghorgs []string
	ghorg := os.Getenv("GHORG")
	if ghorg != "" {
		fmt.Printf("GHORG is set, processing selected GitHub orgs\n")
		ghorgs = strings.Split(ghorg, ",")
	} else {
		fmt.Printf("GHORG is not set, processing all GitHub orgs\n")
		var err error
		ghorgs, err = FetchGitHubOrganisations(ghtoken)
		if err != nil {
			fmt.Printf("Error fetching GitHub organisations: %v\n", err)
			return
		}
	}
	processConcurrently(ghorgs, getOrgWorkerCount(), func(org string) {
		processGitHubOrg(ghtoken, org)
	})
}

func processBitBucketRepo(workspace string, repo BitbucketRepository) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("  Error resolving home directory for %v: %v\n", repo.FullName, err)
		return
	}
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
	if err := os.MkdirAll(BitBucketProjectPath, 0750); err != nil {
		fmt.Printf("  Error creating %s: %v\n", BitBucketProjectPath, err)
		return
	}
	syncGitRepo(BitBucketRepoPath, BitBucketCloneUrl, repo.Mainbranch.Name)
}

func processBitBucketWorkspace(bbUser, bbAppPass, bbWorkspace string) {
	fmt.Printf("Processing BitBucket workspace %v\n", bbWorkspace)
	repos, err := FetchBitbucketRepos(bbUser, bbAppPass, bbWorkspace)
	if err != nil {
		fmt.Printf("  Error fetching repos for BitBucket workspace %v: %v\n", bbWorkspace, err)
		return
	}
	processConcurrently(repos, getWorkerCount(), func(repo BitbucketRepository) {
		processBitBucketRepo(bbWorkspace, repo)
	})
}

// resolveBitbucketAuth picks credentials for Bitbucket Basic auth, preferring
// the newer BBEMAIL + BBTOKEN pair (Atlassian API tokens) and falling back to
// the legacy BBUSER + BBAPPPASS app-password pair. App passwords are being
// deprecated by Atlassian; both paths are accepted during the transition.
func resolveBitbucketAuth() (identity, secret, method string, ok bool) {
	if email, token := os.Getenv("BBEMAIL"), os.Getenv("BBTOKEN"); email != "" && token != "" {
		return email, token, "BBEMAIL+BBTOKEN", true
	}
	if user, pass := os.Getenv("BBUSER"), os.Getenv("BBAPPPASS"); user != "" && pass != "" {
		return user, pass, "BBUSER+BBAPPPASS", true
	}
	return "", "", "", false
}

func processBitBucketWorkspaces() {
	identity, secret, method, ok := resolveBitbucketAuth()
	if !ok {
		fmt.Println("Neither BBEMAIL+BBTOKEN nor BBUSER+BBAPPPASS set, skipping BitBucket repositories")
		return
	}
	fmt.Printf("Processing BitBucket repositories using %s\n", method)
	var bbWorkspaces []string
	if bborg := os.Getenv("BBORG"); bborg != "" {
		fmt.Println("BBORG is set, processing selected BitBucket Workspaces")
		bbWorkspaces = strings.Split(bborg, ",")
	} else {
		fmt.Println("BBORG is not set, processing all BitBucket Workspaces")
		var err error
		bbWorkspaces, err = FetchBitBucketOrganisations(identity, secret)
		if err != nil {
			fmt.Printf("Error fetching BitBucket workspaces: %v\n", err)
			return
		}
	}
	processConcurrently(bbWorkspaces, getOrgWorkerCount(), func(workspace string) {
		processBitBucketWorkspace(identity, secret, workspace)
	})
}

func processGitLabProject(project GitLabProject) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("  Error resolving home directory for %v: %v\n", project.PathWithNamespace, err)
		return
	}
	var projectPath, namespacePath string
	mirror := os.Getenv("BITSYNC_MIRROR")
	if mirror == "true" {
		namespacePath = filepath.Join(homeDir, "mirrors", "gitlab", project.Namespace.FullPath)
		projectPath = filepath.Join(homeDir, "mirrors", "gitlab", project.Namespace.FullPath, project.Path)
	} else {
		namespacePath = filepath.Join(homeDir, "repos", "gitlab", project.Namespace.FullPath)
		projectPath = filepath.Join(homeDir, "repos", "gitlab", project.Namespace.FullPath, project.Path)
	}
	if err := os.MkdirAll(namespacePath, 0750); err != nil {
		fmt.Printf("  Error creating %s: %v\n", namespacePath, err)
		return
	}
	syncGitRepo(projectPath, project.SSHUrlToRepo, project.DefaultBranch)
}

func processGitLabGroup(baseURL, token, name string) {
	fmt.Printf("Processing projects in GitLab group %v\n", name)
	projects, err := FetchGitLabProjects(baseURL, token, name)
	if err != nil {
		fmt.Printf("  Error fetching projects for GitLab group %v: %v\n", name, err)
		return
	}
	processConcurrently(projects, getWorkerCount(), processGitLabProject)
}

func processGitLabGroups() {
	gltoken := os.Getenv("GLTOKEN")
	if gltoken == "" {
		fmt.Println("GLTOKEN is not set, skipping GitLab repositories")
		return
	}
	fmt.Println("GLTOKEN is set, processing GitLab repositories")
	baseURL := strings.TrimRight(os.Getenv("GLURL"), "/")
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	var glgroups []string
	if glgroup := os.Getenv("GLGROUP"); glgroup != "" {
		fmt.Println("GLGROUP is set, processing selected GitLab groups")
		glgroups = strings.Split(glgroup, ",")
	} else {
		fmt.Println("GLGROUP is not set, processing all GitLab groups")
		var err error
		glgroups, err = FetchGitLabGroups(baseURL, gltoken)
		if err != nil {
			fmt.Printf("Error fetching GitLab groups: %v\n", err)
			return
		}
	}
	processConcurrently(glgroups, getOrgWorkerCount(), func(group string) {
		processGitLabGroup(baseURL, gltoken, group)
	})
}

func main() {
	startTime := time.Now()
	fmt.Printf("Starting BitSync %s (%s) at %v\n", version, commit, startTime)
	processGitHubOrgs()
	processGitLabGroups()
	processBitBucketWorkspaces()
	endTime := time.Now()
	fmt.Printf("Finished BitSync process at %v\n", endTime)
	fmt.Printf("Elapsed time %v\n", endTime.Sub(startTime))
}
