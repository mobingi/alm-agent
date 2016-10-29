package code

import git "gopkg.in/libgit2/git2go.v24"

type Git struct {
	url  string
	path string
	ref  string
}

func (g *Git) get() error {
	options := &git.CloneOptions{CheckoutBranch: g.ref}
	_, err := git.Clone(g.url, g.path, options)
	return err
}
