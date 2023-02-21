package provider

import (
	"github.com/phroggyy/decision/pkg/git"
	"github.com/phroggyy/decision/pkg/github"
	"github.com/phroggyy/decision/pkg/gitlab"
)

var provider Provider

type Provider interface {
	// RaisePullRequest will automatically create a commit, create a branch, and open a pull request, and return the URL to the PR.
	RaisePullRequest(branch string, commitMessage string, path string, content []byte) (string, error)

	// CreateCommit creates a commit with the given content
	CreateCommit(commitMessage string, path string, content []byte) (string, error)

	// GetFolders returns all available folders in the configured repository
	GetFolders() ([]string, error)
}

func GetProvider() Provider {
	if provider != nil {
		return provider
	}

	provider = GetProviderForType(git.ProviderType, git.Token)

	return provider
}

func GetProviderForType(providerType string, token string) Provider {
	switch providerType {
	case "github":
		return github.NewProvider(token)
	case "gitlab":
		return gitlab.NewProvider(token)
	}

	return nil
}
