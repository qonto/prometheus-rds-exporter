# Contributing

Prometheus RDS exporter uses GitHub to manage reviews of pull requests.

* If you are a new contributor, see: [Steps to Contribute](#steps-to-contribute)

* If you have a trivial fix or improvement, go ahead and create a pull request

* Relevant coding style guidelines are the [Go Code Review
  Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best
  Practices for Production
  Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

* Be sure to enable [signed commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on it. This is to prevent duplicated efforts from contributors on the same issue.

All our issues are regularly tagged so you can filter down the issues involving the components you want to work on.

For quickly compiling and testing your changes do:

```bash
# For building.
make build
./rds_exporter

# For testing.
make test         # Make sure all the tests pass before you commit and push :)
```

We use:

* [`pre-commit`](https://pre-commit.com) to make right first time changes. Enable it for this repository with `pre-commit install`.

* [`golangci-lint`](https://github.com/golangci/golangci-lint) for linting the code. If it reports an issue and you think that the warning needs to be disregarded or is a false-positive, you can add a special comment `//nolint:linter1[,linter2,...]` before the offending line. Use this sparingly though, fixing the code to comply with the linter's recommendation is in general the preferred course of action.

* [`markdownlint-cli2`](https://github.com/DavidAnson/markdownlint-cli2) for linting the Markdown documents.

* [`yamllint`](https://github.com/adrienverge/yamllint) for linting the YAML documents.

## Pull Request Checklist

* Branch from the `main` branch and, if needed, rebase to the current main branch before submitting your pull request. If it doesn't merge cleanly with main you may be asked to rebase your changes.

* Commits should be as small as possible while ensuring each commit is correct independently (i.e., each commit should compile and pass tests).

* Add tests relevant to the fixed bug or new feature.

* New Prometheus metrics must follow the [metric and label naming guidelines](https://prometheus.io/docs/practices/naming/) and be added in the `README.md`.

## Dependency management

Project uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies on external packages.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go get example.com/some/module/pkg@latest

# Pick a specific version.
go get example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files:

```bash
# The GO111MODULE variable can be omitted when the code isn't located in GOPATH.
GO111MODULE=on go mod tidy
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the pull request.

## Install pre-commit

1. Install [pre-commit](https://pre-commit.com/)

1. Install [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2)

1. Enable pre-commit for the repository

    ```bash
    pre-commit install
    ```

## Update dashboard

1. Start local development environment

1. Update dashboard

1. Export the dashboard

    Click on `Share dashboard > Export`, then select `Export for sharing externally` and click on `View JSON`.

1. Open an issue with the suggestion

1. The project manager will review it and submit it to Grafana.com.
