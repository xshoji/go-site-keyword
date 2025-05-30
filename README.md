# go-site-keyword

> **Note**: This repository contains code that was largely generated with the assistance of GitHub Copilot (Claude 3.7 Sonnet and GPT-4.1).

A command-line tool that extracts and analyzes keywords from web pages. Additionally, this tool identifies keywords based on simple scoring and operates as a standalone application.

## Features

- Extract and analyze keywords from any web page
- Retrieve page titles and meta tags
- Calculate keyword relevance scores
- Display top keywords ranked by importance
- Support for both English and Japanese web pages with language-specific keyword extraction

## Installation

### Build from source

```
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath ./cmd/sitekeyword
```

### Using Go

```
go install github.com/xshoji/go-site-keyword/cmd/sitekeyword@latest
```

## Usage

Basic usage requires providing a URL to analyze:

```
sitekeyword -u https://example.com
```

or using the long option format:

```
sitekeyword --url https://example.com
```

### Available Options

- `-u, --url` (Required): The URL to analyze
- `-p, --pretty`: Format JSON output with indentation
- `-d, --detail`: Output all details including title and meta tags (By default, only keywords are displayed)

### Example output

By default, the tool outputs keywords in JSON format:

```json
{"keywords":[{"keyword":"example","score":15},{"keyword":"domain","score":12},{"keyword":"website","score":8}]}
```

With the pretty option:

```
sitekeyword -u https://example.com -p
```

```json
{
  "keywords": [
    {
      "keyword": "example",
      "score": 15
    },
    {
      "keyword": "domain",
      "score": 12
    },
    {
      "keyword": "website",
      "score": 8
    }
  ]
}
```

With the detail option:

```
sitekeyword -u https://example.com -p -d
```

```json
{
  "title": "Example Domain",
  "meta_tags": {
    "description": "This is an example website"
  },
  "keywords": [
    {
      "keyword": "example",
      "score": 15
    },
    {
      "keyword": "domain",
      "score": 12
    },
    {
      "keyword": "website",
      "score": 8
    }
  ]
}
```

## Important Considerations

When using this tool, please be aware of the following:

- Always respect the terms of service of the websites you analyze
- Do not use this tool to extract personal information or copyrighted content
- Be considerate of the website's server load by limiting request frequency

## Release

The release flow for this repository is automated with GitHub Actions.
Pushing Git tags triggers the release job.

```
# Release
git tag v0.0.2 && git push --tags


# Delete tag
echo "v0.0.1" |xargs -I{} bash -c "git tag -d {} && git push origin :{}"

# Delete tag and recreate new tag and push
echo "v0.0.2" |xargs -I{} bash -c "git tag -d {} && git push origin :{}; git tag {} -m \"Release beta version.\"; git push --tags"
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
