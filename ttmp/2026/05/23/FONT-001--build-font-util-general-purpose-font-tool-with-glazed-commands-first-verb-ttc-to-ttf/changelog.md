# Changelog

## 2026-05-23

- Initial workspace created


## 2026-05-23

Created ticket FONT-001, wrote comprehensive implementation guide (design-doc), started diary, added tasks for 5 implementation phases, related 5 key repo files

### Related Files

- /home/manuel/code/wesen/go-go-golems/font-util/ttmp/2026/05/23/FONT-001--build-font-util-general-purpose-font-tool-with-glazed-commands-first-verb-ttc-to-ttf/design-doc/01-implementation-guide-font-util-with-glazed-commands.md — Main deliverable - intern-ready implementation guide
- /home/manuel/code/wesen/go-go-golems/font-util/ttmp/2026/05/23/FONT-001--build-font-util-general-purpose-font-tool-with-glazed-commands-first-verb-ttc-to-ttf/reference/01-diary.md — Diary of investigation and design work


## 2026-05-23

Validated ticket with docmgr doctor (all checks pass), uploaded design-doc + diary bundle to reMarkable at /ai/2026/05/23/FONT-001


## 2026-05-23

Completed full implementation: TTC parser, TTF writer, ttc2ttf command, tests passing on 3 real TTC files, lint clean, README updated. Commits: ff638bb, 9907509

### Related Files

- /home/manuel/code/wesen/go-go-golems/font-util/README.md — Usage examples and output format docs
- /home/manuel/code/wesen/go-go-golems/font-util/cmd/font-util/cmds/ttc2ttf.go — Glazed command wired to parser/writer
- /home/manuel/code/wesen/go-go-golems/font-util/pkg/ttc/parser.go — TTC binary parser with name extraction
- /home/manuel/code/wesen/go-go-golems/font-util/pkg/ttc/parser_test.go — Tests with Didot/Futura/GillSans TTC files
- /home/manuel/code/wesen/go-go-golems/font-util/pkg/ttc/writer.go — TTF extraction writer with offset recalculation


## 2026-05-23

Integrated typo-copy-generator: 5 packages copied, 3 Glazed commands created (init-template, inspect-font, render), TTC support added, lint clean, all tests pass. Commits: 41d915b, fc16756


## 2026-05-23

Added in-memory TTC extraction (ExtractFontBytes), --font-index flag for inspect-font and render, 6 new tests (21 total). Commit: 19dcc1c


## 2026-05-23

Added inspect-ttc command, --list flag on ttc2ttf, OTC/CFF extension support (.otf), native TTC in fontmetrics via opentype.ParseCollection. Commit: 3c7cf2e

