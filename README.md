# ghadoc

### 🐙 🐈‍⬛ 📖

GitHub Actions Documentation Generator (`ghadoc`) is a CLI and
[pre-commit](https://pre-commit.com/) hook that automatically generates a
markdown table summarizing the GitHub Action workflows of a repository.

The resulting markdown looks like [this](example/workflows.md).

Ideally, `ghadoc` is incorporated into your pre-commit hooks so that the markdown
table can be updated any time your workflows change.

## Installation

```bash
go install github.com/droctothorpe/ghadoc@latest
```
### Usage

```bash
ghadoc generate -w example/workflows -o example/workflows.md
```


## Pre-commit hook setup

Update your `.pre-commit-config.yaml` file to include the following:

```yaml
repos:
  - repo: https://github.com/droctothorpe/ghadoc
    rev: 147a3312ae3d8e1baafa466c8f88153dfd3b0a6b
    hooks:
      - id: ghadoc
```