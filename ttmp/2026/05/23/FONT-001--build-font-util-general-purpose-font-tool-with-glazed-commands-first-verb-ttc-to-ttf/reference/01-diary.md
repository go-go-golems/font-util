---
type: reference
title: Diary
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

# Diary

## Goal

Capture the step-by-step investigation, design, and documentation journey for building `font-util` — a general-purpose font tool built on the Glazed commands framework, with the first verb being TTC-to-TTF conversion.

## Step 1: Ticket Creation and Evidence Gathering

This step covers the initial ticket creation, codebase inspection, and external research needed to produce the implementation guide. The font-util repo exists as a bare go-go-golems template (module `github.com/go-go-golems/XXX`, empty `main.go`, empty `pkg/doc.go`). The task is to design the full system before writing any code, and produce an intern-ready implementation guide.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to build a font-util general purpose tool using glazed commands framework, with the first verb being to convert ttc to ttf files (just barecommand). Create a detailed implementation guide. Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and then upload to remarkable. Keep a diary as you work. Don't work outside of this diary."

**Assistant interpretation:** Create a docmgr ticket FONT-001 for building font-util, write a comprehensive intern-ready implementation guide covering all subsystems, store in the ticket, and upload to reMarkable. Keep a diary of all work.

**Inferred user intent:** Have a complete, onboarding-quality design document that a new team member could use to implement font-util from scratch, without needing to ask questions about the Glazed framework, TTC binary format, or Go project structure conventions.

### What I did

- Read the glazed-command-authoring skill for current Glazed API patterns and import paths
- Read the ticket-research-docmgr-remarkable and diary skills for workflow conventions
- Inspected the font-util repo: empty template with `go.mod` (module `github.com/go-go-golems/XXX`), empty `cmd/XXX/main.go`, empty `pkg/doc.go`
- Examined the glazed CLI root command pattern from `glazed/cmd/glaze/main.go`
- Searched for Go libraries that parse TTC files: `golang.org/x/image/font/opentype.ParseCollection`, `github.com/ConradIrwin/font/sfnt`, `github.com/golang/freetype/truetype`
- Researched TTC binary format specification from Microsoft OpenType docs and Apple TrueType Reference Manual
- Created docmgr ticket FONT-001 with topics: fonts, glazed, cli, go, binary-parsing
- Created design-doc "Implementation Guide" and reference "Diary"
- Wrote the diary (this document) and the main implementation guide

### Why

The user wants a thorough design document that serves as a self-contained onboarding guide for an intern. This requires covering not just the font-util-specific code, but also the Glazed framework, the go-go-golems project conventions, the TTC binary format, and the full directory structure — everything an intern needs to implement the tool without prior context.

### What worked

- The glazed-command-authoring skill provided the exact current API patterns (import paths, constructor pattern, `RunIntoGlazeProcessor` signature)
- The glazed `main.go` served as a concrete reference for root command wiring
- The `golang.org/x/image/font/opentype.ParseCollection` API is well-documented and provides exactly what's needed for parsing TTC files
- The TTC format specification from Microsoft is clear and implementable

### What didn't work

- The `ConradIrwin/font` library is older and less maintained than `golang.org/x/image`; decided to prefer the standard x/image library for the implementation guide
- The freetype/truetype library also works but is lower-level; `opentype.ParseCollection` is the cleanest entry point for TTC parsing

### What I learned

- The `golang.org/x/image/font/opentype` package has `ParseCollection` which returns `*Collection` with `Fonts []Font` — this is the cleanest Go API for reading TTC files
- However, `ParseCollection` returns parsed font objects, not raw bytes; for TTC-to-TTF extraction we actually need to work at the binary level (reading TTC header, offset tables, and reassembling individual TTF byte streams). The x/image library can validate/inspect but doesn't directly produce standalone TTF file bytes from a TTC.
- This means the primary implementation approach should parse the TTC binary format directly, since we need to produce valid standalone TTF files (with corrected offset tables and table directory entries)
- The go-go-golems template repo is very bare (module still named XXX, main.go is empty), so the guide needs to cover renaming the module and setting up the full directory structure

### What was tricky to build

- Understanding the distinction between "parsing a TTC for rendering" (what x/image does) vs "extracting individual TTF byte streams from a TTC" (what font-util needs). The x/image library parses fonts for rasterization; it doesn't provide a way to serialize back to standalone TTF. The binary-level approach is required.
- The TTC format has shared tables between member fonts — extracting a single font requires copying both its private tables and any shared tables it references into the output TTF.

### What warrants a second pair of eyes

- The decision to implement TTC parsing at the binary level rather than wrapping `opentype.ParseCollection` — this is the right call for file extraction but worth confirming
- The reassembly of the TTF offset table: the output file needs a new offset table with recalculated `searchRange`, `entrySelector`, and `rangeShift` values per the TrueType spec
- Whether `golang.org/x/image/font/opentype` should still be used as a validation/inspection library (e.g., for a future `inspect` command)

### What should be done in the future

- Implement the `ttc2ttf` command per the guide
- Add an `inspect` verb that uses `opentype.ParseCollection` for font metadata
- Add WOFF/WOFF2 support in a future verb
- Consider using `ConradIrwin/font/sfnt` for OTF/CFF-based collection (OTC) support

### Code review instructions

- Start at the implementation guide (design doc) for the full picture
- Verify the TTC binary format description against the Microsoft OpenType spec
- Check that the Glazed command wiring pattern matches the current glazed-command-authoring skill

