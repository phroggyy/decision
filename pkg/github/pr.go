package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/phroggyy/decision/pkg/git"
	"log"
)

func (p *Provider) RaisePullRequest(branch string, commitMessage string, path string, content []byte) (string, error) {
	ref, err := p.getBranchRef(branch)
	if err != nil {
		log.Printf("Unable to get/create the commit reference: %s\n", err)
		return "", err
	}
	if ref == nil {
		log.Printf("No error where returned but the reference is nil")
		return "", errors.New("no reference could be retrieved")
	}

	if _, err = p.createFileOnBranch(commitMessage, path, branch, content); err != nil {
		return "", err
	}

	prURL, err := p.createPR(commitMessage, branch)
	if err != nil {
		log.Printf("Error while creating the pull request: %s", err)
		return "", err
	}

	return prURL, nil
}

func (p *Provider) createPR(prSubject string, branch string) (url string, err error) {
	prBody := git.PullRequestBody(prSubject)
	headBranch := p.HeadBranch()

	newPR := &github.NewPullRequest{
		Title:               &prSubject,
		Head:                &branch,
		Base:                &headBranch,
		Body:                &prBody,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := p.client.PullRequests.Create(context.Background(), p.GetOwner(), p.GetRepository(), newPR)
	if err != nil {
		return "", err
	}

	prURL := pr.GetHTMLURL()
	fmt.Printf("PR created: %s\n", prURL)

	return prURL, nil
}

// getRef returns the commit branch reference object if it exists or creates it
// from the base branch before returning it.
func (p *Provider) getBranchRef(commitBranch string) (ref *github.Reference, err error) {
	ctx := context.Background()
	if ref, _, err = p.client.Git.GetRef(ctx, p.GetOwner(), p.GetRepository(), "refs/heads/"+commitBranch); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = p.client.Git.GetRef(ctx, p.GetOwner(), p.GetRepository(), "refs/heads/"+p.HeadBranch()); err != nil {
		return nil, err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = p.client.Git.CreateRef(ctx, p.GetOwner(), p.GetRepository(), newRef)
	return ref, err
}
