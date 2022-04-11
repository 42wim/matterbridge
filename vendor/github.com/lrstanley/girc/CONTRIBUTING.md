<!-- THIS FILE IS GENERATED! DO NOT EDIT! Maintained by Terraform. -->
# :handshake: Contributing

This document outlines some of the guidelines that we try and adhere to while
working on this project.

> :point_right: **Note**: before participating in the community, please read our
> [Code of Conduct][coc].
> By interacting with this repository, organization, or community you agree to
> abide by our Code of Conduct.
>
> Additionally, if you contribute **any source code** to this repository, you
> agree to the terms of the [Developer Certificate of Origin][dco]. This helps
> ensure that contributions aren't in violation of 3rd party license terms.

## :lady_beetle: Issue submission

When [submitting an issue][issues] or bug report,
please follow these guidelines:

   * Provide as much information as possible (logs, metrics, screenshots,
     runtime environment, etc).
   * Ensure that you are running on the latest stable version (tagged), or
     when using `master`, provide the specific commit being used.
   * Provide the minimum needed viable source to replicate the problem.

## :bulb: Feature requests

When [submitting a feature request][issues], please
follow these guidelines:

   * Does this feature benefit others? or just your usecase? If the latter,
     it will likely be declined, unless it has a more broad benefit to others.
   * Please include the pros and cons of the feature.
   * If possible, describe how the feature would work, and any diagrams/mock
     examples of what the feature would look like.

## :rocket: Pull requests

To review what is currently being worked on, or looked into, feel free to head
over to the [open pull requests][pull-requests] or [issues list][issues].

## :raised_back_of_hand: Assistance with discussions

   * Take a look at the [open discussions][discussions], and if you feel like
     you'd like to help out other members of the community, it would be much
     appreciated!

## :pushpin: Guidelines

### :test_tube: Language agnostic

Below are a few guidelines if you would like to contribute:

   * If the feature is large or the bugfix has potential breaking changes,
     please open an issue first to ensure the changes go down the best path.
   * If possible, break the changes into smaller PRs. Pull requests should be
     focused on a specific feature/fix.
   * Pull requests will only be accepted with sufficient documentation
     describing the new functionality/fixes.
   * Keep the code simple where possible. Code that is smaller/more compact
     does not mean better. Don't do magic behind the scenes.
   * Use the same formatting/styling/structure as existing code.
   * Follow idioms and community-best-practices of the related language,
     unless the previous above guidelines override what the community
     recommends.
   * Always test your changes, both the features/fixes being implemented, but
     also in the standard way that a user would use the project (not just
     your configuration that fixes your issue).
   * Only use 3rd party libraries when necessary. If only a small portion of
     the library is needed, simply rewrite it within the library to prevent
     useless imports.

### :hamster: Golang

   * See [golang/go/wiki/CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)
   * This project uses [golangci-lint](https://golangci-lint.run/) for
     Go-related files. This should be available for any editor that supports
     `gopls`, however you can also run it locally with `golangci-lint run`
     after installing it.




## :clipboard: References

   * [Open Source: How to Contribute](https://opensource.guide/how-to-contribute/)
   * [About pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests)
   * [GitHub Docs](https://docs.github.com/)

## :speech_balloon: What to do next?

   * :old_key: Find a vulnerability? Check out our [Security and Disclosure][security] policy.
   * :link: Repository [License][license].
   * [Support][support]
   * [Code of Conduct][coc].

<!-- definitions -->
[coc]: https://github.com/lrstanley/girc/blob/master/CODE_OF_CONDUCT.md
[dco]: https://developercertificate.org/
[discussions]: https://github.com/lrstanley/girc/discussions
[issues]: https://github.com/lrstanley/girc/issues/new/choose
[license]: https://github.com/lrstanley/girc/blob/master/LICENSE
[pull-requests]: https://github.com/lrstanley/girc/issues/new/choose
[security]: https://github.com/lrstanley/girc/security/policy
[support]: https://github.com/lrstanley/girc/blob/master/SUPPORT.md
