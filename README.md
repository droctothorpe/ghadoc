# gha-docs

### 🐙 🐈‍⬛ 📖

GitHub Actions Documentation Generator (`gha-docs`) is a CLI and
[pre-commit](https://pre-commit.com/) hook that automatically generates a
markdown table summarizing the GitHub Actions workflows of a repository.

The resulting markdown table looks like this:

---

# GitHub Workflows Summary

| Filename | Description | Triggers |
| --- | --- | --- |
| [add-ci-passed-label.yml](example/workflows/add-ci-passed-label.yml) | Adds the 'ci-passed' label to a pull request once the 'CI Check' workflow completes successfully. | workflow_run |
| [api-server-tests.yml](example/workflows/api-server-tests.yml) | Runs integration tests against API. | pull_request, push, workflow_dispatch |
| [backend-visualization.yml](example/workflows/backend-visualization.yml) | Runs unit tests against backend visualization server. | pull_request, push |
| [build-and-push.yml](example/workflows/build-and-push.yml) | Builds and pushes images to GitHub Container Registry. | workflow_call, workflow_dispatch |
| [e2e-tests.yml](example/workflows/e2e-tests.yml) | Runs end-to-end tests against the backend. | pull_request, push |
| [unit-tests.yml](example/workflows/unit-tests.yml) | Runs unit tests against the backend. | pull_request, push |

---

Ideally, `gha-docs` is incorporated into your pre-commit hooks so that the markdown
table can be updated any time your workflows change.

## Installation

```bash
go install github.com/droctothorpe/gha-doc@latest
```
## Usage

```bash
gha-docs generate -w example/workflows -o example/workflows.md
```


## Pre-commit hook setup

Update your `.pre-commit-config.yaml` file to include the following:

```yaml
repos:
  - repo: https://github.com/droctothorpe/gha-docs
    rev: 3d45eedd95fe9a417f03b58ea350fe6d90d6c3bf
    hooks:
      - id: gha-docs
```

## Populate descriptions

At the top of each GitHub workflow file, add one or more comment lines that begin with
`##`. These will be extracted to populate the `Description` column of the
markdown table.