### Technical details

**Key external references:**
- Microsoft OpenType spec (TTC Header): https://learn.microsoft.com/en-us/typography/opentype/spec/otff
- Apple TrueType Reference Manual: https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html
- Go x/image/font/opentype: https://pkg.go.dev/golang.org/x/image/font/opentype
- Go freetype/truetype: https://pkg.go.dev/github.com/golang/freetype/truetype

**Key repo files inspected:**
- `/home/manuel/code/wesen/go-go-golems/font-util/go.mod` — module still named `XXX`
- `/home/manuel/code/wesen/go-go-golems/font-util/cmd/XXX/main.go` — empty main
- `/home/manuel/code/wesen/go-go-golems/font-util/pkg/doc.go` — empty package
- `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go` — reference root command pattern

**TTC binary format summary:**
```
TTC Header:
  Tag:           "ttcf" (4 bytes)
  MajorVersion:  uint16 (1)
  MinorVersion:  uint16 (0)
  NumFonts:      uint32
  OffsetTable:   []uint32 (NumFonts entries, each points to an OffsetTable)
  (Optional) DigitalSignature fields

Per-font OffsetTable (at each offset):
  SFNTVersion:   uint32 (0x00010000 for TrueType, "OTTO" for CFF)
  NumTables:     uint16
  SearchRange:   uint16
  EntrySelector: uint16
  RangeShift:    uint16
  TableDirectory: []TableRecord (NumTables entries)
    Tag:     [4]byte
    CheckSum: uint32
    Offset:   uint32
    Length:   uint32
```

---

## Step 2: Writing the Implementation Guide and Ticket Bookkeeping

This step covers writing the main design-doc deliverable (the intern-ready implementation guide), creating ticket tasks for the 5 implementation phases, relating key repo files, and updating the changelog. The guide covers 13 sections: executive summary, problem statement, TTC binary format, Go font parsing ecosystem, Glazed framework, project structure, parser implementation, writer implementation, Glazed command wiring, testing strategy, phased implementation plan, risks/alternatives, and API reference.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Write the full implementation guide, create docmgr bookkeeping, then upload to reMarkable.

**Inferred user intent:** A comprehensive, self-contained document that a new intern can use to implement font-util from scratch.

### What I did
- Wrote the 53KB implementation guide in the design-doc slot, covering all 13 sections
- Added 5 tasks for the phased implementation plan (Phases 1-5)
- Related 5 key repo files (go.mod, main.go, doc.go, Makefile, glazed reference)
- Updated changelog with the creation of the guide and diary
- Updated diary with Step 2

### Why
The user explicitly asked for a detailed analysis/design/implementation guide for a new intern, with prose paragraphs, bullet points, pseudocode, diagrams, API references, and file references. The guide must be exhaustive enough for someone unfamiliar with the codebase or the Glazed framework.

### What worked
- The glazed-command-authoring skill provided the exact current API patterns, which made the command wiring section precise
- The TTC binary format specification from Microsoft was clear enough to write the parser pseudocode directly
- Breaking the implementation into 5 phases with day estimates gives the intern a realistic timeline

### What didn't work
- Nothing failed in this step — it was a documentation-only step

### What I learned
- The guide ended up quite long (~53KB) because covering the Glazed framework, TTC format, Go ecosystem, and implementation details all in one document requires substantial prose. This is appropriate for an intern onboarding guide but would be overkill for a senior engineer.
- The distinction between bare commands and row-emitting commands in Glazed is important and needed explicit explanation since ttc2ttf is a bare command.

### What was tricky to build
- Deciding the right level of detail for the TTC binary format section — too little and the intern can't implement the parser, too much and it becomes a spec dump. I aimed for the "explain the structure with exact byte offsets and then provide pseudocode" level.
- The font name extraction from the `name` table is a mini-format within a format — I documented it but left the full implementation for the code phase.

### What warrants a second pair of eyes
- Whether the `head` table `checkSumAdjustment` needs recomputation after extraction — I flagged this as an open question in the guide
- Whether the guide should include a complete, copy-pasteable implementation or just pseudocode — I went with pseudocode + partial real code, which is the right balance for learning

### What should be done in the future
- Upload the guide to reMarkable (next step) ← **DONE**
- Run `docmgr doctor` to validate before upload ← **DONE**
- Implement the actual code per the phased plan

### Code review instructions
- Read the design-doc for completeness and accuracy
- Verify the TTC format description against the Microsoft spec
- Check that the Glazed command pattern matches the current glazed-command-authoring skill

### Technical details

**Document path:** `ttmp/2026/05/23/FONT-001--.../design-doc/01-implementation-guide-font-util-with-glazed-commands.md`
**Size:** ~53KB
**Sections:** 13 (Executive Summary through Key File References)
**Tasks created:** 5 (Phases 1-5)
**Files related:** 5 (go.mod, main.go, doc.go, Makefile, glazed reference)

---

## Step 3: Doctor Validation and reMarkable Upload

This step covers running `docmgr doctor` to validate the ticket artifacts, fixing vocabulary and related-file warnings, and uploading the bundle to reMarkable.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Validate ticket health and upload the design doc + diary to reMarkable.

**Inferred user intent:** Ensure the ticket is clean and the deliverable is accessible on reMarkable for offline reading.

