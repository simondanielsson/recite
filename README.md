<p align="center">
  <img src="images/logo.png" alt="recite logo" width="200"/>
</p>

# recite

[![Go Lint](https://github.com/simondanielsson/recite/actions/workflows/lint.yml/badge.svg)](https://github.com/simondanielsson/recite/actions/workflows/lint.yml)
[![Unit Tests](https://github.com/simondanielsson/recite/actions/workflows/test.yml/badge.svg)](https://github.com/simondanielsson/recite/actions/workflows/test.yml)

Listen to your favorite articles as if it was an audiobook.

This is a simple and accessible wrapper around OpenAI's TTS models.

## How to run

The app comes in two version: as a CLI and as an API.

To build and launch the CLI version of `recite`, run

```bash
make build
./cli "https://example-article.com/index"
```

To run the API version, run

```bash
make api # or ./api after building
```

## Contributing

Always verify your code using

```bash
make ci
```

