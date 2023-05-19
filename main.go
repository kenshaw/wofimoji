package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/kenshaw/emoji"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	name    = "wofimoji"
	version = "0.0.0-dev"
)

func main() {
	if err := run(context.Background(), os.Args); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			fmt.Fprintf(os.Stderr, "%s", exitError.Stderr)
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cliargs []string) error {
	args := &Args{
		Action: "copy",
	}
	c := &cobra.Command{
		Use:     name,
		Short:   name + ", the wofi emoji picker",
		Version: version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return args.Run(cmd.Context())
		},
	}
	_ = c.Flags().StringP("config", "c", "", "config file")
	c.Flags().StringVar(&args.Selector, "selector", "wofi", "selector command")
	c.Flags().StringSliceVarP(&args.SelectorArgs, "selector-args", "f", nil, "selector args")
	c.Flags().StringVar(&args.Clipboarder, "clipboarder", "wl-copy", "clipboarder command")
	c.Flags().StringVar(&args.Typer, "typer", "wtype", "typer command")
	c.Flags().VarP(actionFlag{&args.Action}, "action", "a", "action")
	c.Flags().StringVar(&args.Prompt, "prompt", "emoji", "wofi prompt")
	c.Flags().VarP(skinToneFlag{&args.SkinTone}, "skin-tone", "t", "skin tone")
	c.Flags().VarP(templateFlag{args}, "template", "T", "template file")
	c.SetVersionTemplate("{{ .Name }} {{ .Version }}\n")
	c.InitDefaultHelpCmd()
	c.SetArgs(cliargs[1:])
	c.SilenceErrors, c.SilenceUsage = true, true
	return c.ExecuteContext(ctx)
}

type Args struct {
	Selector     string
	SelectorArgs []string
	Clipboarder  string
	Typer        string
	Action       string
	Prompt       string
	SkinTone     emoji.SkinTone
	Template     *template.Template
}

func (args *Args) Run(ctx context.Context) error {
	if args.Template == nil {
		var err error
		if args.Template, err = args.NewTemplate(string(defaultTpl)); err != nil {
			return err
		}
	}
	r, w := io.Pipe()
	defer r.Close()
	m := make(map[string]emoji.Emoji)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(args.WriteEmojis(ctx, w, m))
	eg.Go(args.RunCommand(ctx, r, m))
	return eg.Wait()
}

func (args *Args) WriteEmojis(ctx context.Context, w *io.PipeWriter, m map[string]emoji.Emoji) func() error {
	return func() error {
		var err error
		var b []byte
		buf := new(bytes.Buffer)
		for _, e := range emoji.Gemoji() {
			buf.Reset()
			if err = args.Template.Execute(buf, e); err != nil {
				return err
			}
			b = bytes.TrimSpace(buf.Bytes())
			m[string(b)] = e
			if _, err = w.Write(append(b, '\n')); err != nil {
				return err
			}
		}
		return w.Close()
	}
}

func (args *Args) RunCommand(ctx context.Context, r *io.PipeReader, m map[string]emoji.Emoji) func() error {
	return func() error {
		flags := args.SelectorArgs
		flags = append(
			flags,
			"--dmenu",
			"--prompt", args.Prompt,
		)
		cmd := exec.CommandContext(ctx, args.Selector, flags...)
		cmd.Stdin = r
		buf, err := cmd.Output()
		if err != nil {
			return err
		}
		key := string(bytes.TrimSpace(buf))
		switch args.Action {
		case "copy":
			return args.Copy(ctx, m[key])
		case "type":
			return args.Type(ctx, m[key])
		}
		_, err = fmt.Fprintln(os.Stdout, key)
		return err
	}
}

func (args *Args) Copy(ctx context.Context, e emoji.Emoji) error {
	cmd := exec.CommandContext(ctx, args.Clipboarder)
	cmd.Stdin = strings.NewReader(e.Tone(args.SkinTone))
	return cmd.Run()
}

func (args *Args) Type(ctx context.Context, e emoji.Emoji) error {
	cmd := exec.CommandContext(ctx, args.Typer, e.Tone(args.SkinTone))
	return cmd.Run()
}

// NewTemplate creates a new template for s.
func (args *Args) NewTemplate(s string) (*template.Template, error) {
	return template.New(s).Funcs(map[string]any{
		"tone": func(e emoji.Emoji) string {
			return e.Tone(args.SkinTone)
		},
		"unique": unique,
	}).Parse(s)
}

type actionFlag struct {
	action *string
}

func (f actionFlag) String() string {
	if f.action != nil {
		return *f.action
	}
	return ""
}

func (f actionFlag) Type() string {
	return "action"
}

func (f actionFlag) Set(s string) error {
	switch s = strings.ToLower(s); s {
	case "":
		*f.action = "print"
		return nil
	case "copy", "type", "print": /*, "clipboard", "unicode", "copy-unicode", "menu" */
		*f.action = s
		return nil
	}
	return errors.New("invalid action")
}

type skinToneFlag struct {
	skinTone *emoji.SkinTone
}

func (f skinToneFlag) String() string {
	if f.skinTone != nil {
		return f.skinTone.String()
	}
	return ""
}

func (f skinToneFlag) Type() string {
	return "tone"
}

func (f skinToneFlag) Set(s string) error {
	switch strings.ToLower(s) {
	case "", "none", "neutral":
		*f.skinTone = emoji.Neutral
		return nil
	case "light":
		*f.skinTone = emoji.Light
		return nil
	case "medium-light":
		*f.skinTone = emoji.MediumLight
		return nil
	case "medium":
		*f.skinTone = emoji.Medium
		return nil
	case "medium-dark":
		*f.skinTone = emoji.MediumDark
		return nil
	case "dark":
		*f.skinTone = emoji.Dark
		return nil
	}
	return errors.New("invalid skin tone")
}

type templateFlag struct {
	args *Args
}

func (f templateFlag) String() string {
	return ""
}

func (f templateFlag) Set(s string) error {
	buf, err := os.ReadFile(s)
	if err != nil {
		return err
	}
	tpl, err := f.args.NewTemplate(string(buf))
	if err != nil {
		return err
	}
	*f.args.Template = *tpl
	return nil
}

func (f templateFlag) Type() string {
	return "file"
}

// unique returns unique words in v.
func unique(v ...interface{}) string {
	m := make(map[string]bool)
	var out []string
	f := func(z string) {
		z = cleanRE.ReplaceAllString(z, "")
		for _, s := range strings.FieldsFunc(z, func(r rune) bool {
			return unicode.IsSpace(r) || r == '_' || r == '-'
		}) {
			s = strings.ToLower(s)
			if !m[s] {
				out = append(out, s)
			}
			m[s] = true
		}
	}
	for _, x := range v {
		switch z := x.(type) {
		case string:
			f(z)
		case []string:
			for _, z := range z {
				f(z)
			}
		}
	}
	return strings.Join(out, " ")
}

// cleanRE is a cleaning regexp.
var cleanRE = regexp.MustCompile(`(:|,|\.)`)

//go:embed default.tpl
var defaultTpl []byte