### What I did
- Ran `docmgr doctor --ticket FONT-001 --stale-after 30`
- Fixed vocabulary: added 5 missing topic slugs (binary-parsing, cli, fonts, glazed, go)
- Fixed related file: removed broken relative path for glazed main.go, re-added with absolute path
- Re-ran doctor — all checks passed
- Uploaded bundle (design-doc + diary) to reMarkable at `/ai/2026/05/23/FONT-001`
- Bundle name: "FONT-001 font-util Implementation Guide"

### Why
The deliverable checklist requires both doctor validation and reMarkable upload before handoff.

### What worked
- Doctor validation caught the vocabulary and path issues early
- reMarkable upload succeeded on the first attempt with `--toc-depth 2`

### What didn't work
- The initial `docmgr doc relate --file-note` with a path outside the repo created a broken relative path. Had to manually edit the frontmatter to use an absolute path.

### What I learned
- `docmgr doc relate` normalizes paths relative to the ticket directory, which breaks for files outside the repo. Use `--remove-files` and manual frontmatter edits for cross-repo references.

### What was tricky to build
- The related-file path resolution: docmgr stores relative paths, but when the referenced file is outside the repo tree, the relative path becomes deeply nested (`../../../../../../../...`) and doesn't resolve correctly from the ticket directory.

### What warrants a second pair of eyes
- The reMarkable PDF rendering — need to verify on the device that the code blocks and tables render correctly. Large code blocks can overflow margins in pandoc-generated PDFs.

### What should be done in the future
- Verify the PDF rendering on the reMarkable device
- Begin Phase 1 implementation (project skeleton)

### Code review instructions
- Check the reMarkable upload at `/ai/2026/05/23/FONT-001`
- Verify doctor passes: `docmgr doctor --ticket FONT-001 --stale-after 30`

### Technical details

**reMarkable upload:** `OK: uploaded FONT-001 font-util Implementation Guide.pdf -> /ai/2026/05/23/FONT-001`
**Doctor result:** ✅ All checks passed
**Vocabulary entries added:** 5 (binary-parsing, cli, fonts, glazed, go)

---

## Step 4: Begin Implementation — Phase 1 Project Skeleton

The user asked to go through the tasks one by one, keeping a detailed diary and committing at appropriate intervals. They also uploaded three TTC test files (Didot.ttc, Futura.ttc, GillSans.ttc) to the repo root. I added `*.ttc` to `.gitignore` to avoid committing large binary test data, and replaced the coarse 5-phase tasks with 19 granular tasks covering the full implementation.

### Prompt Context

**User prompt (verbatim):** "Add tasks to the ticket, then go through them task by task, keeping a detailed diary as you work, and committing at appropriate intervals. I'm uploading a few ttc in this directory too."

**Assistant interpretation:** Add granular tasks to FONT-001, then implement each one sequentially, updating the diary after each meaningful commit, and committing code at natural boundaries.

**Inferred user intent:** Actually build font-util end-to-end, not just design it. Use the TTC files as real test data.

### What I did
- Added `*.ttc` to `.gitignore` (724KB + 487KB + 1.2MB = too large for git)
- Replaced 5 coarse tasks with 19 granular tasks in `tasks.md`
- Beginning Phase 1 implementation

### Why
The user wants working code, not just design docs. The TTC files provide real test data for validation.

### What worked
- The TTC files are standard macOS system fonts — perfect test data

### What didn't work
- N/A (just starting)

### What I learned
- GillSans.ttc is the largest (1.2MB) — likely contains the most fonts/most shared tables

### What was tricky to build
- N/A yet

### What warrants a second pair of eyes
- N/A yet

### What should be done in the future
- Complete all Phase 1 tasks (1.1–1.8)

### Code review instructions
- Check `.gitignore` includes `*.ttc`

### Technical details
**TTC files:**
- Didot.ttc: 724,008 bytes
- Futura.ttc: 487,620 bytes
- GillSans.ttc: 1,254,028 bytes

---

## Step 5: Phase 1 Implementation — Project Skeleton (Tasks 1.1–1.8)

Implemented all Phase 1 tasks: renamed the Go module, restructured directories, added Glazed/Cobra dependencies, created the ttc2ttf command stub with full flag definitions and help text, updated all config files (goreleaser, Makefile, AGENT.md), and verified the binary compiles and `--help` works correctly.

### Prompt Context

**User prompt (verbatim):** (see Step 4)

**Assistant interpretation:** Execute all tasks sequentially, commit at natural boundaries.

**Inferred user intent:** Get the project skeleton working so we can iterate on the actual font parsing logic.

**Commit (code):** 774777e — "feat: project skeleton - rename module, add Glazed CLI with ttc2ttf stub"

### What I did
- Task 1.1: Renamed Go module from `github.com/go-go-golems/XXX` to `github.com/go-go-golems/font-util` in go.mod
- Task 1.2: Moved `cmd/XXX/` to `cmd/font-util/`
- Task 1.3: Added Glazed v1.2.14, Clay v0.4.9, Cobra v1.10.2, pkg/errors v0.9.1
- Task 1.4: Created `pkg/ttc/` directory
- Task 1.5: Removed empty `pkg/doc.go`
- Task 1.6: Updated `.goreleaser.yaml` — project_name, binary, main paths, homepage, brew name all changed from XXX to font-util
- Task 1.7: Updated `Makefile` — replaced XXX_BINARY with FONT_UTIL_BINARY, added `run` target, updated install target, updated release module path
- Task 1.8: Updated `AGENT.md` — module name, binary name, build commands, project structure
- Created `cmd/font-util/cmds/ttc2ttf.go` with full Glazed command definition (flags, arguments, sections, help text)
- Created `cmd/font-util/main.go` with root Cobra command, logging, help system, ttc2ttf wiring
- Verified: `go build ./...` succeeds, `go run ./cmd/font-util --help` shows help, `go run ./cmd/font-util ttc2ttf --help` shows ttc2ttf help with all Glazed output flags

