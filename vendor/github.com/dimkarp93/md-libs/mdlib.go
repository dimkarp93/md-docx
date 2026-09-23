package mdlib

import (
	"fmt"

	"github.com/dimkarp93/md-libs/filter"
	"github.com/dimkarp93/md-libs/markdown"
)

type Options struct {
	Pages            string
	Heads            string
	HideMatchedHeads bool
}

func Prepare(src string, opts Options) ([]markdown.Block, error) {
	pages := markdown.SplitPages(src)
	maxPage := len(pages) - 1

	var selected []int
	if opts.Pages == "" {
		for i := 1; i <= maxPage; i++ {
			selected = append(selected, i)
		}
	} else {
		req, err := markdown.ParsePageRanges(opts.Pages)
		if err != nil {
			return nil, fmt.Errorf("invalid pages: %w", err)
		}
		for _, idx := range req {
			if idx >= 0 && idx <= maxPage {
				selected = append(selected, idx)
			}
		}
	}

	var filters []filter.HeadFilter
	if opts.Heads != "" {
		var err error
		filters, err = filter.ParseHeads(opts.Heads)
		if err != nil {
			return nil, fmt.Errorf("invalid heads: %w", err)
		}
	}

	var blocks []markdown.Block
	for i, pi := range selected {
		if i > 0 {
			blocks = append(blocks, markdown.Block{Kind: markdown.PageBreak})
		}
		blocks = append(blocks, markdown.Parse(pages[pi])...)
	}

	if filters != nil {
		blocks = filter.ByHeads(blocks, filters, opts.HideMatchedHeads)
	}

	return blocks, nil
}
