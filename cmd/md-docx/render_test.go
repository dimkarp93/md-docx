package main

import (
	"strings"
	"testing"

	"github.com/dimkarp93/md-libs/markdown"
)

func TestXMLEscape(t *testing.T) {
	got := xmlEscape(`a & b < c > d "e"`)
	want := `a &amp; b &lt; c &gt; d &quot;e&quot;`
	if got != want {
		t.Errorf("xmlEscape() = %q, want %q", got, want)
	}
}

func TestParagraphRendersInlineMarkup(t *testing.T) {
	d := &docxRenderer{}
	d.Paragraph("plain **bold** _italic_ ***both***")
	got := d.Body()

	for _, want := range []string{
		`<w:p><w:pPr><w:spacing w:after="120"/></w:pPr>`,
		`<w:r><w:t xml:space="preserve">plain </w:t></w:r>`,
		`<w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">bold</w:t></w:r>`,
		`<w:r><w:rPr><w:i/></w:rPr><w:t xml:space="preserve">italic</w:t></w:r>`,
		`<w:r><w:rPr><w:b/><w:i/></w:rPr><w:t xml:space="preserve">both</w:t></w:r>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Paragraph() missing %q\ngot: %s", want, got)
		}
	}
}

func TestParagraphKeepsSoftLineBreaks(t *testing.T) {
	d := &docxRenderer{}
	d.Paragraph("first\nsecond")
	got := d.Body()

	if n := strings.Count(got, `<w:r><w:br/></w:r>`); n != 1 {
		t.Errorf("Paragraph() has %d line breaks, want 1\ngot: %s", n, got)
	}
	if n := strings.Count(got, "<w:p>"); n != 1 {
		t.Errorf("Paragraph() produced %d paragraphs, want 1 (soft break, not a new block)", n)
	}
}

func TestParagraphEscapesSpecialCharacters(t *testing.T) {
	d := &docxRenderer{}
	d.Paragraph("a < b & c")
	if got := d.Body(); !strings.Contains(got, `a &lt; b &amp; c`) {
		t.Errorf("Paragraph() did not escape XML: %s", got)
	}
}

func TestHeadingSizesMatchTheSharedScale(t *testing.T) {
	want := map[int]string{
		1: `<w:sz w:val="36"/><w:szCs w:val="36"/>`,
		2: `<w:sz w:val="32"/><w:szCs w:val="32"/>`,
		3: `<w:sz w:val="28"/><w:szCs w:val="28"/>`,
		4: `<w:sz w:val="26"/><w:szCs w:val="26"/>`,
		5: `<w:sz w:val="24"/><w:szCs w:val="24"/>`,
		6: `<w:sz w:val="22"/><w:szCs w:val="22"/>`,
	}
	for level, size := range want {
		d := &docxRenderer{}
		d.Heading(level, "Title")
		got := d.Body()
		if !strings.Contains(got, size) {
			t.Errorf("Heading(%d) missing %q\ngot: %s", level, size, got)
		}
		if !strings.Contains(got, "<w:b/>") {
			t.Errorf("Heading(%d) is not bold\ngot: %s", level, got)
		}
	}
}

func TestHeadingSizeIsHalfPointsOfThePtScale(t *testing.T) {
	for level := 1; level <= 6; level++ {
		if got, want := headingSizeHalfPoints(level), []int{0, 36, 32, 28, 26, 24, 22}[level]; got != want {
			t.Errorf("headingSizeHalfPoints(%d) = %d, want %d", level, got, want)
		}
	}
}

func TestHeadingIgnoresUnderscoreMarkup(t *testing.T) {
	d := &docxRenderer{}
	d.Heading(2, "snake_case_name")
	if got := d.Body(); !strings.Contains(got, `<w:t xml:space="preserve">snake_case_name</w:t>`) {
		t.Errorf("Heading() must not treat underscores as italics\ngot: %s", got)
	}
}

func TestCodeRendersOneParagraphPerLine(t *testing.T) {
	d := &docxRenderer{}
	d.Code([]string{"line one", "", "a < b"})
	got := d.Body()

	if n := strings.Count(got, "<w:p>"); n != 3 {
		t.Errorf("Code() produced %d paragraphs, want 3", n)
	}
	if !strings.Contains(got, `<w:t xml:space="preserve"> </w:t>`) {
		t.Errorf("Code() must render an empty line as a space\ngot: %s", got)
	}
	if !strings.Contains(got, `a &lt; b`) {
		t.Errorf("Code() did not escape XML\ngot: %s", got)
	}
	if !strings.Contains(got, `w:ascii="Consolas"`) {
		t.Errorf("Code() is not monospaced\ngot: %s", got)
	}
}

func TestPageBreak(t *testing.T) {
	d := &docxRenderer{}
	d.PageBreak()
	if got, want := d.Body(), `<w:p><w:r><w:br w:type="page"/></w:r></w:p>`; got != want {
		t.Errorf("PageBreak() = %q, want %q", got, want)
	}
}

func TestRenderBodyDispatchesAllBlockKinds(t *testing.T) {
	body := renderBody([]markdown.Block{
		{Kind: markdown.Heading, Level: 1, Text: "Title"},
		{Kind: markdown.Paragraph, Text: "text"},
		{Kind: markdown.Code, Lines: []string{"code"}},
		{Kind: markdown.PageBreak},
	})

	for _, want := range []string{
		`<w:sz w:val="36"/>`,
		`<w:t xml:space="preserve">text</w:t>`,
		`w:ascii="Consolas"`,
		`<w:br w:type="page"/>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("renderBody() missing %q\ngot: %s", want, body)
		}
	}
}

func TestRenderBodyOnEmptyInput(t *testing.T) {
	if got := renderBody(nil); got != "" {
		t.Errorf("renderBody(nil) = %q, want empty", got)
	}
}
