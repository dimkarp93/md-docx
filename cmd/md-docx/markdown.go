package main

import (
	"bufio"
	"regexp"
	"strings"
)

type blockKind int

const (
	blockParagraph blockKind = iota
	blockHeading
	blockCode
	blockPageBreak
)

type block struct {
	kind  blockKind
	level int
	text  string
	lines []string
}

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

func parseMarkdown(src string) []block {
	var blocks []block
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	inCode := false
	var codeLines []string
	var paraLines []string

	flushPara := func() {
		if len(paraLines) > 0 {
			blocks = append(blocks, block{kind: blockParagraph, text: strings.Join(paraLines, "\n")})
			paraLines = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			flushPara()
			if inCode {
				blocks = append(blocks, block{kind: blockCode, lines: codeLines})
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
			blocks = append(blocks, block{kind: blockHeading, level: len(m[1]), text: strings.TrimSpace(m[2])})
			continue
		}

		paraLines = append(paraLines, line)
	}

	flushPara()
	if inCode && len(codeLines) > 0 {
		blocks = append(blocks, block{kind: blockCode, lines: codeLines})
	}

	return blocks
}

type run struct {
	text   string
	bold   bool
	italic bool
}

var (
	boldItalicRe  = regexp.MustCompile(`^\*\*\*(.+?)\*\*\*`)
	boldStarRe    = regexp.MustCompile(`^\*\*(.+?)\*\*`)
	boldUnderRe   = regexp.MustCompile(`^__(.+?)__`)
	italicStarRe  = regexp.MustCompile(`^\*(.+?)\*`)
	italicUnderRe = regexp.MustCompile(`^_(.+?)_`)
)

func parseInline(text string, allowUnderscore bool) []run {
	var runs []run
	var plain strings.Builder

	flush := func() {
		if plain.Len() > 0 {
			runs = append(runs, run{text: plain.String()})
			plain.Reset()
		}
	}

	for len(text) > 0 {
		if m := boldItalicRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, run{text: m[1], bold: true, italic: true})
			text = text[len(m[0]):]
			continue
		}
		if m := boldStarRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, run{text: m[1], bold: true})
			text = text[len(m[0]):]
			continue
		}
		if allowUnderscore {
			if m := boldUnderRe.FindStringSubmatch(text); m != nil {
				flush()
				runs = append(runs, run{text: m[1], bold: true})
				text = text[len(m[0]):]
				continue
			}
		}
		if m := italicStarRe.FindStringSubmatch(text); m != nil {
			flush()
			runs = append(runs, run{text: m[1], italic: true})
			text = text[len(m[0]):]
			continue
		}
		if allowUnderscore {
			if m := italicUnderRe.FindStringSubmatch(text); m != nil {
				flush()
				runs = append(runs, run{text: m[1], italic: true})
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

func plainText(text string) string {
	runs := parseInline(text, false)
	var sb strings.Builder
	for _, r := range runs {
		sb.WriteString(r.text)
	}
	return sb.String()
}
