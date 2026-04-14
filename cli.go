package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/midbel/distance"
	"github.com/midbel/textwrap"
)

var ErrUsage = errors.New("usage")

var (
	Stdout = os.Stdout
	Stderr = os.Stderr
)

func NewFlagSet(name string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	return set
}

type UsageError struct {
	Name    string
	Message string
	Usage   string
}

func (e UsageError) Error() string {
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
}

type SuggestionError struct {
	Name   string
	Others []string
}

func (e SuggestionError) Error() string {
	return fmt.Sprintf("%s: unknown sub command", e.Name)
}

type Command struct {
	Name      string
	Alias     []string
	Summary   string
	Help      string
	Usage     string
	Handler
}

func Help(summary, help string) *Command {
	return &Command{
		Summary: summary,
		Help:    help,
		Handler: helpHandler{},
	}
}

func (c *Command) getHelp() string {
	return c.Help
}

func (c *Command) getSummary() string {
	return c.Summary
}

func (c *Command) getAliases() []string {
	return c.Alias
}

type Handler interface {
	Run([]string) error
}

type helpHandler struct{}

func (helpHandler) Run(_ []string) error {
	return flag.ErrHelp
}

type CommandNode struct {
	Name     string
	Children map[string]*CommandNode
	cmd      *Command
}

func createNode(name string) *CommandNode {
	return &CommandNode{
		Name:     name,
		Children: make(map[string]*CommandNode),
	}
}

func (c CommandNode) Help() {
	if c.cmd.Summary != "" {
		fmt.Fprintln(os.Stderr, textwrap.Wrap(c.cmd.Summary, 72))
		fmt.Fprintln(os.Stderr)
	}
	if c.cmd.Help != "" {
		fmt.Fprintln(os.Stderr, textwrap.Wrap(c.cmd.Help, 72))
		fmt.Fprintln(os.Stderr)
	}
	if len(c.Children) > 0 {
		fmt.Fprintln(os.Stderr, "Available sub command(s)")
		for s, n := range c.Children {
			fmt.Fprintf(os.Stderr, "- %s: %s", s, n.cmd.getSummary())
			fmt.Fprintln(os.Stderr)
		}
	}
	if c.cmd.Usage != "" {
		fmt.Fprintf(os.Stderr, "Usage: %s", c.cmd.Usage)
		fmt.Fprintln(os.Stderr)
	}
}

type CommandTrie struct {
	root    *CommandNode
	summary string
	help    string
}

func New() *CommandTrie {
	trie := CommandTrie{
		root: createNode(""),
	}
	return &trie
}

func (t *CommandTrie) SetSummary(summary string) {
	t.summary = summary
}

func (t *CommandTrie) SetHelp(help string) {
	t.help = help
}

func (t *CommandTrie) Register(paths []string, cmd *Command) error {
	if cmd.Handler == nil {
		return nil
	}
	node := t.root
	for _, name := range paths {
		if node.Children[name] == nil {
			node.Children[name] = createNode(name)
		}
		node = node.Children[name]
	}
	node.cmd = cmd
	return nil
}

func (t *CommandTrie) Help() {
	if t.summary != "" {
		fmt.Fprintln(os.Stderr, textwrap.Wrap(t.summary, 72))
		fmt.Fprintln(os.Stderr)
	}
	if t.help != "" {
		fmt.Fprintln(os.Stderr, textwrap.Wrap(t.help, 72))
		fmt.Fprintln(os.Stderr)
	}
	if len(t.root.Children) == 0 {
		return
	}
	var longest int
	for k := range t.root.Children {
		longest = max(longest, len(k))
	}
	longest++
	commands := slices.Collect(maps.Keys(t.root.Children))
	slices.Sort(commands)
	fmt.Fprintln(os.Stderr, "Available commands")
	for _, k := range commands {
		n := t.root.Children[k]

		var summary string
		if n.cmd != nil {
			summary = n.cmd.getSummary()
		}
		fmt.Printf("- %-*s: %s", longest, k, summary)
		fmt.Fprintln(os.Stderr)
	}
}

func (t *CommandTrie) Execute(args []string) error {
	var (
		node = t.root
		ix   int
	)
	for _, name := range args {
		child := node.Children[name]
		if child == nil {
			break
		}
		node = child
		ix++
	}
	if node.cmd == nil {
		var found bool
		for _, c := range node.Children {
			if c.cmd == nil {
				continue
			}
			found = slices.Contains(c.cmd.getAliases(), args[ix])
			if found {
				ix++
				node = c
				break
			}
		}
		if !found {
			list := slices.Collect(maps.Keys(node.Children))
			return t.sugget(args[ix], list)
		}
	}
	err := node.cmd.Run(args[ix:])
	if errors.Is(err, flag.ErrHelp) {
		node.Help()
		return nil
	} else if errors.Is(err, ErrUsage) {
		return fmt.Errorf("Usage: %s", node.cmd.Usage)
	}
	return err
}

func (t *CommandTrie) sugget(name string, others []string) error {
	return SuggestionError{
		Name:   name,
		Others: distance.Levenshtein(name, others),
	}
}
