package gitlab

import (
	"fmt"
	"github.com/phroggyy/decision/pkg/git"
	"github.com/xanzy/go-gitlab"
)

func (p *Provider) CreateCommit(commitMessage string, path string, content []byte) (string, error) {
	return p.createCommitOnBranch(commitMessage, path, string(content), p.HeadBranch())
}

func (p *Provider) createCommitOnBranch(commitMessage, path, content, branch string) (string, error) {
	createAction := gitlab.FileCreate
	headBranch := p.HeadBranch()
	commit, _, err := p.client.Commits.CreateCommit(
		p.RepositoryID(),
		&gitlab.CreateCommitOptions{
			Branch:        &branch,
			StartBranch:   &headBranch,
			CommitMessage: &commitMessage,
			Actions: []*gitlab.CommitActionOptions{
				{
					Action:   &createAction,
					FilePath: &path,
					Content:  &content,
				},
			},
			AuthorEmail: &git.AuthorEmail,
			AuthorName:  &git.AuthorName,
		},
	)

	if err != nil {
		fmt.Printf("Error creating commit: %s", err)
		return "", err
	}

	return commit.WebURL, nil
}
