package main

import (
	"fmt"
	"strings"

	"github.com/dimkarp93/md-libs/markdown"
	"github.com/dimkarp93/md-libs/render"
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

type docxRenderer struct {
	sb strings.Builder
}

func (d *docxRenderer) Body() string {
	return d.sb.String()
}

func renderRun(r markdown.Run) string {
	var rPr strings.Builder
	if r.Bold {
		rPr.WriteString(`<w:b/>`)
	}
	if r.Italic {
		rPr.WriteString(`<w:i/>`)
	}
	rPrXML := ""
	if rPr.Len() > 0 {
		rPrXML = "<w:rPr>" + rPr.String() + "</w:rPr>"
	}
	return fmt.Sprintf(`<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>`, rPrXML, xmlEscape(r.Text))
}

func (d *docxRenderer) Paragraph(text string) {
	lines := strings.Split(text, "\n")
	d.sb.WriteString(`<w:p><w:pPr><w:spacing w:after="120"/></w:pPr>`)
	for i, line := range lines {
		if i > 0 {
			d.sb.WriteString(`<w:r><w:br/></w:r>`)
		}
		for _, r := range markdown.ParseInline(line, true) {
			d.sb.WriteString(renderRun(r))
		}
	}
	d.sb.WriteString("</w:p>")
}

func headingSizeHalfPoints(level int) int {
	return int(render.HeadingSizePt(level) * 2)
}

func (d *docxRenderer) Heading(level int, text string) {
	size := headingSizeHalfPoints(level)
	runs := markdown.ParseInline(text, false)
	d.sb.WriteString(`<w:p><w:pPr><w:spacing w:before="240" w:after="120"/></w:pPr>`)
	for _, r := range runs {
		var rPr strings.Builder
		rPr.WriteString("<w:b/>")
		if r.Italic {
			rPr.WriteString("<w:i/>")
		}
		rPr.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, size, size))
		d.sb.WriteString(fmt.Sprintf(`<w:r><w:rPr>%s</w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`, rPr.String(), xmlEscape(r.Text)))
	}
	d.sb.WriteString("</w:p>")
}

func (d *docxRenderer) Code(lines []string) {
	for _, line := range lines {
		text := line
		if text == "" {
			text = " "
		}
		d.sb.WriteString(fmt.Sprintf(
			`<w:p><w:pPr><w:shd w:val="clear" w:color="auto" w:fill="F2F2F2"/><w:spacing w:before="0" w:after="0"/></w:pPr>`+
				`<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas" w:cs="Consolas"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>`+
				`<w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
			xmlEscape(text),
		))
	}
}

func (d *docxRenderer) PageBreak() {
	d.sb.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
}

func renderBody(blocks []markdown.Block) string {
	d := &docxRenderer{}
	render.Blocks(d, blocks)
	return d.Body()
}
