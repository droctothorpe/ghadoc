# gha-docs

### 🐙 🐈‍⬛ 📖

GitHub Actions Documentation Generator (`gha-docs`) is a CLI and
[pre-commit](https://pre-commit.com/) hook that automatically generates a
markdown table summarizing the GitHub Action workflows of a repository.

The resulting markdown looks like [this](example/workflows/_workflows.md).

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