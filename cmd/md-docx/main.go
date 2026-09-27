package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dimkarp93/install-libs/buildinfo"
	"github.com/dimkarp93/install-libs/shellcomplete"
	mdlib "github.com/dimkarp93/md-libs"
)

var (
	version  string
	origin   string
	upstream string
	commit   string
	channel  string
)

func build() buildinfo.Info {
	return buildinfo.Info{
		Version:  version,
		Origin:   origin,
		Upstream: upstream,
		Commit:   commit,
		Channel:  channel,
	}
}

type convertFlags struct {
	in           *string
	out          *string
	pages        *string
	heads        *string
	rootHeadHide *bool
}

func newConvertFlagSet() (*flag.FlagSet, convertFlags) {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	f := convertFlags{
		in:           fs.String("in", "", "input markdown file (default: stdin)"),
		out:          fs.String("out", "", "output docx file (default: stdout)"),
		pages:        fs.String("pages", "", "pages to include, e.g. 1,3-5 (default: all pages except 0)"),
		heads:        fs.String("heads", "", "heading filters, e.g. h2:result,h3:resume"),
		rootHeadHide: fs.Bool("root-head-hide", false, "hide headings matched by --heads, keep their content"),
	}
	return fs, f
}

func completionSpec() shellcomplete.Spec {
	convertFS, _ := newConvertFlagSet()
	return shellcomplete.Spec{
		Bin: "md-docx",
		Flags: []shellcomplete.Flag{
			{Name: "--version", Bool: true},
			{Name: "-v", Bool: true},
			{Name: "--origin", Bool: true},
			{Name: "--buildinfo", Bool: true},
		},
		Commands: []shellcomplete.Command{
			{Name: "convert", Flags: shellcomplete.With(shellcomplete.FromFlagSet(convertFS),
				shellcomplete.Flag{Name: "-in", Files: true},
				shellcomplete.Flag{Name: "-out", Files: true},
			)},
			{Name: "help"},
		},
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage: md-docx <command> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  convert [flags]                render markdown into a .docx (see: md-docx convert -h)")
	fmt.Fprintln(w, "  help                            show this help")
	fmt.Fprintln(w, "  completion bash|zsh             print a shell completion script, for: source <(md-docx completion bash)")
	fmt.Fprintln(w, "  install-completions [SHELL]     install the completion script and wire it into ~/.bashrc / ~/.zshrc")
	fmt.Fprintln(w, "  uninstall-completions [SHELL]   remove the installed completion script")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --version, -v                   print version and exit")
	fmt.Fprintln(w, "  --origin                        print the repository the binary was built from and exit")
	fmt.Fprintln(w, "  --buildinfo                      print build metadata and exit")
	fmt.Fprintln(w, "  -h, --help, -help                show this help")
}

func main() {
	args := os.Args[1:]

	spec := completionSpec()
	if code, ok := spec.Handle(os.Stdout, os.Stderr, args); ok {
		os.Exit(code)
	}

	if len(args) == 0 {
		usage(os.Stderr)
		os.Exit(2)
	}

	switch args[0] {
	case "help", "-h", "--help", "-help":
		usage(os.Stdout)
		return
	case "--version", "-v":
		fmt.Println(build().VersionString())
		return
	case "--origin":
		fmt.Println(build().OriginString())
		return
	case "--buildinfo":
		build().Print()
		return
	case "convert":
		runConvert(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "md-docx: unknown command %q\n\n", args[0])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func runConvert(args []string) {
	fs, f := newConvertFlagSet()
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: md-docx convert [flags]")
		fmt.Fprintln(os.Stderr)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	var data []byte
	var err error
	if *f.in == "" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(*f.in)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	blocks, err := mdlib.Prepare(string(data), mdlib.Options{
		Pages:            *f.pages,
		Heads:            *f.heads,
		HideMatchedHeads: *f.rootHeadHide,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	body := renderBody(blocks)

	out := io.Writer(os.Stdout)
	if *f.out != "" {
		file, err := os.Create(*f.out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", *f.out, err)
			os.Exit(1)
		}
		defer file.Close()
		out = file
	}

	if err := writeDocx(out, body); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