### Why
The skeleton needs to compile and show help before we can add the actual parsing logic.

### What worked
- Everything compiled on the first attempt after `go mod tidy`
- The Glazed framework automatically provides all output formatting flags (--output, --fields, --filter, etc.)
- The `--log-level` flag comes free from the logging section

### What didn't work
- Initially forgot to run `go mod tidy` after adding imports — got "no required module provides package" errors. Fixed by running `go mod tidy`.

### What I learned
- The Glazed CLI framework provides an enormous amount of functionality out of the box: 60+ flags for output formatting, filtering, templating, jq queries, pagination, etc. All free just by using `settings.NewGlazedSchema()` and `cli.NewCommandSettingsSection()`.

### What was tricky to build
- Nothing tricky in Phase 1 — it's mechanical skeleton work.

### What warrants a second pair of eyes
- The `.goreleaser.yaml` has CGO_ENABLED=1 set. This is the default from the template but might not be needed for font-util (no C dependencies). Could change to CGO_ENABLED=0 for simpler static builds. However, leaving it as-is is safe.

### What should be done in the future
- Consider CGO_ENABLED=0 for simpler cross-compilation

### Code review instructions
- Run `go build ./...` and `go run ./cmd/font-util ttc2ttf --help`
- Check `.goreleaser.yaml` project_name is `font-util`
- Check `go.mod` module is `github.com/go-go-golems/font-util`

### Technical details

**Commit:** 774777e
**Glazed version:** v1.2.14
**Cobra version:** v1.10.2

---

## Step 6: Phase 2+3 — TTC Parser and TTF Writer Implementation

Implemented the full TTC binary parser (`pkg/ttc/parser.go`) and TTF writer (`pkg/ttc/writer.go`), wired them into the ttc2ttf Glazed command, and wrote comprehensive tests. The tool now works end-to-end on real TTC files.

**Commit (code):** ff638bb — "feat: implement TTC parser, TTF writer, and wire ttc2ttf command"

### What I did
- Implemented `pkg/ttc/parser.go`: TTC header parsing, per-font offset table parsing, table record parsing, bounds validation, font name extraction from `name` table (Name ID 6, Platform 3 Windows + Platform 1 Mac Roman fallback), UTF-16BE decoding, filename sanitization, `CalcSearchFields`
- Implemented `pkg/ttc/writer.go`: `ExtractFont` (reassembles standalone TTF with recalculated offsets and search fields, 4-byte padding), `ExtractAllFonts` (convenience function with force/overwrite protection)
- Implemented `pkg/ttc/parser_test.go`: Tests with all 3 TTC files (Didot 3 fonts, Futura 5 fonts, GillSans 9 fonts), error rejection tests, CalcSearchFields unit tests, round-trip extraction test, ExtractAllFonts test
- Wired `cmd/font-util/cmds/ttc2ttf.go` to call `ttc.ExtractAllFonts` and emit summary rows
- All tests pass, all extracted files validated with `fc-scan`

### Why
This is the core functionality — reading a TTC binary and producing valid standalone TTF files.

### What worked
- The binary parser worked on the first attempt against all 3 real TTC files
- Font name extraction correctly found Name ID 6 from the Windows platform for all 17 fonts across the 3 TTC files
- `fc-scan` validates that the extracted TTF files are properly formed
- The `--force` flag correctly prevents accidental overwrites

### What didn't work
- Initially had a Go naming conflict: `Tag` field and `Tag()` method had the same name. Fixed by renaming the field to `TagBytes`.
- `nonamedreturns` linter flagged `CalcSearchFields`. Fixed by using unnamed return types.

### What I learned
- macOS TTC files consistently use Platform 3 (Windows), Encoding 1 (Unicode BMP) for name records. The Mac Roman fallback wasn't needed for these test files but is there for robustness.
- Shared tables in TTC files don't cause any issues during extraction — we copy the raw bytes regardless of whether they're shared, so each output TTF is self-contained.
- The Go `encoding/binary` package makes the TTC parsing straightforward — no need for struct-based binary reading, just explicit slice indexing.

### What was tricky to build
- The `name` table is a mini-format within a format. You have to: (1) find the `name` table record, (2) parse the name table header to find the string storage offset, (3) iterate through name records checking platform/encoding/name IDs, (4) read the string at the correct offset within the string storage, (5) decode UTF-16BE to UTF-8. Each step has its own bounds checking.
- The `CalcSearchFields` function had to change from named returns (which the linter rejects) to unnamed returns, which means the caller has to remember the return order. This is acceptable for a simple utility function.

### What warrants a second pair of eyes
- The `CalcSearchFields` function now returns `(uint16, uint16, uint16)` — the caller must match the order (searchRange, entrySelector, rangeShift). This is error-prone. Consider a small struct if more callers are added.

### What should be done in the future
- Add a `--list` flag to ttc2ttf that shows font names without extracting
- Add an `--index` flag to extract only a specific font
- Add head table checkSumAdjustment recomputation (open question from design doc)
- Consider OTC/CFF support (SFNTVersion = "OTTO" → use .otf extension)

