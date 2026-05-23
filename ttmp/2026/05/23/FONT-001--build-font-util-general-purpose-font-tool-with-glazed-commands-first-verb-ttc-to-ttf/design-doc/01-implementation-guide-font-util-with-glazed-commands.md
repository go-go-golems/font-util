---
type: design-doc
title: Implementation Guide: font-util with Glazed Commands
status: active
intent: long-term
topics:
  - fonts
  - glazed
  - cli
  - go
  - binary-parsing
created: 2026-05-23
owners: []
ticket: FONT-001
---

# Implementation Guide: font-util — A General-Purpose Font Tool Built on Glazed

## Executive Summary

`font-util` is a command-line tool for font file manipulation, built using the go-go-golems **Glazed** commands framework. It follows the same architectural patterns as other tools in the ecosystem (`glaze`, `geppetto`, `escuse-me`) — each verb is a self-contained Glazed command that emits structured rows through a processor pipeline, giving the user automatic access to JSON, YAML, CSV, table, and other output formats without any per-command formatting code.

The first verb, **`ttc2ttf`**, extracts individual TrueType Font (`.ttf`) files from a TrueType Collection (`.ttc`) file. This verb is implemented as a **bare command** — it writes files to disk rather than emitting rows, because its primary output is binary font files, not structured data. However, it still uses Glazed for flag parsing, help integration, and the standard CLI infrastructure.

This guide is written for a new intern who needs to understand every layer of the system: the font binary formats, the Go standard library and ecosystem tools available, the Glazed framework's architecture and APIs, the go-go-golems project conventions, and the step-by-step implementation plan. Each section builds on the previous one, so reading in order is recommended.

---

## 1. Problem Statement and Scope

### What problem does font-util solve?

TrueType Collection (`.ttc`) files bundle multiple related fonts into a single binary file. This saves disk space because fonts in a collection share common tables (glyph outlines, metrics, etc.). However, many tools, libraries, and workflows expect individual `.ttf` files — for example, web font serving, font subsetting, and some operating system font installers.

There is currently no widely-used, simple, dependency-free Go CLI that can:

1. Parse a TTC file's binary structure
2. Extract each member font into a standalone, valid TTF file
3. Correctly reassemble the offset table and table directory for each extracted font

Existing solutions are either online converters (requiring uploads), Python scripts using `fonttools`, or Ruby gems — none of which fit cleanly into a Go-based build pipeline or can be distributed as a single static binary.

### Scope for the first verb

The `ttc2ttf` verb handles:

- **Input**: A single `.ttc` file path (positional argument)
- **Output**: One `.ttf` file per member font in the collection, written to an output directory
- **Naming**: Uses the font's internal name (from the `name` table) if extractable, otherwise falls back to `{basename}-{index}.ttf`
- **Flags**: `--output-dir` (destination directory, default: `.`), `--force` (overwrite existing files)

Future verbs (out of scope for now) include `inspect` (font metadata), `subset` (glyph subsetting), and `convert` (WOFF/WOFF2).

---

## 2. The TrueType Collection (TTC) Binary Format

To implement `ttc2ttf`, you need a precise understanding of how TTC and TTF files are structured at the byte level. This section walks through the binary layout with enough detail to write a parser from scratch.

### 2.1 What is a TTC file?

A TrueType Collection (TTC) is a binary container that holds one or more fonts. All fonts in the collection share the same SFNT table structure (the same container format used by individual `.ttf` and `.otf` files), but they are packed into a single file with a TTC header that points to each font's offset table.

The key insight is that **tables can be shared between fonts** within a collection. A glyph outline table (`glyf` + `loca`) might appear once in the file but be referenced by multiple offset tables. When extracting a single font, we must copy all tables that font references — including shared ones — into the standalone TTF output.

### 2.2 TTC Header Layout

The TTC file begins with a header:

```
┌──────────────────────────────────────────────────────┐
│ TTC Header                                           │
├──────────┬───────────┬───────────────────────────────┤
│ Offset   │ Size      │ Field                         │
├──────────┼───────────┼───────────────────────────────┤
│ 0        │ 4 bytes   │ Tag: "ttcf"                   │
│ 4        │ 2 bytes   │ MajorVersion (1)              │
│ 6        │ 2 bytes   │ MinorVersion (0)              │
│ 8        │ 4 bytes   │ NumFonts                      │
│ 12       │ 4×N bytes │ OffsetTable[N] (uint32 each) │
│ 12+4×N   │ 4 bytes   │ DigitalSigOffset (if v2.0)    │
│ 16+4×N   │ 4 bytes   │ DigitalSigLength (if v2.0)    │
└──────────┴───────────┴───────────────────────────────┘
```

- **Tag**: Always the ASCII bytes `ttcf` (0x74746366). This is how you identify that a file is a TTC rather than a standalone TTF.
- **NumFonts**: The number of fonts in the collection. If this is 1, the file might as well be a TTF, but it's still technically valid as a TTC.
- **OffsetTable[]**: An array of `uint32` offsets, each pointing to the start of a font's **Offset Table** within the file. These offsets are relative to the beginning of the file.

### 2.3 Font Offset Table (SFNT Header)

Each font in the collection starts with an Offset Table (also called the SFNT header):

```
┌──────────────────────────────────────────────────────┐
│ Font Offset Table (per font)                         │
├──────────┬───────────┬───────────────────────────────┤
│ Offset   │ Size      │ Field                         │
├──────────┼───────────┼───────────────────────────────┤
│ 0        │ 4 bytes   │ SFNTVersion                   │
│ 4        │ 2 bytes   │ NumTables                     │
│ 6        │ 2 bytes   │ SearchRange                   │
│ 8        │ 2 bytes   │ EntrySelector                 │
│ 10       │ 2 bytes   │ RangeShift                    │
│ 12       │ 16×M     │ TableDirectory[M]              │
└──────────┴───────────┴───────────────────────────────┘
```

- **SFNTVersion**: `0x00010000` for TrueType fonts, or `0x4F54544F` ("OTTO") for CFF-based OpenType fonts. This is how you distinguish a TrueType font from a CFF font within the collection.
- **NumTables**: The number of table records that follow. Each font in the collection can have a different number of tables (especially if some tables are shared and others are private).
- **SearchRange**, **EntrySelector**, **RangeShift**: Binary search optimization fields. These must be recalculated when writing a standalone TTF file because the table directory may be reorganized. The formulas are:
  - `SearchRange = (2^floor(log2(NumTables))) × 16`
  - `EntrySelector = floor(log2(NumTables))`
  - `RangeShift = NumTables × 16 - SearchRange`

