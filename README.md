# recite

[![Go Lint](https://github.com/simondanielsson/recite/actions/workflows/lint.yml/badge.svg)](https://github.com/simondanielsson/recite/actions/workflows/lint.yml)
[![Unit Tests](https://github.com/simondanielsson/recite/actions/workflows/test.yml/badge.svg)](https://github.com/simondanielsson/recite/actions/workflows/test.yml)

Listen to your favorite articles as if it was an audiobook.

This is a simple and accessible wrapper around OpenAI's TTS models.

## How to run

The app comes in two version: as a CLI and as an API.

To build and launch the CLI version of `recite`, run

```bash
make cli
```

To run the API version, run

```bash
make api
```

## Contributing

Always verify your code using

```bash
make ci
```

