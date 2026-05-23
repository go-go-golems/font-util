# font-util

A general-purpose font manipulation tool built on the [Glazed](https://github.com/go-go-golems/glazed) commands framework.

## Installation

```bash
go install github.com/go-go-golems/font-util/cmd/font-util@latest
```

## Commands

### ttc2ttf — Extract fonts from TTC collections

Extract individual TTF files from a TrueType Collection (.ttc) file.

```bash
# Extract all fonts from a TTC into the current directory
font-util ttc2ttf fonts.ttc

# Extract to a specific directory
font-util ttc2ttf fonts.ttc --output-dir ./extracted

# Overwrite existing files
font-util ttc2ttf fonts.ttc --force

# JSON output of extracted font names
font-util ttc2ttf fonts.ttc --output json
```

Output files are named using the PostScript name (Name ID 6) from each font's `name` table, e.g., `GillSans-Bold.ttf`, `Futura-Medium.ttf`.

### init-template — Create a practice sheet template

Create a starter YAML template for typography copy-practice sheets.

```bash
font-util init-template --font ./font.otf --out practice.yaml --pdf-out practice.pdf
```

### inspect-font — Inspect font metrics and shaping

Load a font and print its metrics (ascender, descender, x-height, cap height) and optionally shape text examples.

```bash
# Basic metrics
font-util inspect-font ./font.otf

# With shaping examples
font-util inspect-font ./font.otf --text "AV,To,fi,office"

# Works with TTC files (inspects the first font)
font-util inspect-font fonts.ttc --text "fi" --output json
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

# Works with TTC files (uses the first font)
font-util render --font fonts.ttc --text "AV" --out av.pdf
```

The PDF includes model rows (font drawn as vector outlines) and blank practice rows with baseline, x-height, and cap-height helper lines.

## Output Formats

All commands support Glazed output formats:

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