### 2.4 Table Record (Table Directory Entry)

Each table in the font is described by a 16-byte record:

```
┌──────────────────────────────────────────────────────┐
│ Table Record                                          │
├──────────┬───────────┬───────────────────────────────┤
│ Offset   │ Size      │ Field                         │
├──────────┼───────────┼───────────────────────────────┤
│ 0        │ 4 bytes   │ Tag (4-char ASCII, e.g. "head")│
│ 4        │ 4 bytes   │ CheckSum                      │
│ 8        │ 4 bytes   │ Offset (from file start)      │
│ 12       │ 4 bytes   │ Length (bytes)                 │
└──────────┴───────────┴───────────────────────────────┘
```

- **Tag**: A 4-character ASCII identifier. Common tags include `head` (font header), `name` (naming table), `glyf` (glyph outlines), `loca` (glyph locations), `cmap` (character-to-glyph mapping), `post` (PostScript name mapping), and `OS/2` (OS/2 and Windows metrics).
- **CheckSum**: A modular checksum of the table's contents. When extracting, you should preserve the existing checksum unless the table data changes (it won't for a straight extraction).
- **Offset**: The byte offset from the start of the file to the table data. **This is the critical field** — in a standalone TTF, all offsets must be relative to the start of the new file, not the original TTC.
- **Length**: The length of the table data in bytes. Table data is padded to 4-byte boundaries.

### 2.5 The Extraction Algorithm

Given the binary layout above, extracting a single font from a TTC works like this:

