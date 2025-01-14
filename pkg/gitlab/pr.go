package gitlab

import (
	"fmt"
	"github.com/phroggyy/decision/pkg/git"
	"github.com/xanzy/go-gitlab"
)

func (p *Provider) RaisePullRequest(branch string, commitMessage string, path string, content []byte) (string, error) {
	_, err := p.createCommitOnBranch(commitMessage, path, string(content), branch)

	if err != nil {
		return "", err
	}

	removeBranch := true
	squash := true
	description := git.PullRequestBody(commitMessage)
	headBranch := p.HeadBranch()
	mr, _, err := p.client.MergeRequests.CreateMergeRequest(p.RepositoryID(), &gitlab.CreateMergeRequestOptions{
		Title:              &commitMessage,
		Description:        &description,
		SourceBranch:       &branch,
		TargetBranch:       &headBranch,
		RemoveSourceBranch: &removeBranch,
		Squash:             &squash,
	})

	if err != nil {
		fmt.Printf("Error opening MR: %s", err)
		return "", err
	}

	return mr.WebURL, nil
}
