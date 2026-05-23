# font-util

A general-purpose font manipulation tool built on the [Glazed](https://github.com/go-go-golems/glazed) commands framework.

## Installation

```bash
go install github.com/go-go-golems/font-util/cmd/font-util@latest
```

## Commands

### ttc2ttf — Extract fonts from TTC collections

Extract individual TTF files from a TrueType Collection (.ttc) file.
For CFF-based collections (OTC), files are written with `.otf` extension.

```bash
# Extract all fonts from a TTC
font-util ttc2ttf fonts.ttc

# Extract to a specific directory
font-util ttc2ttf fonts.ttc --output-dir ./extracted

# Overwrite existing files
font-util ttc2ttf fonts.ttc --force

# List fonts without extracting
font-util ttc2ttf fonts.ttc --list
```

### inspect-ttc — List fonts in a collection

List all member fonts in a TTC file with index, name, type, and table count.

```bash
font-util inspect-ttc fonts.ttc
font-util inspect-ttc fonts.ttc --output json
font-util inspect-ttc fonts.ttc --fields index,name
```

### inspect-font — Inspect font metrics and shaping

Load a font and print its metrics (ascender, descender, x-height, cap height)
and optionally shape text examples.

```bash
# Basic metrics
font-util inspect-font ./font.otf

# With shaping examples
font-util inspect-font ./font.otf --text "AV,To,fi,office"

# Works with TTC files (use --font-index to select font)
font-util inspect-font fonts.ttc --font-index 2 --text "fi" --output json
```

### init-template — Create a practice sheet template

Create a starter YAML template for typography copy-practice sheets.

```bash
font-util init-template --font ./font.otf --out practice.yaml --pdf-out practice.pdf
```

### render — Render typography practice PDFs

Render a typography copy-practice PDF from a font and optional YAML template.

```bash
# From a template
font-util render --yaml-template practice.yaml

# Quick mode (no template needed)
font-util render --font ./font.otf --text "A,V,AV,To,fi,office" --blank-lines 3 --out practice.pdf

# Debug layout instead of creating a PDF
font-util render --yaml-template practice.yaml --dry-run

# Works with TTC files (use --font-index to select font)
font-util render --font fonts.ttc --font-index 1 --text "AV" --out av.pdf
```

## Output Formats

`inspect-ttc` and `inspect-font` support Glazed output formats:

```bash
font-util inspect-font font.otf --output table   # default
font-util inspect-font font.otf --output json
font-util inspect-font font.otf --output yaml
font-util inspect-font font.otf --output csv
```

## Build

```bash
make build
make test
make lint
```

## License

MIT
