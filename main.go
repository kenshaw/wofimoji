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
	"github.com/xo/ox"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/unicode/runenames"
)

var (
	name    = "wofimoji"
	version = "0.0.0-dev"
)

func main() {
	args := &Args{}
	ox.RunContext(
		context.Background(),
		ox.Defaults(),
		ox.Usage(name, "the wofi emoji picker"),
		ox.From(args),
		ox.Sort(true),
		ox.Exec(func(ctx context.Context) error {
			err := args.Run(ctx)
			if e := (*exec.ExitError)(nil); errors.As(err, &e) {
				os.Stderr.Write(e.Stderr)
				os.Exit(e.ExitCode())
			}
			return err
		}),
	)
}

type Args struct {
	Selector     string         `ox:"selector command,default:wofi"`
	SelectorArgs []string       `ox:"selector args,short:f"`
	Clipboarder  string         `ox:"clipboarder command,default:wl-copy"`
	Typer        string         `ox:"typer command,default:wtype"`
	Action       string         `ox:"action,default:copy,short:a"`
	Prompt       string         `ox:"wofi prompt"`
	Unicode      bool           `ox:"enable named unicode runes"`
	SkinTone     emoji.SkinTone //`ox:"skin tone,default:neutral,short:t"`
	Template     string         `ox:"template file,spec:file,short:T"`
}

func (args *Args) Run(ctx context.Context) error {
	tpl, err := newTemplate(args.Template, args.SkinTone)
	if err != nil {
		return err
	}
	r, w := io.Pipe()
	defer r.Close()
	m := make(map[string]emoji.Emoji)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(args.WriteEmojis(ctx, w, m, tpl))
	eg.Go(args.RunCommand(ctx, r, m))
	return eg.Wait()
}

func (args *Args) WriteEmojis(ctx context.Context, w *io.PipeWriter, m map[string]emoji.Emoji, tpl *template.Template) func() error {
	return func() error {
		var err error
		var b []byte
		buf := new(bytes.Buffer)
		for _, e := range emoji.Gemoji() {
			buf.Reset()
			if err = tpl.Execute(buf, e); err != nil {
				return err
			}
			b = bytes.TrimSpace(buf.Bytes())
			m[string(b)] = e
			if _, err = w.Write(append(b, '\n')); err != nil {
				return err
			}
		}
		if args.Unicode {
			var e emoji.Emoji
			for r := range rune(1_000_000) {
				s := runenames.Name(r)
				if s != "" && !strings.HasPrefix(s, "<") {
					buf.Reset()
					e.Emoji = string(r)
					e.Description = s
					if err = tpl.Execute(buf, e); err != nil {
						return err
					}
					b = bytes.TrimSpace(buf.Bytes())
					m[string(b)] = e
					if _, err = w.Write(append(b, '\n')); err != nil {
						return err
					}
				}
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

// newTemplate creates a new template for s.
func newTemplate(name string, tone emoji.SkinTone) (*template.Template, error) {
	var s string
	if name == "" {
		s = string(defaultTpl)
	} else {
		buf, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		s = string(buf)
	}
	return template.New(name).Funcs(map[string]any{
		"tone": func(e emoji.Emoji) string {
			return e.Tone(tone)
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

// unique returns unique words in v.
func unique(v ...any) string {
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
var cleanRE = regexp.MustCompile(`[:,\.]`)

//go:embed default.tpl
var defaultTpl []byte
