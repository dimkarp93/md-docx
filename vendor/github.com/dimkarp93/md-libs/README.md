# md-libs

Shared core for markdown converters: parsing, heading-based filtering, the page pipeline and the rendering contract. Used by [md-docx](https://github.com/dimkarp93/md-docx) and [md-pdf](https://github.com/dimkarp93/md-pdf).

The library has no external dependencies and contains no concrete output backends — those are implemented by its consumers.

## Installation

```sh
go get github.com/dimkarp93/md-libs
```

## Packages

### `mdlib` — the pipeline

The full path from raw markdown to a filtered list of blocks.

```go
blocks, err := mdlib.Prepare(src, mdlib.Options{
    Pages:            "1,3-5", // empty — every page except page zero
    Heads:            "h2:result,h3:resume",
    HideMatchedHeads: true,
})
```

### `markdown` — parsing

```go
type Block struct {
    Kind  BlockKind // Paragraph | Heading | Code | PageBreak
    Level int
    Text  string
    Lines []string
}

func Parse(src string) []Block
func ParseInline(text string, allowUnderscore bool) []Run
func PlainText(text string) string
func SplitPages(src string) []string
func ParsePageRanges(spec string) ([]int, error)
```

A document is split into pages on `---` lines (a separator inside a code fence is ignored). If the file starts with a closed `---...---` block, its contents become **page zero** (front matter) and are excluded from the output by default.

### `filter` — selecting by headings

```go
filters, err := filter.ParseHeads("h2:result,resume")
blocks = filter.ByHeads(blocks, filters, hideMatched)
```

The `h<N>:<name>` spec binds a filter to a heading level; a bare name matches any level. Matching is case-insensitive and done against the "plain" text, so `**Result**` is matched by the `h2:result` filter. The output keeps the heading and all of its content up to the next heading of the same or a higher level. `hideMatched` drops the heading itself, keeping its content.

### `render` — the rendering contract

```go
type Renderer interface {
    Heading(level int, text string)
    Paragraph(text string)
    Code(lines []string)
    PageBreak()
}

func Blocks(r Renderer, blocks []markdown.Block)
func HeadingSizePt(level int) float64
```

A backend implements the four methods and `render.Blocks` dispatches the list of blocks to them. `HeadingSizePt` is the single scale of heading sizes (18/16/14/13/12/11 pt), so that docx and pdf do not drift apart.

## Versioning

The version lives in `versions.txt`. A push to `main` with a new version triggers the workflow, which runs the tests and creates a release tagged `vX.Y.Z` — exactly the format a Go module needs.

```sh
make bump-patch   # or bump-minor / bump-major
```

While the API is not settled, the series stays at `v0.x`: within the zero major version breaking changes are allowed without a `/v2` suffix in the module path.

## Development

```sh
make build                        # compile every package
make test                         # tests
make test-v                       # the same, with test names
make test-run T=TestParseInline   # a single test or a mask
make cover                        # coverage + total percentage
make vet
make check                        # vet + test
```

`make test-all` additionally runs the tests in `../md-docx` and `../md-pdf` if they are cloned next to this repository — a single command checks both the library and both of its consumers.

What is tested where:

- **here** — parsing (`markdown`), filtering (`filter`), the pipeline (`mdlib`) and the rendering contract (`render`): everything shared by all consumers;
- **in md-docx and md-pdf** — the concrete renderers (Word XML, fpdf output) and the end-to-end CLI tests: everything specific to each of them.

md-docx and md-pdf vendor this library: to use a change in them, publish a new version with a `vX.Y.Z` tag and run `go get` + `make vendor` in the consumer.

## Package for manual transfer

`make pack` builds the library as a **file-based module proxy** — the same format `proxy.golang.org` serves:

```sh
make pack
# dist/md-libs-0.1.1-proxy.tar.gz
```

Inside is `github.com/dimkarp93/md-libs/@v/` with the files `v0.1.1.zip`, `v0.1.1.mod`, `v0.1.1.info` and `list`. This makes it possible to build a dependent project against a plain `require github.com/dimkarp93/md-libs v0.1.1` — **without `replace` and without editing `go.mod`**:

```sh
tar -xzf md-libs-0.1.1-proxy.tar.gz -C /opt
cd ../md-docx
GOFLAGS=-mod=mod GOPROXY=file:///opt/proxy GOSUMDB=off go build ./...
```

That is how a project builds when it has no other external dependencies (md-docx) or when all of them are already in the module cache. If the remaining dependencies have to be fetched from the network (md-pdf depends on `fpdf`), add the network to the proxy chain and disable sumdb **selectively**:

```sh
GOFLAGS=-mod=mod \
GOPROXY=file:///opt/proxy,https://proxy.golang.org,direct \
GONOSUMDB='github.com/dimkarp93/*' \
go build ./...
```

- `GOSUMDB=off` / `GONOSUMDB` is mandatory while the module is not published: otherwise Go goes to `sum.golang.org` and gets a 404. A global `GOSUMDB=off` breaks verification of the other modules downloaded from the network, so a mixed scenario needs `GONOSUMDB` specifically.
- **Do not use `GOPRIVATE`**: it also sets `GONOPROXY`, and Go will fetch md-libs straight from GitHub, bypassing the file proxy.
- `GOFLAGS=-mod=mod` is needed so that Go writes `go.sum`; after that a normal build works without it.
- If the project has a workspace enabled (`go.work`), add `GOWORK=off` — otherwise it overrides the proxy.

`pack` produces a self-contained artifact with a specific version — for a machine without access to the repository.

The package is built from the working tree as is, so run `make pack` on a clean tree: the contents of the zip determine the checksum in `go.sum`, and it has to match the one GitHub computes later for the `vX.Y.Z` tag.

One external resource is still needed: `go.mod` declares `go 1.26.1`, and Go downloads such a toolchain as a module. On a machine without network access it has to be installed already (or build with `GOTOOLCHAIN=local` using a recent enough Go).

## License

[MIT](LICENSE)
