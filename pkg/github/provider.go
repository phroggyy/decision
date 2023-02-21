package github

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/phroggyy/decision/pkg/git"
	"golang.org/x/oauth2"
)

type Provider struct {
	client      *github.Client
	sourceOwner string
	sourceRepo  string
}

func NewProvider(accessToken string) *Provider {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(context.Background(), ts)

	return &Provider{
		client: github.NewClient(tc),
	}
}

func (p *Provider) SetRepository(owner, repository string) {
	p.sourceOwner = owner
	p.sourceRepo = repository
}

func (p *Provider) GetOwner() string {
	if p.sourceOwner == "" {
		return git.SourceOwner
	}

	return p.sourceOwner
}

func (p *Provider) GetRepository() string {
	if p.sourceRepo == "" {
		return git.SourceRepo
	}

	return p.sourceRepo
}
