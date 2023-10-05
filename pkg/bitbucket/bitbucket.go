package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	configuration "github.com/diggerhq/lib-digger-config"
	orchestrator "github.com/diggerhq/lib-orchestrator"
	"net/http"
	"time"
)

// Define the base URL for the Bitbucket API.
const bitbucketBaseURL = "https://api.bitbucket.org/2.0"

// BitbucketAPI is a struct that holds the required authentication information.
type BitbucketAPI struct {
	AuthToken     string
	HttpClient    http.Client
	RepoWorkspace string
	RepoName      string
}

func (b *BitbucketAPI) sendRequest(method, url string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.AuthToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// {"pagelen":500,"values":[{"type":"diffstat","status":"modified","lines_removed":1,"lines_added":2,"old":{"path":"setup.py","escaped_path":"setup.py","type":"commit_file","links":{"self":{"href":"https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/e1749643d655d7c7014001a6c0f58abaf42ad850/setup.py"}}},"new":{"path":"setup.py","escaped_path":"setup.py","type":"commit_file","links":{"self":{"href":"https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/d222fa235229c55dad20b190b0b571adf737d5a6/setup.py"}}}}],"page":1,"size":1}
type DiffStat struct {
	Pagelen int `json:"pagelen"`
	Values  []struct {
		Type         string `json:"type"`
		Status       string `json:"status"`
		LinesRemoved int    `json:"lines_removed"`
		LinesAdded   int    `json:"lines_added"`
		Old          struct {
			Path        string `json:"path"`
			EscapedPath string `json:"escaped_path"`
			Type        string `json:"type"`
			Links       struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"old"`
		New struct {
			Path        string `json:"path"`
			EscapedPath string `json:"escaped_path"`
			Type        string `json:"type"`
			Links       struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"new"`
	} `json:"values"`
	Page int `json:"page"`
	Size int `json:"size"`
}

func (b *BitbucketAPI) GetChangedFiles(prNumber int) ([]string, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/diffstat", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get changed files. Status code: %d", resp.StatusCode)
	}

	diffStat := &DiffStat{}
	err = json.NewDecoder(resp.Body).Decode(diffStat)
	if err != nil {
		return nil, err
	}
	var files []string

	for _, v := range diffStat.Values {
		files = append(files, v.Old.Path)
	}
	return files, nil
}

func (b *BitbucketAPI) PublishComment(prNumber int, comment string) error {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/comments", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	commentBody := map[string]interface{}{
		"content": map[string]string{
			"raw": comment,
		},
	}

	commentJSON, err := json.Marshal(commentBody)
	if err != nil {
		return err
	}

	resp, err := b.sendRequest("POST", url, commentJSON)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to publish comment. Status code: %d", resp.StatusCode)
	}

	return nil
}

func (b *BitbucketAPI) EditComment(prNumber int, id interface{}, comment string) error {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/comments/%s", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber, id)

	commentBody := map[string]string{
		"content": comment,
	}

	commentJSON, err := json.Marshal(commentBody)
	if err != nil {
		return err
	}

	resp, err := b.sendRequest("PUT", url, commentJSON)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to edit comment. Status code: %d", resp.StatusCode)
	}

	return nil
}

//	GetComments(prNumber int) ([]Comment, error)
//	// SetStatus set status of specified pull/merge request, status could be: "pending", "failure", "success"
//	SetStatus(prNumber int, status string, statusContext string) error
//	GetCombinedPullRequestStatus(prNumber int) (string, error)
//	MergePullRequest(prNumber int) error
//	// IsMergeable is still open and ready to be merged
//	IsMergeable(prNumber int) (bool, error)
//	// IsMerged merged and closed
//	IsMerged(prNumber int) (bool, error)
//	// IsClosed closed without merging
//	IsClosed(prNumber int) (bool, error)
//	GetBranchName(prNumber int) (string, error)

type Comment struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Values   []struct {
		Id      int `json:"id"`
		Content struct {
			Raw string `json:"raw"`
		}
	} `json:"values"`
}

func (b *BitbucketAPI) GetComments(prNumber int) ([]orchestrator.Comment, error) {

	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/comments", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get comments. Status code: %d", resp.StatusCode)
	}

	var commentResponse Comment
	err = json.NewDecoder(resp.Body).Decode(&commentResponse)
	if err != nil {
		return nil, err
	}

	var comments []orchestrator.Comment

	for _, v := range commentResponse.Values {
		comments = append(comments, orchestrator.Comment{
			Id:   v.Id,
			Body: &v.Content.Raw,
		})
	}

	return comments, nil

}