### Code review instructions
- Run `go test ./pkg/ttc/ -v -count=1` — all tests should pass
- Run `go run ./cmd/font-util ttc2ttf Didot.ttc` — should extract 3 fonts
- Run `fc-scan` on any extracted TTF to verify font validity

### Technical details

**Parsed results:**
- Didot.ttc: 3 fonts (Didot, Didot-Italic, Didot-Bold)
- Futura.ttc: 5 fonts (Futura-Medium, Futura-MediumItalic, Futura-Bold, Futura-CondensedMedium, Futura-CondensedExtraBold)
- GillSans.ttc: 9 fonts (GillSans, GillSans-Bold, GillSans-Italic, GillSans-BoldItalic, GillSans-SemiBold, GillSans-SemiBoldItalic, GillSans-UltraBold, GillSans-Light, GillSans-LightItalic)

---

## Step 7: Phase 5 — Lint, Test, README (Tasks 5.1–5.4)

Fixed lint issues, verified all tests pass, and updated the README with usage examples. All tasks complete.

**Commit (code):** 9907509 — "chore: fix lint issues, update README with usage examples"

### What I did
- Task 5.1: Fixed `nonamedreturns` lint error in `CalcSearchFields` (changed to unnamed returns)
- Task 5.1: Fixed gofmt formatting issue in parser.go
- Task 5.2: `make test` passes — all pkg/ttc tests green
- Task 5.3: Wrote README.md with ttc2ttf usage examples, output format docs, build instructions
- Task 5.4: Committed

### Why
Clean lint and tests are required before any release. README is the first thing users see.

### What worked
- Lint passed cleanly after the two fixes
- No test changes needed

### What didn't work
- N/A

### What I learned
- The `nonamedreturns` linter in this project is strict — it flags all named returns, not just unused ones. This is a valid style choice (named returns can be confusing), but it means utility functions with multiple same-typed returns need careful documentation.

### What was tricky to build
- Nothing tricky in this step — it was cleanup.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Push to origin
- Consider adding more verbs (inspect, list, subset)

### Code review instructions
- Run `make lint && make test && make build` — all should pass
- Read README.md for accuracy

### Technical details
All 20 tasks completed. All checks pass.

---

## Step 8: Integrating typo-copy-generator into font-util

The user asked to integrate the functionality of `~/code/wesen/2026-05-22--typo-copy-generator/` into font-util. This is a standalone Go CLI that creates printable typography copy-practice PDFs from OpenType/TrueType fonts and YAML worksheet templates. It has 3 subcommands: `init-template`, `inspect-font`, `render`. The code is in `internal/` packages: fontmetrics, spec, shape, layout, renderpdf. I will copy these packages verbatim into `pkg/` and then adapt the CLI layer to use Glazed commands.

### Prompt Context

**User prompt (verbatim):** "ok, now integrate the functionality of ~/code/wesen/2026-05-22--typo-copy-generator/ into this tool as well, you can literally copy it verbatim and then adapt if you want."

**Assistant interpretation:** Copy the typo-copy-generator packages into font-util's pkg/ directory, adapt the module path imports, and wrap the 3 subcommands (init-template, inspect-font, render) as Glazed commands.

**Inferred user intent:** Consolidate both font tools into a single font-util binary. The copy-practice PDF generator becomes another verb group under font-util.

### What I did
- Inspected all source files in typo-copy-generator: 6 Go packages + main + examples
- Identified the packages to copy: fontmetrics, spec, shape, layout, renderpdf
- Identified the CLI subcommands to convert: init-template, inspect-font, render
- Identified additional dependencies: go-pdf/fpdf, go-text/typesetting, golang.org/x/image, gopkg.in/yaml.v3
- Started copying packages

### Why
Consolidating into one binary is more convenient than maintaining two separate tools, and the font-util name covers this functionality well.

### What worked
- The typo-copy-generator code is well-structured with clean package boundaries
- Tests are self-contained (use gofont/goregular embedded test data)

### What didn't work
- N/A (investigation phase)

### What I learned
- The typo-copy-generator uses `github.com/go-text/typesetting/harfbuzz` for HarfBuzz shaping, with a fallback SFNT kerning path. This is a heavy dependency but important for proper ligature/kerning rendering.
- The renderpdf package uses `github.com/go-pdf/fpdf` for PDF generation with vector glyph outlines.
- The spec package handles YAML template parsing with comprehensive defaults.

### What was tricky to build
- N/A yet

### What warrants a second pair of eyes
- The go-text/typesetting dependency is large. Need to verify it doesn't cause build issues.

### What should be done in the future
- Copy packages, fix imports, add Glazed commands, wire into root

### Code review instructions
- Compare copied packages against originals to verify no content was lost

### Technical details

**Source packages to copy:**
- `internal/fontmetrics` → `pkg/fontmetrics` (font loading, OS/2 metrics extraction, scaling)
- `internal/spec` → `pkg/spec` (YAML template parsing, defaults, validation, resolution)
- `internal/shape` → `pkg/shape` (HarfBuzz shaping + SFNT kern fallback)
- `internal/layout` → `pkg/layout` (page layout, row/cell placement, page breaks)
- `internal/renderpdf` → `pkg/renderpdf` (fpdf rendering, glyph vector drawing, helper lines)

**Dependencies to add:**
- `github.com/go-pdf/fpdf v0.9.0`
- `github.com/go-text/typesetting v0.3.4`
- `golang.org/x/image v0.41.0`
- `gopkg.in/yaml.v3 v3.0.1`

