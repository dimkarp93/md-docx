package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dimkarp93/install-libs/buildinfo"
	mdlib "github.com/dimkarp93/md-libs"
)

var (
	version  string
	origin   string
	upstream string
	commit   string
	channel  string
)

func build() buildinfo.Info {
	return buildinfo.Info{
		Version:  version,
		Origin:   origin,
		Upstream: upstream,
		Commit:   commit,
		Channel:  channel,
	}
}

func main() {
	inPath := flag.String("in", "", "input markdown file (default: stdin)")
	outPath := flag.String("out", "", "output docx file (default: stdout)")
	pagesFlag := flag.String("pages", "", "pages to include, e.g. 1,3-5 (default: all pages except 0)")
	headsFlag := flag.String("heads", "", "heading filters, e.g. h2:result,h3:resume")
	rootHeadHide := flag.Bool("root-head-hide", false, "hide headings matched by --heads, keep their content")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit")
	showOrigin := flag.Bool("origin", false, "print the repository the binary was built from and exit")
	showBuildInfo := flag.Bool("buildinfo", false, "print build metadata and exit")
	flag.Parse()

	switch {
	case *showVersion:
		fmt.Println(build().VersionString())
		os.Exit(0)
	case *showOrigin:
		fmt.Println(build().OriginString())
		os.Exit(0)
	case *showBuildInfo:
		build().Print()
		os.Exit(0)
	}

	var data []byte
	var err error
	if *inPath == "" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(*inPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	blocks, err := mdlib.Prepare(string(data), mdlib.Options{
		Pages:            *pagesFlag,
		Heads:            *headsFlag,
		HideMatchedHeads: *rootHeadHide,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	body := renderBody(blocks)

	out := io.Writer(os.Stdout)
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", *outPath, err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	if err := writeDocx(out, body); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
