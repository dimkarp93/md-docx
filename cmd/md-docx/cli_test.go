package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const cliSample = `---
front: matter
---
# Intro

intro text

---

## Result

keep me

## Other

drop me
`

var cliBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "md-docx-cli")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	cliBinary = filepath.Join(dir, "md-docx")
	build := exec.Command("go", "build", "-o", cliBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	os.Exit(m.Run())
}

func runCLI(t *testing.T, stdin string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command(cliBinary, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v failed: %v\nstderr: %s", cliBinary, args, err, stderr.String())
	}
	return out
}

func documentXML(t *testing.T, docx []byte) string {
	t.Helper()
	return readDocx(t, docx)["word/document.xml"]
}

func TestCLIReadsStdinAndWritesStdout(t *testing.T) {
	doc := documentXML(t, runCLI(t, cliSample, "convert"))

	if !strings.Contains(doc, "intro text") {
		t.Error("stdout docx is missing content from page 1")
	}
	if strings.Contains(doc, "front: matter") {
		t.Error("page 0 must be excluded by default")
	}
}

func TestCLIWritesOutputFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.docx")
	in := filepath.Join(t.TempDir(), "in.md")
	if err := os.WriteFile(in, []byte(cliSample), 0o644); err != nil {
		t.Fatal(err)
	}

	runCLI(t, "", "convert", "--in", in, "--out", out)

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(documentXML(t, data), "intro text") {
		t.Error("output file is missing content")
	}
}

func TestCLIPagesFlag(t *testing.T) {
	doc := documentXML(t, runCLI(t, cliSample, "convert", "--pages", "2"))

	if !strings.Contains(doc, "keep me") {
		t.Error("--pages 2 dropped the requested page")
	}
	if strings.Contains(doc, "intro text") {
		t.Error("--pages 2 must exclude page 1")
	}
}

func TestCLIHeadsFlagWiring(t *testing.T) {
	doc := documentXML(t, runCLI(t, cliSample, "convert", "--heads", "h2:result"))

	if !strings.Contains(doc, "keep me") {
		t.Error("--heads dropped the matched section")
	}
	if !strings.Contains(doc, "Result") {
		t.Error("--heads without --root-head-hide must keep the heading itself")
	}
	for _, unwanted := range []string{"intro text", "drop me"} {
		if strings.Contains(doc, unwanted) {
			t.Errorf("--heads must exclude %q", unwanted)
		}
	}
}

func TestCLIRootHeadHideWiring(t *testing.T) {
	doc := documentXML(t, runCLI(t, cliSample, "convert", "--heads", "h2:result", "--root-head-hide"))

	if !strings.Contains(doc, "keep me") {
		t.Error("--root-head-hide dropped the section content")
	}
	if strings.Contains(doc, "Result") {
		t.Error("--root-head-hide must hide the matched heading")
	}
}

func TestCLIVersion(t *testing.T) {
	if got := strings.TrimSpace(string(runCLI(t, "", "--version"))); got == "" {
		t.Error("--version printed nothing")
	}
}

func TestCLIRejectsBadPages(t *testing.T) {
	cmd := exec.Command(cliBinary, "convert", "--pages", "5-1")
	cmd.Stdin = strings.NewReader(cliSample)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected a non-zero exit for an invalid range")
	}
	if !strings.Contains(string(out), "5-1") {
		t.Errorf("error message does not mention the bad range: %s", out)
	}
}

func TestCLIHelp(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"-h"}, {"--help"}, {"-help"}} {
		cmd := exec.Command(cliBinary, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
		if !strings.Contains(string(out), "Usage:") {
			t.Errorf("%v: no usage text: %s", args, out)
		}
	}
}

func TestCLIUnknownCommandIsAnError(t *testing.T) {
	cmd := exec.Command(cliBinary, "bogus")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected a non-zero exit, got: %s", out)
	}
	if !strings.Contains(string(out), "bogus") {
		t.Errorf("error does not name the bad command: %s", out)
	}
}

func TestCLINoArgsIsAnError(t *testing.T) {
	cmd := exec.Command(cliBinary)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected a non-zero exit, got: %s", out)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("no usage text: %s", out)
	}
}