**Commands to create:**
- `font-util init-template` (bare command - writes YAML file)
- `font-util inspect-font` (row-emitting - outputs metrics/shaping data)
- `font-util render` (bare command - writes PDF file)

---

## Step 9: Integrate typo-copy-generator into font-util

Copied 5 packages from `~/code/wesen/2026-05-22--typo-copy-generator/` into `pkg/`, fixed import paths, added dependencies, created 3 Glazed commands (init-template, inspect-font, render), wired them into the root command, and added TTC file support for inspect-font and render.

**Commit (code):** 41d915b — "feat: copy typo-copy-generator packages into font-util"
**Commit (code):** fc16756 — "feat: integrate typo-copy-generator as Glazed commands (init-template, inspect-font, render)"

### Prompt Context

**User prompt (verbatim):** "ok, now integrate the functionality of ~/code/wesen/2026-05-22--typo-copy-generator/ into this tool as well, you can literally copy it verbatim and then adapt if you want."

**Assistant interpretation:** Copy the typo-copy-generator source packages verbatim, fix import paths, wrap the 3 subcommands as Glazed commands, and wire into font-util's root.

**Inferred user intent:** Consolidate both font tools into a single binary with shared infrastructure.

### What I did
- Copied 5 packages: fontmetrics, spec, shape, layout, renderpdf
- Fixed all import paths from `github.com/wesen/typo-copy-generator/internal/` to `github.com/go-go-golems/font-util/pkg/`
- Added dependencies: go-pdf/fpdf v0.9.0, go-text/typesetting v0.3.4, golang.org/x/image v0.41.0, gopkg.in/yaml.v3 v3.0.1
- Created `init-template` command (bare command, writes YAML file)
- Created `inspect-font` command (row-emitting, outputs metrics + shaping data)
- Created `render` command (bare command, writes PDF; dry-run mode emits layout rows)
- Added TTC file support in inspect-font and render (extracts first font to temp file)
- Fixed lint issues: gofmt, max→maxf rename, errcheck for os.Remove/tmpFile.Close
- Updated README with all 4 commands
- All 7 package tests pass, lint clean

### Why
Consolidating font tools reduces tool sprawl and gives all commands the shared Glazed infrastructure (output formatting, help system, logging).

### What worked
- The verbatim copy approach worked well — the packages are self-contained with clean boundaries
- All 6 original tests passed immediately after fixing import paths
- TTC file support was easy to add via the existing ttc package — just extract to a temp file
- The render command's `--dry-run` mode works beautifully with Glazed's output formatting

### What didn't work
- Had a flag name conflict: `--template` was defined in both the render command and the Glazed output section. Renamed to `--yaml-template`.
- The `max` function in layout.go conflicted with Go 1.21+'s built-in `max`. Renamed to `maxf`.
- errcheck flagged unchecked returns on `os.Remove` and `tmpFile.Close` in the temp file pattern.
- Had to clean up a botched sed deletion of the `marshalJSON` function.

### What I learned
- Go 1.21+ predeclares `max` and `min`, so any package that defines its own will fail the `predeclared` linter. The typo-copy-generator was written before this became an issue.
- The `go-text/typesetting` harfbuzz integration provides significantly better shaping than the SFNT kern fallback — ligature glyphs (like "fi") are properly resolved as single glyphs.
- TTC files cannot be directly read by `golang.org/x/image/font/sfnt.Parse` — it rejects collections with "invalid single font". The workaround of extracting to a temp TTF file works but is not ideal for production. A future improvement would be to add native TTC support to the fontmetrics package.

### What was tricky to build
- The TTC temp-file pattern is a pragmatic workaround but has edge cases: the temp file is removed on defer, which means if the render is long-running, the file could be cleaned up early. In practice this is fine since render is a one-shot CLI command.
- The `marshalJSON` function cleanup via sed was messy — it left a dangling comment and partial function body. Had to manually fix.

### What warrants a second pair of eyes
- The TTC temp file pattern in inspect-font and render — both extract the first font. For a TTC with multiple fonts, the user might want to specify which font to inspect/render. This is a future feature.
- The `--yaml-template` flag name is a departure from the original `--template`. Users familiar with typo-copy-generator will need to adjust.

### What should be done in the future
- Add `--font-index` flag to inspect-font and render to select a specific font from TTC
- Add native TTC support to fontmetrics (avoid temp file extraction)
- Consider adding an `inspect-ttc` command that lists all fonts in a collection with their metrics
- Add CLI integration tests (like the original `TestCLISmokeDryRunAndPDF`)

### Code review instructions
- Run `go test ./...` — all 7 package test suites should pass
- Run `go run ./cmd/font-util --help` — should show 4 commands
- Run `go run ./cmd/font-util inspect-font Didot.ttc` — should show Didot metrics
- Run `go run ./cmd/font-util render --font Didot.ttc --text "AV" --out /tmp/test.pdf` — should create PDF

### Technical details

**Commands now available:**
- `ttc2ttf` — Extract TTF files from TTC
- `init-template` — Create starter YAML practice template
- `inspect-font` — Inspect font metrics and shaping (TTC-aware)
- `render` — Render typography practice PDF (TTC-aware, dry-run mode)

**Package count:** 7 (ttc, fontmetrics, spec, shape, layout, renderpdf + cmd/font-util/cmds)
**Total test count:** 15 tests across 7 packages

---

## Step 10: Strip Glazed output from bare commands

