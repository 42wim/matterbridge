# Contributing

## Issue submission

   * When submitting an issue or bug report, please ensure to provide as much
     information as possible, please ensure that you are running on the latest
     stable version (tagged), or when using master, provide the specific commit
     being used.
   * Provide the minimum needed viable source to replicate the problem.

## Pull requests

To review what is currently being worked on, or looked into, feel free to head
over to the [issues list](../../issues).

Below are a few guidelines if you would like to contribute. Keep the code
clean, standardized, and much of the quality should match Golang's standard
library and common idioms.

   * Always test using the latest Go version.
   * Always use `gofmt` before committing anything.
   * Always have proper documentation before committing.
   * Keep the same whitespacing, documentation, and newline format as the
     rest of the project.
   * Only use 3rd party libraries if necessary. If only a small portion of
     the library is needed, simply rewrite it within the library to prevent
     useless imports.
   * Also see [golang/go/wiki/CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

If you would like to assist, and the pull request is quite large and/or it has
the potential of being a breaking change, please open an issue first so it can
be discussed.
