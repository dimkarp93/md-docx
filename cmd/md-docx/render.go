package main

import (
	"fmt"
	"strings"
)

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}

func renderRun(r run) string {
	var rPr strings.Builder
	if r.bold {
		rPr.WriteString(`<w:b/>`)
	}
	if r.italic {
		rPr.WriteString(`<w:i/>`)
	}
	rPrXML := ""
	if rPr.Len() > 0 {
		rPrXML = "<w:rPr>" + rPr.String() + "</w:rPr>"
	}
	return fmt.Sprintf(`<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>`, rPrXML, xmlEscape(r.text))
}

func renderParagraph(text string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	sb.WriteString(`<w:p><w:pPr><w:spacing w:after="120"/></w:pPr>`)
	for i, line := range lines {
		if i > 0 {
			sb.WriteString(`<w:r><w:br/></w:r>`)
		}
		for _, r := range parseInline(line, true) {
			sb.WriteString(renderRun(r))
		}
	}
	sb.WriteString("</w:p>")
	return sb.String()
}

func headingSizeHalfPoints(level int) int {
	switch level {
	case 1:
		return 36 // 18pt
	case 2:
		return 32 // 16pt
	case 3:
		return 28 // 14pt
	case 4:
		return 26 // 13pt
	case 5:
		return 24 // 12pt
	default:
		return 22 // 11pt
	}
}

func renderHeading(level int, text string) string {
	size := headingSizeHalfPoints(level)
	runs := parseInline(text, false)
	var sb strings.Builder
	sb.WriteString(`<w:p><w:pPr><w:spacing w:before="240" w:after="120"/></w:pPr>`)
	for _, r := range runs {
		var rPr strings.Builder
		rPr.WriteString("<w:b/>")
		if r.italic {
			rPr.WriteString("<w:i/>")
		}
		rPr.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, size, size))
		sb.WriteString(fmt.Sprintf(`<w:r><w:rPr>%s</w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`, rPr.String(), xmlEscape(r.text)))
	}
	sb.WriteString("</w:p>")
	return sb.String()
}

func renderCodeLine(line string) string {
	text := line
	if text == "" {
		text = " "
	}
	return fmt.Sprintf(
		`<w:p><w:pPr><w:shd w:val="clear" w:color="auto" w:fill="F2F2F2"/><w:spacing w:before="0" w:after="0"/></w:pPr>`+
			`<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas" w:cs="Consolas"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>`+
			`<w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
		xmlEscape(text),
	)
}

func renderPageBreak() string {
	return `<w:p><w:r><w:br w:type="page"/></w:r></w:p>`
}

func renderBody(blocks []block) string {
	var sb strings.Builder
	for _, b := range blocks {
		switch b.kind {
		case blockHeading:
			sb.WriteString(renderHeading(b.level, b.text))
		case blockCode:
			for _, line := range b.lines {
				sb.WriteString(renderCodeLine(line))
			}
		case blockPageBreak:
			sb.WriteString(renderPageBreak())
		default:
			sb.WriteString(renderParagraph(b.text))
		}
	}
	return sb.String()
}
