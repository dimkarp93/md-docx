package markdown

import (
	"bufio"
	"regexp"
	"strings"
)

type BlockKind int

const (
	Paragraph BlockKind = iota
	Heading
	Code
	PageBreak
)

type Block struct {
	Kind  BlockKind
	Level int
	Text  string
	Lines []string
}

type Run struct {
	Text   string
	Bold   bool
	Italic bool
}

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

func Parse(src string) []Block {
	var blocks []Block
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	inCode := false
	var codeLines []string
	var paraLines []string

	flushPara := func() {
		if len(paraLines) > 0 {
			blocks = append(blocks, Block{Kind: Paragraph, Text: strings.Join(paraLines, "\n")})
			paraLines = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			flushPara()
			if inCode {
				blocks = append(blocks, Block{Kind: Code, Lines: codeLines})
				codeLines = nil
				inCode = false
			} else {
				inCode = true
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		if strings.TrimSpace(line) == "" {
			flushPara()
			continue
		}

		if m := headingRe.FindStringSubmatch(line); m != nil {
			flushPara()
			blocks = append(blocks, Block{Kind: Heading, Level: len(m[1]), Text: strings.TrimSpace(m[2])})
			continue
		}

		paraLines = append(paraLines, line)
	}

	flushPara()
	if inCode && len(codeLines) > 0 {
		blocks = append(blocks, Block{Kind: Code, Lines: codeLines})
	}

	return blocks
}

var (
	boldItalicRe  = regexp.MustCompile(`^\*\*\*(.+?)\*\*\*`)
	boldStarRe    = regexp.MustCompile(`^\*\*(.+?)\*\*`)
	boldUnderRe   = regexp.MustCompile(`^__(.+?)__`)
	italicStarRe  = regexp.MustCompile(`^\*(.+?)\*`)
	italicUnderRe = regexp.MustCompile(`^_(.+?)_`)
)

func ParseInline(text string, allowUnderscore bool) []Run {
	var runs []Run
	var plain strings.Builder

	flush := func() {
		if plain.Len() > 0 {
			runs = append(runs, Run{Text: plain.String()})
			plain.Reset()
		}
	}

	for len(text) > 0 {
		if m := boldItalicRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, Run{Text: m[1], Bold: true, Italic: true})
			text = text[len(m[0]):]
			continue
		}
		if m := boldStarRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, Run{Text: m[1], Bold: true})
			text = text[len(m[0]):]
			continue
		}
		if allowUnderscore {
			if m := boldUnderRe.FindStringSubmatch(text); m != nil {
				flush()
				runs = append(runs, Run{Text: m[1], Bold: true})
				text = text[len(m[0]):]
				continue
			}
		}
		if m := italicStarRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, Run{Text: m[1], Italic: true})
			text = text[len(m[0]):]
			continue
		}
		if allowUnderscore {
			if m := italicUnderRe.FindStringSubmatch(text); m != nil {
				flush()
				runs = append(runs, Run{Text: m[1], Italic: true})
				text = text[len(m[0]):]
				continue
			}
		}
		plain.WriteByte(text[0])
		text = text[1:]
	}
	flush()
	return runs
}

func PlainText(text string) string {
	runs := ParseInline(text, false)
	var sb strings.Builder
	for _, r := range runs {
		sb.WriteString(r.Text)
	}
	return sb.String()
}