The user pointed out that most commands don't need Glazed structured output — only `inspect-font` genuinely benefits from `--output json/table/yaml` since it produces tabular data. The other three commands are file-writing operations that should just print status to stdout.

### Prompt Context

**User prompt (verbatim):** "you don't need glazed structured output for most of these, right?"

**Assistant interpretation:** Remove the Glazed output sections from bare commands (ttc2ttf, init-template, render). Keep it only on inspect-font.

**Inferred user intent:** Simplify the commands — don't carry 60+ Glazed formatting flags on commands that just write files.

### What I did
- Removed `settings.NewGlazedSchema()` and `cli.NewCommandSettingsSection()` from ttc2ttf, init-template, and render
- Changed ttc2ttf to print "FontName -> /path" per font and "Extracted N font(s)" summary
- Changed init-template to just silently write the YAML file
- Changed render to print "Created X (N pages, font: Y)" or JSON on --dry-run
- Kept Glazed output on inspect-font (the one command where it's genuinely useful)
- Extracted shared helpers (loadFont, isTTCFile, splitCSV, firstNonEmpty) into helpers.go
- Lint clean, all tests pass

### Why
The Glazed output pipeline adds ~60 flags (--output, --fields, --filter, --template, --jq, etc.) to every command that includes it. For bare commands that just write files, this is noise. Users don't need `--output json` when extracting TTF files.

### What worked
- The refactor was clean — just removed the section wiring and swapped `gp.AddRow()` for `fmt.Printf()`
- The shared helpers.go file eliminated duplication nicely
- dry-run mode on render now outputs clean JSON to stdout instead of Glazed rows

### What didn't work
- N/A

### What I learned
- The Glazed framework is great for structured-data commands but overkill for file-writing commands. The right pattern is: use Glazed for flag parsing and Cobra integration on all commands, but only add the output section to commands that produce rows.

### What was tricky to build
- Nothing tricky — just a mechanical removal of sections and replacement of row emission with fmt.Printf.

### What warrants a second pair of eyes
- The dry-run JSON output on render — it now dumps the full layout document as JSON. This is useful for debugging but different from the previous Glazed row output. Users who were parsing the old output format would need to adjust.

### What should be done in the future
- N/A

### Code review instructions
- `go run ./cmd/font-util ttc2ttf GillSans.ttc` — should print names and paths, no Glazed flags in --help
- `go run ./cmd/font-util inspect-font Didot.ttc --output json` — should still have Glazed output
- `go run ./cmd/font-util render --font Didot.ttc --text "AV" --out /tmp/test.pdf` — should print one status line
- `make lint && make test` — all clean

### Technical details
**Commit:** 9df66be
**Lines removed:** 199 → 97 net reduction

---

## Step 11: In-memory TTC extraction, --font-index, integration tests

Replaced the temp-file TTC hack with in-memory extraction via `ExtractFontBytes`, added `--font-index` flag to inspect-font and render, and added CLI integration tests.

**Commit (code):** 19dcc1c

### Prompt Context

**User prompt (verbatim):** "go ahead."

**Assistant interpretation:** Implement the next steps from the previous session: --font-index for TTC selection, native TTC support in fontmetrics (in-memory), CLI integration tests.

**Inferred user intent:** Polish the tool by eliminating the temp-file hack and making TTC font selection configurable.

### What I did
- Added `ExtractFontBytes` to `pkg/ttc/writer.go`: reassembles a standalone TTF byte slice in-memory. `ExtractFont` now delegates to `ExtractFontBytes` then writes the bytes.
- Replaced temp-file hack in `loadFont` (helpers.go) with `ExtractFontBytes` + `fontmetrics.LoadBytes`. No more `os.CreateTemp` / `os.Remove` / `tmpFile.Close` dance.
- `loadFont` now accepts a `fontIndex int` parameter (0-based, -1 = first).
- Added `--font-index` flag (default 0) to both `inspect-font` and `render` commands.
- Out-of-range index gives clear error: "font index 99 out of range (TTC has 9 fonts)"
- Added 2 ttc package tests:
  - `TestExtractFontBytesRoundTrip`: extracts Futura-Bold in-memory, validates SFNT version and NumTables
  - `TestExtractFontBytesVersusFile`: byte-for-byte comparison — both methods produce identical output
- Added 4 CLI integration tests in `commands_test.go`:
  - `TestInspectFontOnSingleFont`: loads Go-Regular, checks metrics
  - `TestLoadFontTTCWithIndex`: loads GillSans font 0 ("Gill Sans") and font 1 ("Gill Sans Bold"), verifies names
  - `TestIsTTCFile`: detects TTC vs TTF
  - `TestInitTemplateCreatesFile`: creates and writes a template

### Why
The temp-file pattern was a pragmatic workaround but had edge cases and added I/O overhead. In-memory extraction is cleaner, faster, and eliminates the errcheck noise from unchecked `os.Remove`/`Close` returns.

### What worked
- `ExtractFontBytes` was a clean refactor — just moved the buffer-building logic out of `ExtractFont` into its own function, then had `ExtractFont` call it and write the result.
- The byte-for-byte test proved the two code paths are identical.
- `--font-index 5` on GillSans.ttc correctly loads "Gill Sans SemiBold Italic" (font 5 of 9).

### What didn't work
- N/A

### What I learned
- `fontmetrics.LoadBytes` already existed in the copied fontmetrics package — it takes `[]byte` and calls `sfnt.Parse`. This is exactly what we needed for in-memory TTC support.
- The `--font-index` flag is clean because it's silently ignored for non-TTC files (loadFont just ignores it when the file isn't a TTC).

