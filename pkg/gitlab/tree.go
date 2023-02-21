package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
)

func (p *Provider) GetFolders() ([]string, error) {
	recurse := true
	headBranch := p.HeadBranch()
	nodes, _, err := p.client.Repositories.ListTree(p.RepositoryID(), &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
		Ref:       &headBranch,
		Recursive: &recurse,
	})

	if err != nil {
		fmt.Printf("Error getting tree: %s", err)
		return nil, err
	}

	folders := make([]string, 0)
	for _, node := range nodes {
		if node.Type == "tree" {
			folders = append(folders, node.Path)
		}
	}

	return folders, nil
}
