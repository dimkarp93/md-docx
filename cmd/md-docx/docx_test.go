package main

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	mdlib "github.com/dimkarp93/md-libs"
)

func readDocx(t *testing.T, data []byte) map[string]string {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("output is not a valid zip: %v", err)
	}

	files := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		files[f.Name] = string(content)
	}
	return files
}

func TestWriteDocxProducesRequiredParts(t *testing.T) {
	var buf bytes.Buffer
	if err := writeDocx(&buf, "<w:p/>"); err != nil {
		t.Fatalf("writeDocx() unexpected error: %v", err)
	}

	files := readDocx(t, buf.Bytes())
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"} {
		if _, ok := files[name]; !ok {
			t.Errorf("docx is missing %s", name)
		}
	}

	doc := files["word/document.xml"]
	if !strings.Contains(doc, "<w:p/>") {
		t.Errorf("word/document.xml does not contain the body: %s", doc)
	}
	if !strings.Contains(doc, "<w:sectPr>") {
		t.Errorf("word/document.xml is missing page setup: %s", doc)
	}
}

func TestMarkdownToDocxEndToEnd(t *testing.T) {
	src := "---\nfront: matter\n---\n# Title\n\nHello **world**\n\n---\n\n## Result\n\nUnder result\n"

	blocks, err := mdlib.Prepare(src, mdlib.Options{})
	if err != nil {
		t.Fatalf("Prepare() unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := writeDocx(&buf, renderBody(blocks)); err != nil {
		t.Fatalf("writeDocx() unexpected error: %v", err)
	}

	doc := readDocx(t, buf.Bytes())["word/document.xml"]

	for _, want := range []string{"Title", "Hello ", "world", "Result", "Under result"} {
		if !strings.Contains(doc, want) {
			t.Errorf("document.xml missing %q", want)
		}
	}
	if strings.Contains(doc, "front: matter") {
		t.Error("front matter (page 0) must not be rendered by default")
	}
	if n := strings.Count(doc, `<w:br w:type="page"/>`); n != 1 {
		t.Errorf("document.xml has %d page breaks, want 1 between the two pages", n)
	}
}

func TestPageAndHeadFiltersReachTheDocument(t *testing.T) {
	src := "# Intro\n\nintro text\n\n---\n\n## Result\n\nkeep me\n\n## Other\n\ndrop me\n"

	blocks, err := mdlib.Prepare(src, mdlib.Options{
		Pages:            "2",
		Heads:            "h2:result",
		HideMatchedHeads: true,
	})
	if err != nil {
		t.Fatalf("Prepare() unexpected error: %v", err)
	}

	doc := func() string {
		var buf bytes.Buffer
		if err := writeDocx(&buf, renderBody(blocks)); err != nil {
			t.Fatalf("writeDocx() unexpected error: %v", err)
		}
		return readDocx(t, buf.Bytes())["word/document.xml"]
	}()

	if !strings.Contains(doc, "keep me") {
		t.Error("filtered content is missing from the document")
	}
	for _, unwanted := range []string{"intro text", "drop me", "Result"} {
		if strings.Contains(doc, unwanted) {
			t.Errorf("document.xml must not contain %q", unwanted)
		}
	}
}
