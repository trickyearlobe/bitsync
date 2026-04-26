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

func main() {
	startTime := time.Now()
	fmt.Printf("Starting BitSync process at %v\n", startTime)
	processGitHubOrgs()
	processBitBucketWorkspaces()
	endTime := time.Now()
	fmt.Printf("Finished BitSync process at %v\n", endTime)
	fmt.Printf("Elapsed time %v\n", endTime.Sub(startTime))
}
