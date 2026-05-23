# font-util

A general-purpose font manipulation tool built on the [Glazed](https://github.com/go-go-golems/glazed) commands framework.

## Installation

```bash
go install github.com/go-go-golems/font-util/cmd/font-util@latest
```

## Commands

### ttc2ttf

Extract individual TTF files from a TrueType Collection (.ttc) file.

```bash
# Extract all fonts from a TTC into the current directory
font-util ttc2ttf fonts.ttc

# Extract to a specific directory
font-util ttc2ttf fonts.ttc --output-dir ./extracted

# Overwrite existing files
font-util ttc2ttf fonts.ttc --force

# Get JSON output of extracted font names
font-util ttc2ttf fonts.ttc --output json
```

Output files are named using the PostScript name (Name ID 6) from each font's `name` table, e.g., `GillSans-Bold.ttf`, `Futura-Medium.ttf`.

## Output Formats

All commands support Glazed output formats:

```bash
font-util ttc2ttf fonts.ttc --output table   # default
font-util ttc2ttf fonts.ttc --output json
font-util ttc2ttf fonts.ttc --output yaml
font-util ttc2ttf fonts.ttc --output csv
font-util ttc2ttf fonts.ttc --output markdown
```

## Build

```bash
make build
make test
make lint
```

## License

MIT
