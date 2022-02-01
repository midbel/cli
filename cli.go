package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
)

var (
	Version     string
	BuildTime   string
	CompileWith string
	CompileHost string
)

const (
	BadExitCode = 1
)

type ExitError struct {
	Err  error
	Code int
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

func Exit(err error, code int) error {
	return &ExitError{
		Err:  err,
		Code: code,
	}
}

func Usage(cmd, help string, cs []*Command) func() {
	sort.Slice(cs, func(i, j int) bool { return cs[i].String() < cs[j].String() })
	f := func() {
		data := struct {
			Name     string
			Commands []*Command
		}{
			Name:     cmd,
			Commands: cs,
		}
		fs := template.FuncMap{
			"join": strings.Join,
		}
		t := template.Must(template.New("help").Funcs(fs).Parse(help))
		t.Execute(os.Stderr, data)

		os.Exit(2)
	}
	return f
}

func RunAndExit(cs []*Command, usage func()) {
	if err := Run(cs, usage); err != nil {
		var (
			exit *ExitError
			code = BadExitCode
		)
		if errors.As(err, &exit) {
			code, err = exit.Code, exit.Err
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}

func Run(cs []*Command, usage func()) error {
	var (
		fset    = flag.NewFlagSet("", flag.ExitOnError)
		version = struct {
			Short bool
			Long  bool
		}{}
	)
	fset.Usage = usage
	fset.BoolVar(&version.Short, "v", false, "")
	fset.BoolVar(&version.Long, "version", false, "")
	if err := fset.Parse(os.Args[1:]); err != nil {
		return err
	}

	var (
		set map[string]*Command
		cmd *Command
	)
	if version.Short || version.Long || (flag.NArg() > 0 && flag.Arg(0) == "version") {
		printVersion()
		return nil
	}
	if fset.NArg() == 0 || fset.Arg(0) == "help" {
		fset.Usage()
		return nil
	}

	set = make(map[string]*Command)
	for _, c := range cs {
		if !c.Runnable() {
			continue
		}
		set[c.String()] = c
		for _, a := range c.Alias {
			set[a] = c
		}
		if cmd == nil && c.Default {
			cmd = c
		}
	}
	args := fset.Args()
	if c, ok := set[fset.Arg(0)]; ok && c.Runnable() {
		c.Flag.Usage = c.Help
		return c.Run(c, args[1:])
	}
	if cmd != nil {
		return cmd.Run(cmd, args)
	}
	n := filepath.Base(os.Args[0])
	return fmt.Errorf(`%s: unknown subcommand "%s". run  "%[1]s help" for usage`, n, args[0])
}

func printVersion() {
	var (
		name    = filepath.Base(os.Args[0])
		syst    = runtime.GOOS
		arch    = runtime.GOARCH
		version = Version
		buf     strings.Builder
	)
	if BuildTime == "" {
		t := time.Now()
		if p, err := os.Executable(); err == nil {
			if i, err := os.Stat(p); err == nil {
				t = i.ModTime().Truncate(time.Hour)
			}
		}
		BuildTime = t.UTC().Format(time.RFC3339)
	}
	if version == "" {
		version = "unknown"
	}

	buf.WriteString(name)
	buf.WriteRune('-')
	buf.WriteString(version)
	buf.WriteRune(' ')

	buf.WriteString(syst)
	buf.WriteRune('/')
	buf.WriteString(arch)
	buf.WriteRune(' ')
	buf.WriteString(BuildTime)

	if CompileWith != "" {
		buf.WriteString(" (compile with ")
		buf.WriteString(CompileWith)
		if CompileHost != "" {
			buf.WriteString(" - ")
			buf.WriteString(CompileHost)
		}
		buf.WriteString(")")
	}
	fmt.Fprintln(os.Stdout, buf.String())
}

type Command struct {
	Desc    string
	Usage   string
	Short   string
	Default bool
	Alias   []string
	Flag    flag.FlagSet
	Run     func(*Command, []string) error
}

func (c *Command) Help() {
	if len(c.Desc) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Desc))
	} else {
		fmt.Fprintln(os.Stderr, c.Short)
	}
	fmt.Fprintf(os.Stderr, "\nusage: %s\n", c.Usage)
	os.Exit(2)
}

func (c *Command) String() string {
	ix := strings.Index(c.Usage, " ")
	if ix < 0 {
		return c.Usage
	}
	return c.Usage[:ix]
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}
