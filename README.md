# Gitrelease

[![Continues Integration](https://github.com/arsham/gitrelease/actions/workflows/go.yml/badge.svg)](https://github.com/arsham/gitrelease/actions/workflows/go.yml)
![License](https://img.shields.io/github/license/arsham/gitrelease)

This program can set the release information based on all commits of a tag. To
see the example visit [Releases](https://github.com/arsham/gitrelease/releases)
page.

1. [Requirements](#requirements)
2. [Installation](#installation)
3. [Usage](#usage)
4. [License](#license)

## Requirements

This program requires `Go >= v1.17`.

Uses your github token with permission scope: **repo**

## Installation

To install:

```bash
go install github.com/arsham/gitrelease@latest
```

Export your github token:
`export GITHUB_TOKEN="ghp_yourgithubtoken"`

## Usage

After you've made a tag, you can publish the current release documents by just
running:

```bash
gitrelease
```

If you want to release an old tag:

```bash
gitrelease -t v0.1.2
```

If you want to use a different remote other than the `origin`:

```bash
gitrelease -r upstream
```

## License

Licensed under the MIT License. Check the [LICENSE](./LICENSE) file for details.

<!--
vim: foldlevel=1
-->