### What was tricky to build
- Nothing tricky — the refactor was mechanical and well-tested.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Consider an `inspect-ttc` command that lists all fonts in a collection with their names and metrics (no extraction needed, just the parser)
- Consider adding a `--list` flag to `ttc2ttf` that shows font names without extracting

### Code review instructions
- `go test ./...` — all 21 tests should pass (7 packages)
- `go run ./cmd/font-util inspect-font GillSans.ttc --font-index 7 --output table` — should show "Gill Sans Light"
- `go run ./cmd/font-util render --font GillSans.ttc --font-index 1 --text "Bold" --out /tmp/bold.pdf` — should create a PDF with Gill Sans Bold

### Technical details

**Test count:** 21 total (7 packages)
**New tests:** 6 (2 ttc + 4 cmds)
**Commit:** 19dcc1c

---

## Step 12: inspect-ttc, --list, OTC/CFF, native TTC in fontmetrics

Implemented 4 improvements in one pass: new `inspect-ttc` GlazeCommand, `--list` flag on `ttc2ttf`, `.otf` extension for CFF fonts, and native TTC support in fontmetrics via `opentype.ParseCollection`.

**Commit (code):** 3c7cf2e

### Prompt Context

**User prompt (verbatim):** "do 2. 1. 4. 6." (referring to the numbered improvements list from Step 11)

**Assistant interpretation:** Implement: (2) inspect-ttc command, (1) --list flag on ttc2ttf, (4) OTC/CFF extension support, (6) native TTC in fontmetrics.

**Inferred user intent:** Add the four most useful improvements from the backlog.

### What I did
- Created `inspect-ttc` as a GlazeCommand (with full Glazed output support): lists index, name, sfnt_version, sfnt_type, tables, output_ext for each font in a TTC
- Added `--list` flag to `ttc2ttf` (BareCommand): prints font names, types, and table counts without extracting
- Fixed `ExtractAllFonts` in writer.go: uses `.otf` extension for CFF fonts (SFNTVersion = 0x4F54544F/"OTTO"), `.ttf` for TrueType
- Added `LoadFromTTC` to fontmetrics: uses `opentype.ParseCollection` for the sfnt.Font (native TTC support), then uses our `ExtractFontBytes` for raw bytes so `parseOS2` reads the correct font's OS/2 table
- Simplified `loadFont` in helpers.go: just calls `fontmetrics.LoadFromTTC(data, fontIndex)` for TTC files
- Added 2 new tests: `TestLoadFromTTC` (fontmetrics), `TestOtfExtensionForCFF` (ttc)
- Updated README with all 5 commands

### Why
These are the 4 most impactful improvements from the backlog. inspect-ttc and --list are high-visibility quality-of-life features. OTC support is a correctness fix. Native TTC support eliminates the remaining extraction overhead.

### What worked
- `opentype.ParseCollection` + `coll.Font(i)` returns a proper `*sfnt.Font` — exactly what we need
- The two-source approach (ParseCollection for sfnt.Font + ExtractFontBytes for raw bytes) gives correct OS/2 metrics without temp files
- The `--list` flag was trivial — just parse the TTC and print names instead of extracting

### What didn't work
- First attempt at `LoadFromTTC` passed the full TTC data to `Extract()`, which caused `parseOS2` to read the first font's OS/2 table regardless of fontIndex. Fixed by extracting the individual font's bytes first.
- Initial implementation tried to use `opentype.Font` fields directly (Name, UnitsPerEm, GlyphCount, HMetric, VMetric) — but `opentype.Font` is just `sfnt.Font`, which doesn't have those fields. Had to go back to using our own `Extract()`.

### What I learned
- `opentype.Collection.Font(i)` returns `(*sfnt.Font, error)` — same type we already use everywhere. The "native TTC support" is just using this API instead of extracting to bytes and calling `sfnt.Parse`.
- The OS/2 table offset in a TTC is relative to each font's offset table, not the file start. Our `parseOS2` function works on raw bytes starting from offset 0, so it needs the extracted standalone font bytes.

### What was tricky to build
- The `LoadFromTTC` implementation needs both `opentype.ParseCollection` (for the sfnt.Font) AND our own `ttc.Parse` + `ttc.ExtractFontBytes` (for the raw bytes). This means we parse the TTC twice, but it's still fast and avoids temp files.

### What warrants a second pair of eyes
- The double-parse in `LoadFromTTC` — we parse with both `opentype.ParseCollection` and our `ttc.Parse`. Could we get the sfnt.Font from our own parser? Not easily — our parser returns structured metadata, not a live sfnt.Font. The double-parse is acceptable.

### What should be done in the future
- Head table checkSumAdjustment recomputation (open question from design doc)
- Mac Roman encoding fallback in name extraction
- More CLI integration tests

### Code review instructions
- `go run ./cmd/font-util inspect-ttc GillSans.ttc` — should list 9 fonts with types and extensions
- `go run ./cmd/font-util ttc2ttf GillSans.ttc --list` — should list without extracting
- `go run ./cmd/font-util inspect-font GillSans.ttc --font-index 7` — should show "Gill Sans Light" with os2 source
- `make lint && make test` — all clean

### Technical details

**Commands now:** 5 (ttc2ttf, inspect-ttc, inspect-font, init-template, render)
**Test count:** 24 across 8 packages
**Commit:** 3c7cf2e