1. **Read the TTC header**: Parse the tag (verify it's `ttcf`), read `NumFonts`, and read the `OffsetTable[]` array.
2. **Select a font**: Use the offset from `OffsetTable[i]` to jump to the i-th font's offset table.
3. **Read the font's offset table**: Parse `SFNTVersion`, `NumTables`, and the table directory.
4. **Copy each table's data**: For each table record, read `Length` bytes starting at `Offset` in the source TTC.
5. **Reassemble the standalone TTF**: Write a new binary file with:
   - A new offset table (with the same `SFNTVersion` and `NumTables`, but recalculated `SearchRange`/`EntrySelector`/`RangeShift`)
   - A new table directory (with updated `Offset` fields pointing to where each table's data lands in the new file)
   - All table data concatenated after the table directory, each padded to 4-byte alignment

**Diagram of the reassembly:**

```
Source TTC:
┌──────────────┐
│ TTC Header   │
│ ──────────── │
│ Font 0:      │
│   OffsetTable│
│   TableDir   │ ──→ [head data] [name data] [glyf data] ...
│ ──────────── │
│ Font 1:      │
│   OffsetTable│
│   TableDir   │ ──→ [head data] [cmap data] [glyf data*] ...
└──────────────┘
                    * = shared table, same bytes as Font 0's glyf

Output TTF (Font 0):
┌──────────────┐
│ OffsetTable  │  ← new header, recalculated search fields
│ TableDir     │  ← updated offsets pointing into this file
│ [head data]  │  ← copied from source
│ [name data]  │  ← copied from source
│ [glyf data]  │  ← copied from source (even if shared in TTC)
│ ...          │
└──────────────┘
```

### 2.6 Getting the Font Name for Output Filenames

The `name` table (tag `name`) contains the human-readable font name. The relevant name records are:

- **Name ID 1**: Font Family (e.g., "Arial")
- **Name ID 2**: Font Subfamily (e.g., "Bold")
- **Name ID 4**: Full Font Name (e.g., "Arial Bold")
- **Name ID 6**: PostScript Name (e.g., "Arial-BoldMT")

For naming output files, **Name ID 6** (PostScript Name) is the best choice because it is guaranteed to be unique within a collection and uses a safe character set (alphanumeric and hyphens). If the `name` table cannot be parsed, fall back to `{original-basename}-{index}.ttf`.

The `name` table has its own binary structure:

```
┌────────────────────────────────────────┐
│ name Table                             │
├──────────┬──────────┬──────────────────┤
│ 0        │ 2 bytes  │ Format (0 or 1)  │
│ 2        │ 2 bytes  │ Count             │
│ 4        │ 2 bytes  │ StringOffset      │
│ 6        │ 12×Count │ NameRecord[]      │
└──────────┴──────────┴──────────────────┘

Each NameRecord:
┌──────────┬──────────┬──────────────────────┐
│ 0        │ 2 bytes  │ PlatformID            │
│ 2        │ 2 bytes  │ EncodingID            │
│ 4        │ 2 bytes  │ LanguageID            │
│ 6        │ 2 bytes  │ NameID                │
│ 8        │ 2 bytes  │ Length                │
│ 10       │ 2 bytes  │ Offset                │
└──────────┴──────────┴──────────────────────┘
```

For name extraction, prefer **Platform 3 (Windows), Encoding 1 (Unicode BMP), Language 0x0409 (English US)** — these records use UTF-16LE encoding. Convert to plain ASCII/UTF-8 for the filename.

---

## 3. The Go Font Parsing Ecosystem

Several Go libraries are relevant to font-util. This section surveys the options and explains why we choose a particular approach for each use case.

### 3.1 `golang.org/x/image/font/opentype`

**Package**: `golang.org/x/image/font/opentype`
**GoDoc**: https://pkg.go.dev/golang.org/x/image/font/opentype

This is the most actively maintained Go font library. Key types:

- `opentype.Font` — A parsed font with glyph rasterization support
- `opentype.Collection` — A parsed font collection (TTC/OTC)
- `opentype.ParseCollection([]byte) (*Collection, error)` — Parses TTC data
- `opentype.Parse([]byte) (*Font, error)` — Parses a single TTF/OTF

**Limitation for ttc2ttf**: `ParseCollection` returns parsed `Font` objects designed for rasterization, not raw byte streams. You **cannot** use it to produce standalone TTF files because there is no serialization API. The library is lossy — it only parses the tables it needs for rendering.

**Use in font-util**: We will use this library for a future `inspect` verb (where we want metadata, not raw bytes), but **not** for the `ttc2ttf` extraction verb, which requires binary-level access.

### 3.2 `golang.org/x/image/font/sfnt`

**Package**: `golang.org/x/image/font/sfnt`
**GoDoc**: https://pkg.go.dev/golang.org/x/image/font/sfnt

A lower-level parser for SFNT-structured fonts. It provides fine-grained access to font tables and glyph data, but like `opentype`, it's designed for reading, not writing. It does not provide a way to reassemble or serialize font data.

### 3.3 `github.com/golang/freetype/truetype`

**Package**: `github.com/golang/freetype/truetype`

The original Go font parser, now superseded by `x/image`. It can parse both TTF and TTC formats but is unmaintained. Not recommended for new code.

### 3.4 `github.com/ConradIrwin/font/sfnt`

**Package**: `github.com/ConradIrwin/font/sfnt`

A third-party library that provides parsing and encoding for OpenType, TrueType, TTC, WOFF, and WOFF2. This is the only Go library that supports **encoding** (writing) font data, not just parsing. However, it is not actively maintained (last meaningful commit ~2020), and its API is less polished than `x/image`.

**Potential use**: For a WOFF/WOFF2 conversion verb in the future, this might be worth investigating. For the current `ttc2ttf` verb, it's overkill — we can implement the binary extraction directly.

### 3.5 Decision: Custom Binary Parser

For the `ttc2ttf` verb, **we implement a custom binary parser** that reads the TTC structure at the byte level and writes standalone TTF files. The reasons:

1. We need raw byte access to copy table data verbatim (preserving checksums)
2. We need to reassemble the output file with recalculated offsets
3. No existing Go library provides TTC-to-TTF extraction with serialization
4. The TTC/TTF binary format is simple and well-specified — a custom parser is ~200 lines of Go
5. This gives us full control over error messages, validation, and edge cases

The custom parser lives in `pkg/ttc/parser.go` and `pkg/ttc/writer.go`.

---

## 4. The Glazed Commands Framework

This section explains the Glazed framework's architecture and APIs in detail. If you've used Glazed before, you can skim the structure and skip to the code patterns. If you're new to Glazed, read this carefully — it's the foundation of every command in `font-util`.

### 4.1 What is Glazed?

Glazed is a Go framework for building CLI commands that produce **structured data**. Instead of each command implementing its own argument parsing, output formatting, and help text, Glazed provides:

- **Command descriptions**: Declarative flag/argument definitions with types, defaults, and help text
- **Processor pipelines**: A row-by-row processing model where commands emit `Row` objects and middlewares transform/filter/format them
- **Output formatters**: Built-in support for JSON, YAML, CSV, tables, Markdown, and more
- **Cobra integration**: Commands are wired into Cobra (the Go CLI framework) automatically
- **Help system**: A built-in help browser with embedded markdown documentation

The key architectural idea is the **separation of concerns**: a command knows *what data to produce*, and the Glazed pipeline knows *how to format it*. This means every command automatically gets `--output json`, `--output yaml`, `--fields name,size`, `--filter`, and so on without writing any formatting code.

### 4.2 Core Types and Their Relationships

```
┌─────────────────────────────────────────────────────────────┐
│ Glazed Architecture                                          │
│                                                              │
│  CommandDescription                                          │
│  ├── Flags (fields.Field)     ← input schema                │
│  ├── Arguments (fields.Field) ← positional input schema     │
│  ├── Sections (schema.Section) ← grouping of flags          │
│  └── RunIntoGlazeProcessor()  ← the command's logic          │
│       │                                                      │
│       ▼                                                      │
│  Values (parsed flags)                                       │
│  └── DecodeSectionInto()    ← decode into settings struct   │
│                                                              │
│  Processor (middlewares.Processor)                           │
│  ├── AddRow(ctx, Row)       ← emit a row of output          │
│  └── Middlewares             ← transform/filter rows        │
│                                                              │
│  Row (types.Row)                                             │
│  └── MRP(key, value)        ← key-value pairs               │
│                                                              │
│  OutputFormatter                                             │
│  └── JSON / YAML / CSV / Table / ...                         │
└─────────────────────────────────────────────────────────────┘
```

### 4.3 Key Import Paths

All Glazed packages live under `github.com/go-go-golems/glazed/pkg`. Here are the imports you'll use in font-util:

```go
import (
    // Command definition and description
    "github.com/go-go-golems/glazed/pkg/cmds"

    // Field/flag definitions
    "github.com/go-go-golems/glazed/pkg/cmds/fields"

    // Schema constants (schema.DefaultSlug) and section types
    "github.com/go-go-golems/glazed/pkg/cmds/schema"

    // Parsed flag/arg values
    "github.com/go-go-golems/glazed/pkg/cmds/values"

    // Logging helpers
    "github.com/go-go-golems/glazed/pkg/cmds/logging"

    // Processor interface for emitting rows
    "github.com/go-go-golems/glazed/pkg/middlewares"

    // Glazed output section
    "github.com/go-go-golems/glazed/pkg/settings"

    // Row/table types
    "github.com/go-go-golems/glazed/pkg/types"

    // Cobra integration
    "github.com/go-go-golems/glazed/pkg/cli"

    // Help system
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
)
```

**Important**: Do not use older, deprecated paths like `glazed/pkg/cmds/parameters/fields` or `glazed/pkg/values`. Always use the paths above.

### 4.4 The Command Struct Pattern

Every Glazed command follows the same three-part pattern:

1. **Command struct** — embeds `*cmds.CommandDescription`
2. **Settings struct** — maps flags to Go fields via `glazed` struct tags
3. **Constructor** — builds the command description with flags, sections, and help text
4. **Run method** — implements `RunIntoGlazeProcessor(ctx, vals, gp)`

Here is the canonical skeleton:

```go
type Ttc2TtfCommand struct {
    *cmds.CommandDescription
}

type Ttc2TtfSettings struct {
    InputFile  string `glazed:"input-file"`
    OutputDir  string `glazed:"output-dir"`
    Force      bool   `glazed:"force"`
}

func NewTtc2TtfCommand() (*Ttc2TtfCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil {
        return nil, err
    }

    commandSettingsSection, err := cli.NewCommandSettingsSection()
    if err != nil {
        return nil, err
    }

    cmdDesc := cmds.NewCommandDescription(
        "ttc2ttf",
        cmds.WithShort("Extract individual TTF files from a TTC collection"),
        cmds.WithLong(`
Extract each font embedded in a TrueType Collection (.ttc) file
into a standalone TrueType Font (.ttf) file.

Examples:
  font-util ttc2ttf fonts.ttc
  font-util ttc2ttf fonts.ttc --output-dir ./extracted
  font-util ttc2ttf fonts.ttc --force
`),
        cmds.WithFlags(
            fields.New(
                "output-dir",
                fields.TypeString,
                fields.WithDefault("."),
                fields.WithHelp("Directory to write extracted TTF files"),
            ),
            fields.New(
                "force",
                fields.TypeBool,
                fields.WithDefault(false),
                fields.WithHelp("Overwrite existing files"),
            ),
        ),
        cmds.WithArguments(
            fields.New(
                "input-file",
                fields.TypeString,
                fields.WithHelp("Path to the .ttc file to extract"),
                fields.WithIsArgument(true),
            ),
        ),
        cmds.WithSections(glazedSection, commandSettingsSection),
    )

    return &Ttc2TtfCommand{CommandDescription: cmdDesc}, nil
}

func (c *Ttc2TtfCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    s := &Ttc2TtfSettings{}
    if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
        return err
    }

    // ... extraction logic ...
    // Emit a summary row for each extracted font
    row := types.NewRow(
        types.MRP("file", outputPath),
        types.MRP("font_name", fontName),
        types.MRP("tables", numTables),
    )
    return gp.AddRow(ctx, row)
}
```

### 4.5 Bare Commands vs Row-Emitting Commands

Glazed commands come in two flavors:

- **Row-emitting commands**: The normal pattern. The command emits `Row` objects through the processor, and the pipeline formats them as JSON/YAML/CSV/etc. Most commands in the ecosystem work this way.

- **Bare commands**: The command does its own I/O (typically writing files to disk) and uses Glazed only for flag parsing, help integration, and the standard CLI infrastructure. The `RunIntoGlazeProcessor` method still exists but the processor might not be used at all, or might be used for a summary row.

The `ttc2ttf` verb is a **bare command** because its primary output is binary font files on disk, not structured data rows. However, we still emit summary rows so the user can see what was extracted (and pipe it to JSON/table output if desired).

### 4.6 Cobra Integration and Root Command Wiring

The root command of `font-util` follows the same initialization pattern as `glaze`:

```go
// cmd/font-util/main.go

var rootCmd = &cobra.Command{
    Use:     "font-util",
    Short:   "A general-purpose font manipulation tool",
    Version: version,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    // 1. Add logging section (--log-level flag)
    err := logging.AddLoggingSectionToRootCommand(rootCmd, "font-util")
    cobra.CheckErr(err)

    // 2. Initialize help system with embedded docs
    helpSystem := help.NewHelpSystem()
    err = doc.AddDocToHelpSystem(helpSystem)
    cobra.CheckErr(err)
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

    // 3. Register the ttc2ttf command
    ttc2ttfCmd, err := cmds.NewTtc2TtfCommand()
    cobra.CheckErr(err)
    cobraCmd, err := cli.BuildCobraCommand(ttc2ttfCmd,
        cli.WithParserConfig(cli.CobraParserConfig{AppName: "font-util"}),
    )
    cobra.CheckErr(err)
    rootCmd.AddCommand(cobraCmd)

    // 4. Execute
    cobra.CheckErr(rootCmd.Execute())
}
```

This pattern gives us:

- `--log-level` for controlling verbosity
- `help` and `help topics` for browsing embedded documentation
- `--output`, `--fields`, `--filter` (from the glazed section) on each command
- `--print-schema`, `--print-yaml` (from the command settings section) for debugging

### 4.7 Section Composition

Sections are groups of flags that are added to a command. Two standard sections are added to most commands:

- **`settings.NewGlazedSchema()`**: Adds output formatting flags (`--output`, `--fields`, `--filter`, `--stream`, etc.). Every Glazed command that emits rows should include this.
- **`cli.NewCommandSettingsSection()`**: Adds debugging flags (`--print-parsed-fields`, `--print-schema`, `--print-yaml`). Useful during development and for power users.

For the `ttc2ttf` bare command, we include both sections. Even though the primary output is files on disk, the summary rows benefit from the output formatting flags.

---

## 5. Project Structure and go-go-golems Conventions

### 5.1 Directory Layout

The final directory structure for `font-util`:

```
font-util/
├── cmd/
│   └── font-util/
│       └── main.go              # Root Cobra command, wiring, help system
├── pkg/
│   ├── ttc/
│   │   ├── parser.go           # TTC binary parser
│   │   ├── writer.go           # TTF binary writer
│   │   └── parser_test.go      # Unit tests with fixture TTC files
│   └── doc.go                  # Package documentation
├── cmd/
│   └── font-util/
│       └── cmds/
│           ├── root.go          # (optional) if adding command groups
│           └── ttc2ttf.go       # Ttc2TtfCommand Glazed command
├── doc/
│   └── (embedded help markdown files)
├── go.mod
├── go.sum
├── Makefile
├── lefthook.yml
├── .github/
│   └── workflows/               # CI: lint, test, release
├── LICENSE
├── README.md
└── AGENT.md
```

**Wait — there's a conflict**: the existing repo has `cmd/XXX/main.go`. We need to:

1. Rename the module from `github.com/go-go-golems/XXX` to `github.com/go-go-golems/font-util` in `go.mod`
2. Rename the directory `cmd/XXX/` to `cmd/font-util/`
3. Create `cmd/font-util/cmds/ttc2ttf.go` for the Glazed command
4. Create `pkg/ttc/parser.go` and `pkg/ttc/writer.go` for the binary format logic

### 5.2 Module Name

The Go module path will be:

```
module github.com/go-go-golems/font-util
```

All import paths within the repo use this as the prefix:

```go
import (
    "github.com/go-go-golems/font-util/pkg/ttc"
)
```

### 5.3 go.mod Dependencies

```go
module github.com/go-go-golems/font-util

go 1.25

require (
    github.com/go-go-golems/glazed v1.0.5+  // Glazed commands framework
    github.com/spf13/cobra v1.8+             // CLI framework (pulled in by Glazed)
    github.com/pkg/errors v0.9+              // Error wrapping
)
```

We do **not** depend on `golang.org/x/image` for the initial `ttc2ttf` verb because we implement the binary parsing ourselves. This dependency can be added later for the `inspect` verb.

### 5.4 Makefile Updates

The Makefile currently has placeholder `XXX_BINARY` references. Update:

```makefile
FONT_UTIL_BINARY=$(shell which font-util)
install:
	GOWORK=off go build -o ./dist/font-util ./cmd/font-util && \
		cp ./dist/font-util $(FONT_UTIL_BINARY)
```

### 5.5 AGENT.md Updates

Update the module name references and add font-specific build commands:

```markdown
## Build Commands
- Run: `go run ./cmd/font-util`
- Build: `go build ./...`
- Test: `go test ./...`
- Run single test: `go test ./pkg/ttc -run TestParseTTC`
```

---

## 6. Implementation: The TTC Parser

This section provides the full design for the TTC binary parser in `pkg/ttc/parser.go`.

### 6.1 Data Structures

```go
// pkg/ttc/parser.go

package ttc

import (
    "encoding/binary"
    "fmt"
    "io"
    "os"
)

// TTCTag is the magic bytes at the start of a TTC file.
const TTCTag = "ttcf"

// TTCHeader represents the top-level header of a TrueType Collection file.
type TTCHeader struct {
    Tag           string   // Must be "ttcf"
    MajorVersion  uint16   // 1
    MinorVersion  uint16   // 0
    NumFonts      uint32
    FontOffsets   []uint32 // Offsets to each font's OffsetTable
}

// FontHeader represents the Offset Table (SFNT header) of a single font.
type FontHeader struct {
    SFNTVersion   uint32        // 0x00010000 for TrueType
    NumTables     uint16
    SearchRange   uint16
    EntrySelector uint16
    RangeShift    uint16
    TableRecords  []TableRecord
}

// TableRecord represents a single entry in the font's table directory.
type TableRecord struct {
    Tag      [4]byte
    CheckSum uint32
    Offset   uint32
    Length   uint32
}

// Tag returns the 4-character tag as a string.
func (tr TableRecord) Tag() string {
    return string(tr.Tag[:])
}

// TTCFile represents a parsed TTC file containing all member fonts.
type TTCFile struct {
    Header TTCHeader
    Fonts  []FontEntry
}

// FontEntry represents a single font within the TTC, with its header
// and access to the raw file data for table extraction.
type FontEntry struct {
    Header  FontHeader
    Index   int
    Name    string // Extracted from the 'name' table (Name ID 6)
}

// ParseFile reads and parses a TTC file from disk.
func ParseFile(path string) (*TTCFile, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading TTC file: %w", err)
    }
    return Parse(data)
}

// Parse reads and parses a TTC file from a byte slice.
func Parse(data []byte) (*TTCFile, error) {
    // Pseudocode:
    // 1. Read tag (4 bytes) — verify it's "ttcf"
    // 2. Read version (2+2 bytes)
    // 3. Read NumFonts (4 bytes)
    // 4. Read FontOffsets (4 × NumFonts bytes)
    // 5. For each font offset:
    //    a. Read the FontHeader at that offset
    //    b. Read each TableRecord
    //    c. Try to extract the font name from the 'name' table
    // 6. Return TTCFile

    if len(data) < 12 {
        return nil, fmt.Errorf("file too small to be a TTC")
    }

    tag := string(data[0:4])
    if tag != TTCTag {
        return nil, fmt.Errorf("not a TTC file: expected tag 'ttcf', got %q", tag)
    }

    header := TTCHeader{
        Tag:          tag,
        MajorVersion: binary.BigEndian.Uint16(data[4:6]),
        MinorVersion: binary.BigEndian.Uint16(data[6:8]),
        NumFonts:     binary.BigEndian.Uint32(data[8:12]),
    }

    // Read font offsets
    header.FontOffsets = make([]uint32, header.NumFonts)
    for i := uint32(0); i < header.NumFonts; i++ {
        offset := 12 + i*4
        if offset+4 > uint32(len(data)) {
            return nil, fmt.Errorf("font offset %d out of bounds", i)
        }
        header.FontOffsets[i] = binary.BigEndian.Uint32(data[offset : offset+4])
    }

    ttcFile := &TTCFile{Header: header}

    // Parse each font
    for i, fontOffset := range header.FontOffsets {
        fontEntry, err := parseFontEntry(data, fontOffset, i)
        if err != nil {
            return nil, fmt.Errorf("parsing font %d: %w", i, err)
        }
        ttcFile.Fonts = append(ttcFile.Fonts, *fontEntry)
    }

    return ttcFile, nil
}

func parseFontEntry(data []byte, offset uint32, index int) (*FontEntry, error) {
    // Pseudocode:
    // 1. Read SFNTVersion (4 bytes)
    // 2. Read NumTables (2 bytes)
    // 3. Read SearchRange, EntrySelector, RangeShift (6 bytes)
    // 4. For each table: read 16-byte TableRecord
    // 5. Find the 'name' table and extract Name ID 6
    // 6. Return FontEntry

    if offset+12 > uint32(len(data)) {
        return nil, fmt.Errorf("font offset %d out of bounds", offset)
    }

    fontHeader := FontHeader{
        SFNTVersion:   binary.BigEndian.Uint32(data[offset : offset+4]),
        NumTables:     binary.BigEndian.Uint16(data[offset+4 : offset+6]),
        SearchRange:   binary.BigEndian.Uint16(data[offset+6 : offset+8]),
        EntrySelector: binary.BigEndian.Uint16(data[offset+8 : offset+10]),
        RangeShift:    binary.BigEndian.Uint16(data[offset+10 : offset+12]),
    }

    // Parse table records
    fontHeader.TableRecords = make([]TableRecord, fontHeader.NumTables)
    for i := uint16(0); i < fontHeader.NumTables; i++ {
        recOffset := offset + 12 + uint32(i)*16
        if recOffset+16 > uint32(len(data)) {
            return nil, fmt.Errorf("table record %d out of bounds", i)
        }
        copy(fontHeader.TableRecords[i].Tag[:], data[recOffset:recOffset+4])
        fontHeader.TableRecords[i].CheckSum = binary.BigEndian.Uint32(data[recOffset+4 : recOffset+8])
        fontHeader.TableRecords[i].Offset = binary.BigEndian.Uint32(data[recOffset+8 : recOffset+12])
        fontHeader.TableRecords[i].Length = binary.BigEndian.Uint32(data[recOffset+12 : recOffset+16])
    }

    // Extract font name from the 'name' table
    name := extractFontName(data, fontHeader.TableRecords, index)

    return &FontEntry{
        Header: fontHeader,
        Index:  index,
        Name:   name,
    }, nil
}
```

### 6.2 Font Name Extraction

```go
// extractFontName reads the 'name' table from the raw data
// and returns the PostScript name (Name ID 6) from the
// Windows platform (Platform 3, Encoding 1).
// Falls back to "font-{index}" if not found.
func extractFontName(data []byte, records []TableRecord, index int) string {
    // Pseudocode:
    // 1. Find the 'name' table record
    // 2. Read name table header: Format, Count, StringOffset
    // 3. Iterate through NameRecords
    // 4. Look for PlatformID=3, EncodingID=1, NameID=6
    // 5. Read the string at StringOffset + record.Offset
    // 6. Decode from UTF-16LE to UTF-8
    // 7. Sanitize for use as a filename
    // 8. Fall back to fmt.Sprintf("font-%d", index) if not found

    for _, rec := range records {
        if rec.Tag() == "name" {
            // ... parse the name table ...
            // (full implementation in parser.go)
        }
    }
    return fmt.Sprintf("font-%d", index)
}
```

---

## 7. Implementation: The TTF Writer

This section provides the design for the TTF writer in `pkg/ttc/writer.go`.

### 7.1 Writing a Standalone TTF

The writer takes a source TTC's raw bytes and a `FontEntry`, and produces a standalone TTF file:

```go
// pkg/ttc/writer.go

package ttc

import (
    "encoding/binary"
    "fmt"
    "os"
    "path/filepath"
)

// ExtractFont extracts the font at the given index from the TTC data
// and writes a standalone TTF file to outputPath.
func ExtractFont(ttcData []byte, font FontEntry, outputPath string) error {
    // Pseudocode:
    // 1. Calculate the size of the output header:
    //    headerSize = 12 (offset table) + 16 × NumTables (table directory)
    // 2. Allocate a buffer for the output file
    // 3. For each table record in the font:
    //    a. Copy the raw table data from ttcData[rec.Offset : rec.Offset+rec.Length]
    //    b. Pad to 4-byte alignment if necessary
    //    c. Calculate the new offset (where this table will be in the output)
    // 4. Write the offset table header:
    //    - SFNTVersion (same as source)
    //    - NumTables
    //    - Recalculated SearchRange, EntrySelector, RangeShift
    // 5. Write the table directory with updated offsets
    // 6. Write all table data
    // 7. Write buffer to outputPath

    numTables := font.Header.NumTables
    headerSize := uint32(12 + numTables*16)

    // Calculate new offsets for each table
    type tableEntry struct {
        record    TableRecord
        data      []byte
        newOffset uint32
    }

    entries := make([]tableEntry, numTables)
    currentOffset := headerSize

    for i, rec := range font.Header.TableRecords {
        // Copy table data from source TTC
        start := rec.Offset
        end := start + rec.Length
        if end > uint32(len(ttcData)) {
            return fmt.Errorf("table %s data out of bounds (offset=%d, length=%d)",
                rec.Tag(), rec.Offset, rec.Length)
        }

        tableData := make([]byte, rec.Length)
        copy(tableData, ttcData[start:end])

        // Pad to 4-byte boundary
        paddedLen := rec.Length
        if paddedLen%4 != 0 {
            paddedLen += 4 - paddedLen%4
        }

        entries[i] = tableEntry{
            record:    rec,
            data:      tableData,
            newOffset: currentOffset,
        }
        currentOffset += paddedLen
    }

    // Build output buffer
    totalSize := currentOffset
    output := make([]byte, totalSize)

    // Write offset table header
    binary.BigEndian.PutUint32(output[0:4], font.Header.SFNTVersion)
    binary.BigEndian.PutUint16(output[4:6], numTables)

    // Recalculate binary search fields
    searchRange, entrySelector, rangeShift := calcSearchFields(numTables)
    binary.BigEndian.PutUint16(output[6:8], searchRange)
    binary.BigEndian.PutUint16(output[8:10], entrySelector)
    binary.BigEndian.PutUint16(output[10:12], rangeShift)

    // Write table directory
    for i, entry := range entries {
        off := uint32(12 + i*16)
        copy(output[off:off+4], entry.record.Tag[:])
        binary.BigEndian.PutUint32(output[off+4:off+8], entry.record.CheckSum)
        binary.BigEndian.PutUint32(output[off+8:off+12], entry.newOffset)
        binary.BigEndian.PutUint32(output[off+12:off+16], entry.record.Length)
    }

    // Write table data
    for _, entry := range entries {
        copy(output[entry.newOffset:], entry.data)
    }

    // Ensure output directory exists
    dir := filepath.Dir(outputPath)
    if dir != "" && dir != "." {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("creating output directory: %w", err)
        }
    }

    return os.WriteFile(outputPath, output, 0644)
}

// calcSearchFields computes the binary search optimization fields
// for a table directory with numTables entries.
func calcSearchFields(numTables uint16) (searchRange, entrySelector, rangeShift uint16) {
    // searchRange = (2^floor(log2(n))) × 16
    // entrySelector = floor(log2(n))
    // rangeShift = n × 16 - searchRange
    power := uint16(1)
    entrySelector = 0
    for power*2 <= numTables {
        power *= 2
        entrySelector++
    }
    searchRange = power * 16
    rangeShift = numTables*16 - searchRange
    return
}
```

### 7.2 Extracting All Fonts

A convenience function extracts all fonts from a TTC file:

```go
// ExtractAllFonts extracts all fonts from a TTC file and writes
// them to the output directory. Returns a slice of output file paths.
func ExtractAllFonts(ttcPath string, outputDir string, force bool) ([]string, error) {
    data, err := os.ReadFile(ttcPath)
    if err != nil {
        return nil, fmt.Errorf("reading TTC: %w", err)
    }

    ttc, err := Parse(data)
    if err != nil {
        return nil, fmt.Errorf("parsing TTC: %w", err)
    }

    var outputPaths []string
    for _, font := range ttc.Fonts {
        filename := fmt.Sprintf("%s.ttf", font.Name)
        outputPath := filepath.Join(outputDir, filename)

        if !force {
            if _, err := os.Stat(outputPath); err == nil {
                return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", outputPath)
            }
        }

        if err := ExtractFont(data, font, outputPath); err != nil {
            return nil, fmt.Errorf("extracting font %d: %w", font.Index, err)
        }

        outputPaths = append(outputPaths, outputPath)
    }

    return outputPaths, nil
}
```

---

## 8. Implementation: The Glazed Command

This section provides the full implementation of the `ttc2ttf` Glazed command.

### 8.1 Command Definition

File: `cmd/font-util/cmds/ttc2ttf.go`

```go
package cmds

import (
    "context"
    "fmt"

    "github.com/go-go-golems/font-util/pkg/ttc"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/values"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
)

type Ttc2TtfCommand struct {
    *cmds.CommandDescription
}

type Ttc2TtfSettings struct {
    InputFile string `glazed:"input-file"`
    OutputDir string `glazed:"output-dir"`
    Force     bool   `glazed:"force"`
}

func NewTtc2TtfCommand() (*Ttc2TtfCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil {
        return nil, err
    }

    commandSettingsSection, err := cli.NewCommandSettingsSection()
    if err != nil {
        return nil, err
    }

    cmdDesc := cmds.NewCommandDescription(
        "ttc2ttf",
        cmds.WithShort("Extract individual TTF files from a TTC collection"),
        cmds.WithLong(`
Extract each font embedded in a TrueType Collection (.ttc) file
into a standalone TrueType Font (.ttf) file.

The output files are named using the PostScript name (Name ID 6)
from each font's 'name' table. If the name cannot be extracted,
the fallback name is "font-{index}".

Examples:
  font-util ttc2ttf fonts.ttc
  font-util ttc2ttf fonts.ttc --output-dir ./extracted
  font-util ttc2ttf fonts.ttc --force
  font-util ttc2ttf fonts.ttc --output json
`),
        cmds.WithFlags(
            fields.New(
                "output-dir",
                fields.TypeString,
                fields.WithDefault("."),
                fields.WithHelp("Directory to write extracted TTF files to"),
            ),
            fields.New(
                "force",
                fields.TypeBool,
                fields.WithDefault(false),
                fields.WithHelp("Overwrite existing output files"),
            ),
        ),
        cmds.WithArguments(
            fields.New(
                "input-file",
                fields.TypeString,
                fields.WithHelp("Path to the .ttc file to extract"),
                fields.WithIsArgument(true),
            ),
        ),
        cmds.WithSections(glazedSection, commandSettingsSection),
    )

    return &Ttc2TtfCommand{CommandDescription: cmdDesc}, nil
}

func (c *Ttc2TtfCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    s := &Ttc2TtfSettings{}
    if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
        return err
    }

    // Validate input file exists
    if s.InputFile == "" {
        return fmt.Errorf("input-file is required")
    }

    outputPaths, err := ttc.ExtractAllFonts(s.InputFile, s.OutputDir, s.Force)
    if err != nil {
        return err
    }

    // Parse the TTC again for metadata (name, table count)
    // (or refactor ExtractAllFonts to return richer info)
    data, _ := /* ... */
    ttcFile, _ := ttc.Parse(data)

    // Emit one summary row per extracted font
    for i, font := range ttcFile.Fonts {
        row := types.NewRow(
            types.MRP("index", i),
            types.MRP("name", font.Name),
            types.MRP("output", outputPaths[i]),
            types.MRP("tables", font.Header.NumTables),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }

    return nil
}
```

### 8.2 Wiring into the Root Command

File: `cmd/font-util/main.go`

```go
package main

import (
    "github.com/go-go-golems/font-util/cmd/font-util/cmds"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
    "github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
    Use:     "font-util",
    Short:   "A general-purpose font manipulation tool",
    Version: version,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    err := logging.AddLoggingSectionToRootCommand(rootCmd, "font-util")
    cobra.CheckErr(err)

    helpSystem := help.NewHelpSystem()
    // TODO: wire embedded docs when doc/ content is created
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

    ttc2ttfCmd, err := cmds.NewTtc2TtfCommand()
    cobra.CheckErr(err)
    cobraCmd, err := cli.BuildCobraCommand(ttc2ttfCmd,
        cli.WithParserConfig(cli.CobraParserConfig{AppName: "font-util"}),
    )
    cobra.CheckErr(err)
    rootCmd.AddCommand(cobraCmd)

    cobra.CheckErr(rootCmd.Execute())
}
```

---

## 9. Testing Strategy

### 9.1 Unit Tests for the Parser

File: `pkg/ttc/parser_test.go`

Test cases:

1. **Parse a valid TTC file**: Use a small, known TTC fixture file. Verify `NumFonts`, font names, and table counts.
2. **Reject non-TTC files**: Feed a standalone TTF, a PNG, an empty file, and random bytes. Verify error messages.
3. **Parse TTC with shared tables**: Verify that both fonts reference the same table offset (shared table scenario).
4. **Name extraction**: Verify Name ID 6 is correctly read from the `name` table. Test with UTF-16LE encoded names.
5. **Edge cases**: TTC with a single font, TTC with many fonts, TTC with missing `name` table.

### 9.2 Unit Tests for the Writer

1. **Round-trip test**: Extract a font from a TTC, then re-parse the output TTF. Verify it has the same tables, checksums, and data.
2. **Offset calculation**: Verify `SearchRange`, `EntrySelector`, `RangeShift` are correctly computed.
3. **Padding**: Verify table data is padded to 4-byte boundaries.
4. **File writing**: Verify the output file exists, has correct size, and can be re-read.
5. **Overwrite protection**: Verify `--force` flag behavior.

### 9.3 Integration Tests

1. **CLI smoke test**: Run `font-util ttc2ttf testdata/example.ttc --output-dir /tmp/test-output` and verify exit code and output files.
2. **Output format test**: Run with `--output json` and verify the summary rows are valid JSON.

### 9.4 Test Fixtures

Create a small TTC fixture file for testing. You can generate one from existing TTF fonts:

```bash
# Using fonttools (Python) to create a test TTC from two TTFs
pip install fonttools
ttx --merge --output testdata/test.ttc font1.ttf font2.ttf
```

Or download a small TTC from the Google Fonts repository. Store test fixtures in `pkg/ttc/testdata/`.

---

## 10. Phased Implementation Plan

### Phase 1: Project Skeleton (Day 1)

**Goal**: Rename the module, set up the directory structure, and get a compilable `font-util` binary that prints help.

Steps:
1. Rename module in `go.mod` from `github.com/go-go-golems/XXX` to `github.com/go-go-golems/font-util`
2. Rename `cmd/XXX/` to `cmd/font-util/`
3. Add Glazed and Cobra dependencies: `go get github.com/go-go-golems/glazed@latest`
4. Create `cmd/font-util/cmds/ttc2ttf.go` with the command struct (empty `RunIntoGlazeProcessor` that returns nil)
5. Create `cmd/font-util/main.go` with root command wiring
6. Update `Makefile` binary name references
7. Update `AGENT.md` with correct module name
8. Verify: `go build ./cmd/font-util && ./dist/font-util --help` shows the help text

**Validation**: `go build ./...` succeeds, `./dist/font-util ttc2ttf --help` shows flag descriptions.

### Phase 2: TTC Parser (Day 1–2)

**Goal**: Implement and test the TTC binary parser.

Steps:
1. Create `pkg/ttc/parser.go` with `Parse()`, `ParseFile()`, and all data structures
2. Implement `extractFontName()` with UTF-16LE decoding
3. Create `pkg/ttc/parser_test.go` with test cases
4. Add a test fixture TTC file in `pkg/ttc/testdata/`
5. Run tests: `go test ./pkg/ttc/ -v`

**Validation**: All parser tests pass. `ParseFile("testdata/example.ttc")` returns correct `NumFonts` and font names.

### Phase 3: TTF Writer (Day 2)

**Goal**: Implement and test the TTF extraction/writing logic.

Steps:
1. Create `pkg/ttc/writer.go` with `ExtractFont()` and `ExtractAllFonts()`
2. Implement `calcSearchFields()`
3. Add writer tests: round-trip extraction, offset verification, padding
4. Verify extracted TTF files can be opened by font viewers (e.g., `fc-scan` on Linux)

**Validation**: Extracted TTF files are valid (re-parseable, font viewers accept them).

### Phase 4: Command Integration (Day 2)

**Goal**: Wire the parser and writer into the Glazed command.

Steps:
1. Fill in `Ttc2TtfCommand.RunIntoGlazeProcessor()` to call `ttc.ExtractAllFonts()`
2. Add summary row emission for each extracted font
3. Test with `go run ./cmd/font-util ttc2ttf testdata/example.ttc`
4. Test with `--output json` and `--output table`
5. Test `--force` and `--output-dir` flags
6. Add error handling: missing input file, invalid TTC, output directory doesn't exist

**Validation**: The full pipeline works: TTC in → TTF files out + summary rows in chosen format.

### Phase 5: Polish and Documentation (Day 3)

**Goal**: Add help docs, fix lint, update CI.

Steps:
1. Create embedded help markdown in `doc/` for the `ttc2ttf` command
2. Wire `doc.AddDocToHelpSystem(helpSystem)` in `main.go`
3. Run `make lint` and fix any issues
4. Run `make test` and ensure everything passes
5. Update `README.md` with usage examples
6. Commit and push

**Validation**: `make lint && make test && make build` all pass. `./dist/font-util help ttc2ttf` shows the embedded docs.

---

## 11. Risks, Alternatives, and Open Questions

### Risks

- **Checksum validation**: The extracted TTF's `head` table contains a `checkSumAdjustment` field that is computed over the entire file. If we copy tables verbatim without modification, the checksum should still be valid. However, some tools may validate the whole-file checksum on load. We should verify this with real TTC files and, if necessary, recompute the `head` checkSumAdjustment.
- **Shared table corruption**: If two fonts in a TTC share a table (e.g., `glyf`), both extracted TTFs will contain independent copies of the same data. This is correct behavior (the extracted files must be self-contained), but it increases total disk size.
- **CFF-based fonts (OTC)**: Some collections use CFF outlines (OpenType Collection, `.otc`). These have `SFNTVersion = "OTTO"` and a different glyph format. The parser handles them structurally (same offset table), but the extracted files are technically OTF, not TTF. We should detect this and use `.otf` extension.

### Alternatives Considered

1. **Using `golang.org/x/image/font/opentype.ParseCollection`**: Rejected for the extraction verb because the library provides no serialization path. The parsed `Font` objects cannot be turned back into standalone TTF bytes.
2. **Using `ConradIrwin/font/sfnt`**: This library supports encoding, but is unmaintained and would add a dependency for functionality we can implement in ~200 lines of straightforward binary I/O.
3. **Shelling out to `fonttools`**: Would work but defeats the purpose of a single-binary Go tool and introduces a Python dependency.

### Open Questions

1. Should we support extracting a specific font by index (e.g., `--index 2`) rather than always extracting all? This is easy to add but wasn't requested.
2. Should the `name` table extraction support Mac Roman encoding (Platform 1) in addition to Windows Unicode? Most modern TTC files use Platform 3, but legacy fonts might not.
3. Should we recompute the `head` table's `checkSumAdjustment` after extraction? This requires reading the entire output file and modifying 4 bytes in the `head` table. It's straightforward but adds complexity.

---

## 12. API Reference Summary

### Package `pkg/ttc`

| Function | Signature | Description |
|---|---|---|
| `Parse` | `func Parse(data []byte) (*TTCFile, error)` | Parse TTC from byte slice |
| `ParseFile` | `func ParseFile(path string) (*TTCFile, error)` | Parse TTC from file path |
| `ExtractFont` | `func ExtractFont(data []byte, font FontEntry, path string) error` | Extract one font to TTF |
| `ExtractAllFonts` | `func ExtractAllFonts(ttcPath, outputDir string, force bool) ([]string, error)` | Extract all fonts |

### Type `TTCFile`

| Field | Type | Description |
|---|---|---|
| `Header` | `TTCHeader` | TTC header with tag, version, font offsets |
| `Fonts` | `[]FontEntry` | Parsed member fonts |

### Type `FontEntry`

| Field | Type | Description |
|---|---|---|
| `Header` | `FontHeader` | Offset table and table directory |
| `Index` | `int` | Zero-based index within the collection |
| `Name` | `string` | PostScript name from the `name` table |

### Type `FontHeader`

| Field | Type | Description |
|---|---|---|
| `SFNTVersion` | `uint32` | `0x00010000` (TrueType) or `0x4F54544F` (CFF) |
| `NumTables` | `uint16` | Number of table records |
| `TableRecords` | `[]TableRecord` | Table directory entries |

### Type `TableRecord`

| Field | Type | Description |
|---|---|---|
| `Tag` | `[4]byte` | 4-character table identifier |
| `CheckSum` | `uint32` | Table checksum |
| `Offset` | `uint32` | Byte offset from file start |
| `Length` | `uint32` | Table data length in bytes |

---

## 13. Key File References

| File | Purpose |
|---|---|
| `/home/manuel/code/wesen/go-go-golems/font-util/go.mod` | Module definition (currently `XXX`, needs rename) |
| `/home/manuel/code/wesen/go-go-golems/font-util/cmd/XXX/main.go` | Entry point (currently empty, needs rename) |
| `/home/manuel/code/wesen/go-go-golems/font-util/Makefile` | Build targets (needs binary name update) |
| `/home/manuel/code/wesen/go-go-golems/font-util/AGENT.md` | Agent guidelines (needs module name update) |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go` | Reference: Glazed root command wiring |
| `/home/manuel/.pi/agent/skills/glazed-command-authoring/SKILL.md` | Reference: Glazed command authoring patterns |
| Microsoft OpenType Spec | https://learn.microsoft.com/en-us/typography/opentype/spec/otff |
| Apple TrueType Reference | https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html |
