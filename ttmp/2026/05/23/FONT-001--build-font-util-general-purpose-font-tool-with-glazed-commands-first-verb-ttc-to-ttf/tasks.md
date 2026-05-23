# Tasks

## TODO

- [ ] 1.1 Rename Go module from `github.com/go-go-golems/XXX` to `github.com/go-go-golems/font-util` in go.mod
- [ ] 1.2 Rename `cmd/XXX/` to `cmd/font-util/` and update main.go package
- [ ] 1.3 Add Glazed + Cobra dependencies to go.mod
- [ ] 1.4 Create `pkg/ttc/` package directory
- [ ] 1.5 Remove empty `pkg/doc.go`
- [ ] 1.6 Update .goreleaser.yaml (project_name, binary, main paths, homepage, brew name)
- [ ] 1.7 Update Makefile (binary name, install target)
- [ ] 1.8 Update AGENT.md (module name, build commands)
- [ ] 2.1 Implement TTC binary parser in `pkg/ttc/parser.go` (TTCHeader, FontHeader, TableRecord, Parse, ParseFile)
- [ ] 2.2 Implement font name extraction from `name` table in parser
- [ ] 2.3 Write parser unit tests in `pkg/ttc/parser_test.go` using Didot.ttc / Futura.ttc / GillSans.ttc
- [ ] 3.1 Implement TTF writer in `pkg/ttc/writer.go` (ExtractFont, ExtractAllFonts, calcSearchFields)
- [ ] 3.2 Write writer unit tests (round-trip, offset verification, padding)
- [ ] 4.1 Implement Ttc2TtfCommand Glazed command in `cmd/font-util/cmds/ttc2ttf.go`
- [ ] 4.2 Wire root Cobra command in `cmd/font-util/main.go` with logging, help, and ttc2ttf
- [ ] 4.3 Integration test: run `font-util ttc2ttf` on real TTC files, verify output
- [ ] 5.1 Run `make lint` and fix any issues
- [ ] 5.2 Run `make test` and ensure all tests pass
- [ ] 5.3 Update README.md with usage examples
- [ ] 5.4 Final commit and push
