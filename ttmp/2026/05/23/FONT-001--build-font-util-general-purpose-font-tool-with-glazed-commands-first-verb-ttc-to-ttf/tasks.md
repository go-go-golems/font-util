# Tasks

## TODO

- [x] 1.1 Rename Go module from `github.com/go-go-golems/XXX` to `github.com/go-go-golems/font-util` in go.mod
- [x] 1.2 Rename `cmd/XXX/` to `cmd/font-util/` and update main.go package
- [x] 1.3 Add Glazed + Cobra dependencies to go.mod
- [x] 1.4 Create `pkg/ttc/` package directory
- [x] 1.5 Remove empty `pkg/doc.go`
- [x] 1.6 Update .goreleaser.yaml (project_name, binary, main paths, homepage, brew name)
- [x] 1.7 Update Makefile (binary name, install target)
- [x] 1.8 Update AGENT.md (module name, build commands)
- [x] 2.1 Implement TTC binary parser in `pkg/ttc/parser.go` (TTCHeader, FontHeader, TableRecord, Parse, ParseFile)
- [x] 2.2 Implement font name extraction from `name` table in parser
- [x] 2.3 Write parser unit tests in `pkg/ttc/parser_test.go` using Didot.ttc / Futura.ttc / GillSans.ttc
- [x] 3.1 Implement TTF writer in `pkg/ttc/writer.go` (ExtractFont, ExtractAllFonts, calcSearchFields)
- [x] 3.2 Write writer unit tests (round-trip, offset verification, padding)
- [x] 4.1 Implement Ttc2TtfCommand Glazed command in `cmd/font-util/cmds/ttc2ttf.go`
- [x] 4.2 Wire root Cobra command in `cmd/font-util/main.go` with logging, help, and ttc2ttf
- [x] 4.3 Integration test: run `font-util ttc2ttf` on real TTC files, verify output
- [x] 5.1 Run `make lint` and fix any issues
- [x] 5.2 Run `make test` and ensure all tests pass
- [x] 5.3 Update README.md with usage examples
- [x] 5.4 Final commit and push
- [ ] Copy typo-copy-generator packages (fontmetrics, spec, shape, layout, renderpdf) into pkg/
- [ ] Fix import paths from wesen/typo-copy-generator to go-go-golems/font-util
- [ ] Add dependencies (fpdf, typesetting, x/image, yaml.v3)
- [ ] Create Glazed commands: init-template, inspect-font, render
- [ ] Wire new commands into root and test
- [ ] Copy examples and run lint + test
