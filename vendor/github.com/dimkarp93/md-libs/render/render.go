package render

import "github.com/dimkarp93/md-libs/markdown"

type Renderer interface {
	Heading(level int, text string)
	Paragraph(text string)
	Code(lines []string)
	PageBreak()
}

func Blocks(r Renderer, blocks []markdown.Block) {
	for _, b := range blocks {
		switch b.Kind {
		case markdown.Heading:
			r.Heading(b.Level, b.Text)
		case markdown.Code:
			r.Code(b.Lines)
		case markdown.PageBreak:
			r.PageBreak()
		default:
			r.Paragraph(b.Text)
		}
	}
}

func HeadingSizePt(level int) float64 {
	switch level {
	case 1:
		return 18
	case 2:
		return 16
	case 3:
		return 14
	case 4:
		return 13
	case 5:
		return 12
	default:
		return 11
	}
}
