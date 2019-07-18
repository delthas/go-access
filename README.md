# go-access [![builds.sr.ht status](https://builds.sr.ht/~delthas/go-access.svg)](https://builds.sr.ht/~delthas/go-access?) [![GoDoc](https://godoc.org/github.com/delthas/go-access?status.svg)](https://godoc.org/github.com/delthas/go-access) [![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

A small Go library for checking whether a user has the permissions to access a file (on *nix).

- checks mode on file and `x` bit on all its parent directories
- checks for permission on all, group, and user modes
- checks for permission on symlinks and resolves them
- the current user needs to have access to the file
- does not check permissions for root but makes all stat system calls regardless
- does not support ACLs

## status

- api stability: [![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental) open issues or PRs if you need anything new for your use case
- bugs: tested locally, probably needs more testing

## using

- import access "github.com/delthas/go-access"

## docs  [![GoDoc](https://godoc.org/github.com/delthas/go-access?status.svg)](https://godoc.org/github.com/delthas/go-access)

- [example](https://github.com/delthas/go-access/blob/master/access_test.go)
- [godoc](https://godoc.org/github.com/delthas/go-access)

## license

MIT