func (b *BitbucketAPI) SetStatus(prNumber int, status string, statusContext string) error {
	url := fmt.Sprintf("%s/repositories/%s/%s/commit/%d/statuses/build", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	statusBody := map[string]interface{}{
		"state": status,
		"key":   statusContext,
	}

	statusJSON, err := json.Marshal(statusBody)
	if err != nil {
		return err
	}

	resp, err := b.sendRequest("POST", url, statusJSON)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to set status. Status code: %d", resp.StatusCode)
	}

	return nil
}

type CommitStatuses struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	Pagelen  int    `json:"pagelen"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Values   []struct {
		State     string    `json:"state"`
		Key       string    `json:"key"`
		UpdatedOn time.Time `json:"updated_on"`
	} `json:"values"`
}

func (b *BitbucketAPI) GetCombinedPullRequestStatus(prNumber int) (string, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/commit/%d/statuses", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get combined status. Status code: %d", resp.StatusCode)
	}

	var statuses CommitStatuses

	err = json.NewDecoder(resp.Body).Decode(&statuses)

	if err != nil {
		return "", err
	}
	// group by key and get latest per key

	type status struct {
		State     string
		UpdatedOn time.Time
	}
	latestStatusByKey := make(map[string]status)

	for _, v := range statuses.Values {
		currentlyKnownStatus, ok := latestStatusByKey[v.Key]
		if !ok {
			latestStatusByKey[v.Key] = status{
				State:     v.State,
				UpdatedOn: v.UpdatedOn,
			}
			continue
		}
		if currentlyKnownStatus.UpdatedOn.Before(v.UpdatedOn) {
			latestStatusByKey[v.Key] = status{
				State:     v.State,
				UpdatedOn: v.UpdatedOn,
			}
		}
	}
	for _, status := range latestStatusByKey {
		if status.State == "FAILED" {
			return "failure", nil
		}
	}

	var allSuccess = true
	for _, status := range latestStatusByKey {
		if status.State != "SUCCESSFUL" {
			allSuccess = false
			break
		}
	}
	if allSuccess {
		return "success", nil
	}

	return "pending", nil

}

func (b *BitbucketAPI) MergePullRequest(prNumber int) error {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d/merge", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("POST", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to merge pull request. Status code: %d", resp.StatusCode)
	}

	return nil
}

func (b *BitbucketAPI) IsMergeable(prNumber int) (bool, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get pull request. Status code: %d", resp.StatusCode)
	}

	var pullRequest struct {
		State string `json:"state"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pullRequest)
	if err != nil {
		return false, err
	}

	return pullRequest.State == "OPEN", nil
}

func (b *BitbucketAPI) IsMerged(prNumber int) (bool, error) {

	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get pull request. Status code: %d", resp.StatusCode)
	}

	var pullRequest struct {
		State string `json:"state"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pullRequest)
	if err != nil {
		return false, err
	}

	return pullRequest.State == "MERGED", nil
}

func (b *BitbucketAPI) IsClosed(prNumber int) (bool, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get pull request. Status code: %d", resp.StatusCode)
	}

	var pullRequest struct {
		State string `json:"state"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pullRequest)
	if err != nil {
		return false, err
	}

	return pullRequest.State == "DECLINED", nil
}

func (b *BitbucketAPI) GetBranchName(prNumber int) (string, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests/%d", bitbucketBaseURL, b.RepoWorkspace, b.RepoName, prNumber)

	resp, err := b.sendRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get pull request. Status code: %d", resp.StatusCode)
	}

	var pullRequest struct {
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pullRequest)
	if err != nil {
		return "", err
	}

	return pullRequest.Source.Branch.Name, nil
}

// Implement the OrgService interface.

func (b *BitbucketAPI) GetUserTeams(organisation string, user string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func FindImpactedProjectsInBitbucket(diggerConfig *configuration.DiggerConfig, prNumber int, prService orchestrator.PullRequestService) ([]configuration.Project, error) {
	changedFiles, err := prService.GetChangedFiles(prNumber)

	if err != nil {
		fmt.Printf("Error getting changed files: %v", err)
		return nil, err
	}

	impactedProjects := diggerConfig.GetModifiedProjects(changedFiles)
	return impactedProjects, nil
}