# gomatrix
[![GoDoc](https://godoc.org/github.com/matrix-org/gomatrix?status.svg)](https://godoc.org/github.com/matrix-org/gomatrix)

A Golang Matrix client.

**THIS IS UNDER ACTIVE DEVELOPMENT: BREAKING CHANGES ARE FREQUENT.**

# Contributing

All contributions are greatly appreciated!

## How to report issues

Please check the current open issues for similar reports
in order to avoid duplicates.

Some general guidelines:

-   Include a [minimal reproducible example](https://stackoverflow.com/help/minimal-reproducible-example) when possible.
-   Describe the expected behaviour and what actually happened
    including a full trace-back in case of exceptions.
-   Make sure to list details about your environment

## Setting up your environment

If you intend to contribute to gomatrix you'll first need Go installed on your machine (version 1.12+ is required). Also, make sure to have golangci-lint properly set up since we use it for pre-commit hooks (for instructions on how to install it, check the [official docs](https://golangci-lint.run/usage/install/#local-installation)).

-   Fork gomatrix to your GitHub account by clicking the [Fork](https://github.com/matrix-org/gomatrix/fork) button.
-   [Clone](https://help.github.com/en/articles/fork-a-repo#step-2-create-a-local-clone-of-your-fork) the main repository (not your fork) to your local machine.

        
        $ git clone https://github.com/matrix-org/gomatrix
        $ cd gomatrix
        

-   Add your fork as a remote to push your contributions.Replace
    ``{username}`` with your username.

        git remote add fork https://github.com/{username}/gomatrix

-   Create a new branch to identify what feature you are working on.

        $ git fetch origin
        $ git checkout -b your-branch-name origin/master
        

-   Make your changes, including tests that cover any code changes you make, and run them as described below.

-   Execute pre-commit hooks by running 

        <gomatrix dir>/hooks/pre-commit

-   Push your changes to your fork and [create a pull request](https://help.github.com/en/articles/creating-a-pull-request) describing your changes.

        $ git push --set-upstream fork your-branch-name

-   Finally, create a [pull request](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests)

## How to run tests

You can run the test suite and example code with `$ go test -v`

# Running Coverage

To run coverage, first generate the coverage report using `go test`

    go test -v -cover -coverprofile=coverage.out

You can now show the generated report as a html page with `go tool`

    go tool cover -html=coverage.out
