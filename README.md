# md-docx

A simple utility that converts a Markdown file into `.docx`. No external dependencies — it uses only the Go standard library, and the `.docx` is assembled directly as an OOXML archive.

## Supported syntax

1. Headings: `#` … `######` (levels 1–6).
2. Bold: `**text**` or `__text__`.
3. Italic: `*text*` or `_text_`.
4. Bold italic: `***text***`.
5. Code blocks: `` ```...``` `` — rendered in a monospace font with a grey fill.

Everything else is treated as a regular text paragraph.

## Build

Requires Go 1.26+.

```bash
make build
```

The binary appears at `./bin/md-docx`.

## Usage

```bash
./bin/md-docx --in input.md --out output.docx
```

- `--in` (optional) — path to the source Markdown file. If omitted, reads from stdin.
- `--out` (optional) — path to the output `.docx`. If omitted, writes to stdout.
- `--pages` (optional) — which pages to include, for example `1,3-5`. By default — every page except page 0.
- `--heads` (optional) — a heading filter, for example `h2:result,h3:resume`. By default there is no filtering.

stdin/stdout are supported as well:

```bash
cat input.md | ./bin/md-docx > output.docx
```

### Frontmatter and pages

If the file starts with a `---` line, everything up to the next `---` line is treated as frontmatter (metadata — YAML/TOML/JSON) and is handled as page 0. It does not make it into the result unless it is requested explicitly through `--pages=0`.

Every other line consisting of exactly `---` is a page break — they split the document into pages 1, 2, 3, … in order. A `---` inside a code block (`` ``` ``) is not a page break.

By default the `.docx` contains every page except page 0. To pick specific pages:

```bash
./bin/md-docx --in input.md --pages=1,3-5 --out output.docx
```

### Heading filter

`--heads` limits the output to the content under the given headings (and their nested content). The format is a comma-separated list where each item looks like `h<level>:<name>`, or just `<name>` without a level — in which case a match is looked for at any level:

```bash
./bin/md-docx --in input.md --heads=h2:result,h3:resume,summary --out output.docx
```

This means: keep only what is inside the level-2 heading "result", inside the level-3 heading "resume", or inside a heading "summary" at any level (the comparison is exact, ignoring case and markdown markup in the heading text). A heading is "nested" inside another if it comes after it and has a deeper level — the nesting ends at the first heading of the same or a higher level.

Heading nesting is computed across the selected document as a whole (after `--pages` is applied), not per page — that is, if a matched heading is on one page while its child headings come after a page break (`---`) on the next one, they still count as its descendants and make it into the output.

Between every two adjacent selected pages a real page break is added to the resulting `.docx` (Word/LibreOffice will start the next page on a new physical page of the document). The break is inserted regardless of `--heads` — if the content on both sides of the boundary passed the filter, the break between them is preserved.

`--root-head-hide` — when passed together with `--heads`, hides the matched headings themselves but keeps their content (including nested headings):

```bash
./bin/md-docx --in input.md --heads=result --root-head-hide --out output.docx
```

If `--heads` is not given, there is no filtering — all the content of the selected pages is emitted.

## Clean

```bash
make clean
```

Removes the built binary.

## Development

Markdown parsing, heading filtering and the rendering contract live in the shared library [md-libs](https://github.com/dimkarp93/md-libs) — the same library is used by [md-pdf](https://github.com/dimkarp93/md-pdf). What remains in this repository is only the OOXML rendering and the flag parsing.

The dependencies, md-libs included, are vendored: `make build` and the tests take them from `vendor/` only (the Makefile exports `GOWORK=off` and `GOFLAGS=-mod=vendor`), nothing has to be set up and no network is needed:

```bash
make build
```

To change the library and the CLI at the same time, publish a new version of md-libs first and then update the dependency here (see below).

### Tests

```sh
make test                  # renderer tests and end-to-end CLI tests
make test-v                # the same, with test names
make test-run T=TestCLIVersion
make cover                 # coverage
make check                 # vet + test
```

The shared core (parsing, filters, pipeline) is tested in [md-libs](https://github.com/dimkarp93/md-libs); only what is specific to this CLI is checked here. `make test-all` in md-libs runs everything at once.

Upgrading to a new version of md-libs:

```bash
GOWORK=off go get github.com/dimkarp93/md-libs@v0.1.1
make vendor
```

`make vendor-check` verifies that `vendor/` matches `go.mod`.

## License

[MIT](LICENSE)
