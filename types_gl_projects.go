package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// FetchGitLabProjects resolves name as a group first, falling back to a user
// namespace if the group endpoint returns 404. Returns a clean error only when
// neither exists. include_subgroups=true folds projects from nested groups
// into the same listing; their namespace.full_path preserves the hierarchy.
func FetchGitLabProjects(baseURL, apiToken, name string) (GitLabProjects, error) {
	encoded := url.PathEscape(name)
	groupURL := fmt.Sprintf("%s/api/v4/groups/%s/projects?per_page=100&include_subgroups=true", baseURL, encoded)
	projects, err := fetchGitLabProjectsFromURL(apiToken, groupURL)
	if !is404(err) {
		return projects, err
	}
	userURL := fmt.Sprintf("%s/api/v4/users/%s/projects?per_page=100", baseURL, encoded)
	projects, err = fetchGitLabProjectsFromURL(apiToken, userURL)
	if is404(err) {
		return nil, fmt.Errorf("no GitLab group or user named %q", name)
	}
	return projects, err
}

func fetchGitLabProjectsFromURL(apiToken, url string) (GitLabProjects, error) {
	var projects GitLabProjects
	for url != "" {
		page, next, err := fetchGitLabProjectsPage(apiToken, url)
		if err != nil {
			return nil, err
		}
		projects = append(projects, page...)
		url = next
	}
	return projects, nil
}

func fetchGitLabProjectsPage(apiToken, url string) (GitLabProjects, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("PRIVATE-TOKEN", apiToken)
	req.Header.Set("Accept", "application/json")
	body, next, err := fetchAPI(req)
	if err != nil {
		return nil, "", err
	}
	var page GitLabProjects
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, "", err
	}
	return page, next, nil
}

type GitLabProjects []GitLabProject

type GitLabProject struct {
	Id                int64           `json:"id"`
	Name              string          `json:"name"`
	Path              string          `json:"path"`
	PathWithNamespace string          `json:"path_with_namespace"`
	DefaultBranch     string          `json:"default_branch"`
	SSHUrlToRepo      string          `json:"ssh_url_to_repo"`
	HttpUrlToRepo     string          `json:"http_url_to_repo"`
	WebURL            string          `json:"web_url"`
	Description       string          `json:"description"`
	Archived          bool            `json:"archived"`
	Visibility        string          `json:"visibility"`
	CreatedAt         time.Time       `json:"created_at"`
	LastActivityAt    time.Time       `json:"last_activity_at"`
	Namespace         GitLabNamespace `json:"namespace"`
}

type GitLabNamespace struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
	Kind     string `json:"kind"`
	WebURL   string `json:"web_url"`
}
