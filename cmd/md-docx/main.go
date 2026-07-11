package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var version = "dev"

func main() {
	inPath := flag.String("in", "", "input markdown file (default: stdin)")
	outPath := flag.String("out", "", "output docx file (default: stdout)")
	pagesFlag := flag.String("pages", "", "pages to include, e.g. 1,3-5 (default: all pages except 0)")
	headsFlag := flag.String("heads", "", "heading filters, e.g. h2:result,h3:resume")
	rootHeadHide := flag.Bool("root-head-hide", false, "hide headings matched by --heads, keep their content")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
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

	pagesRaw := splitPages(string(data))
	maxPage := len(pagesRaw) - 1

	var selected []int
	if *pagesFlag == "" {
		for i := 1; i <= maxPage; i++ {
			selected = append(selected, i)
		}
	} else {
		req, err := parsePageRanges(*pagesFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --pages: %v\n", err)
			os.Exit(1)
		}
		for _, idx := range req {
			if idx >= 0 && idx <= maxPage {
				selected = append(selected, idx)
			}
		}
	}

	var filters []headFilter
	if *headsFlag != "" {
		filters, err = parseHeadFilters(*headsFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --heads: %v\n", err)
			os.Exit(1)
		}
	}

	var allBlocks []block
	for i, pi := range selected {
		if i > 0 {
			allBlocks = append(allBlocks, block{kind: blockPageBreak})
		}
		allBlocks = append(allBlocks, parseMarkdown(pagesRaw[pi])...)
	}
	if filters != nil {
		allBlocks = filterByHeads(allBlocks, filters, *rootHeadHide)
	}

	body := renderBody(allBlocks)

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
