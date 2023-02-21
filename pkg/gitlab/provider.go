package gitlab

import (
	"fmt"
	"github.com/phroggyy/decision/pkg/git"
	"github.com/xanzy/go-gitlab"
)

type Provider struct {
	client      *gitlab.Client
	sourceOwner string
	sourceRepo  string
}

func NewProvider(accessToken string) *Provider {
	c, err := gitlab.NewClient(accessToken)
	if err != nil {
		panic(err)
	}

	return &Provider{
		client: c,
	}
}

func (p *Provider) SetRepository(owner, repository string) {
	p.sourceOwner = owner
	p.sourceRepo = repository
}

func (p *Provider) RepositoryID() string {
	if p.sourceOwner == "" || p.sourceRepo == "" {
		return fmt.Sprintf("%s/%s", git.SourceOwner, git.SourceRepo)
	}

	return fmt.Sprintf("%s/%s", p.sourceOwner, p.sourceRepo)
}

func (p *Provider) HeadBranch() string {
	if git.CommitHeadBranch == "" {
		return "main"
	}

	return git.CommitHeadBranch
}
