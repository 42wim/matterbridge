package cmdhandler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/lrstanley/girc"
)

// Input is a wrapper for events, based around private messages.
type Input struct {
	Origin *girc.Event
	Args   []string
}

// Command is an IRC command, supporting aliases, help documentation and easy
// wrapping for message inputs.
type Command struct {
	// Name of command, e.g. "search" or "ping".
	Name string
	// Aliases for the above command, e.g. "s" for search, or "p" for "ping".
	Aliases []string
	// Help documentation. Should be in the format "<arg> <arg> [arg] --
	// something useful here"
	Help string
	// MinArgs is the minimum required arguments for the command. Defaults to
	// 0, which means multiple, or no arguments can be supplied. If set
	// above 0, this means that the command handler will throw an error asking
	// the person to check "<prefix>help <command>" for more info.
	MinArgs int
	// Fn is the function which is executed when the command is ran from a
	// private message, or channel.
	Fn func(*girc.Client, *Input)
}

func (c *Command) genHelp(prefix string) string {
	out := "{b}" + prefix + c.Name + "{b}"

	if c.Aliases != nil && len(c.Aliases) > 0 {
		out += " ({b}" + prefix + strings.Join(c.Aliases, "{b}, {b}"+prefix) + "{b})"
	}

	out += " :: " + c.Help

	return out
}

// CmdHandler is an irc command parser and execution format which you could
// use as an example for building your own version/bot.
//
// An example of how you would register this with girc:
//
//	ch, err := cmdhandler.New("!")
//	if err != nil {
//		panic(err)
//	}
//
//	ch.Add(&cmdhandler.Command{
//		Name:    "ping",
//		Help:    "Sends a pong reply back to the original user.",
//		Fn: func(c *girc.Client, input *cmdhandler.Input) {
//			c.Commands.ReplyTo(*input.Origin, "pong!")
//		},
//	})
//
//	client.Handlers.AddHandler(girc.PRIVMSG, ch)
type CmdHandler struct {
	prefix string
	re     *regexp.Regexp

	mu   sync.Mutex
	cmds map[string]*Command
}

var cmdMatch = `^%s([a-z0-9-_]{1,20})(?: (.*))?$`

// New returns a new CmdHandler based on the specified command prefix. A good
// prefix is a single character, and easy to remember/use. E.g. "!", or ".".
func New(prefix string) (*CmdHandler, error) {
	re, err := regexp.Compile(fmt.Sprintf(cmdMatch, regexp.QuoteMeta(prefix)))
	if err != nil {
		return nil, err
	}

	return &CmdHandler{prefix: prefix, re: re, cmds: make(map[string]*Command)}, nil
}

var validName = regexp.MustCompile(`^[a-z0-9-_]{1,20}$`)

// Add registers a new command to the handler. Note that you cannot remove
// commands once added, unless you add another CmdHandler to the client.
func (ch *CmdHandler) Add(cmd *Command) error {
	if cmd == nil {
		return errors.New("nil command provided to CmdHandler")
	}

	cmd.Name = strings.ToLower(cmd.Name)
	if !validName.MatchString(cmd.Name) {
		return fmt.Errorf("invalid command name: %q (req: %q)", cmd.Name, validName.String())
	}

	if cmd.Aliases != nil {
		for i := 0; i < len(cmd.Aliases); i++ {
			cmd.Aliases[i] = strings.ToLower(cmd.Aliases[i])
			if !validName.MatchString(cmd.Aliases[i]) {
				return fmt.Errorf("invalid command name: %q (req: %q)", cmd.Aliases[i], validName.String())
			}
		}
	}

	if cmd.MinArgs < 0 {
		cmd.MinArgs = 0
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if _, ok := ch.cmds[cmd.Name]; ok {
		return fmt.Errorf("command already registered: %s", cmd.Name)
	}

	ch.cmds[cmd.Name] = cmd

	// Since we'd be storing pointers, duplicates do not matter.
	for i := 0; i < len(cmd.Aliases); i++ {
		if _, ok := ch.cmds[cmd.Aliases[i]]; ok {
			return fmt.Errorf("alias already registered: %s", cmd.Aliases[i])
		}

		ch.cmds[cmd.Aliases[i]] = cmd
	}

	return nil
}

// Execute satisfies the girc.Handler interface.
func (ch *CmdHandler) Execute(client *girc.Client, event girc.Event) {
	if event.Source == nil || event.Command != girc.PRIVMSG {
		return
	}

	parsed := ch.re.FindStringSubmatch(event.Trailing)
	if len(parsed) != 3 {
		return
	}

	invCmd := strings.ToLower(parsed[1])
	args := strings.Split(parsed[2], " ")
	if len(args) == 1 && args[0] == "" {
		args = []string{}
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if invCmd == "help" {
		if len(args) == 0 {
			client.Cmd.ReplyTo(event, girc.Fmt("type '{b}!help {blue}<command>{c}{b}' to optionally get more info about a specific command."))
			return
		}

		args[0] = strings.ToLower(args[0])

		if _, ok := ch.cmds[args[0]]; !ok {
			client.Cmd.ReplyTof(event, girc.Fmt("unknown command {b}%q{b}."), args[0])
			return
		}

		if ch.cmds[args[0]].Help == "" {
			client.Cmd.ReplyTof(event, girc.Fmt("there is no help documentation for {b}%q{b}"), args[0])
			return
		}

		client.Cmd.ReplyTo(event, girc.Fmt(ch.cmds[args[0]].genHelp(ch.prefix)))
		return
	}

	cmd, ok := ch.cmds[invCmd]
	if !ok {
		return
	}

	if len(args) < cmd.MinArgs {
		client.Cmd.ReplyTof(event, girc.Fmt("not enough arguments supplied for {b}%q{b}. try '{b}%shelp %s{b}'?"), invCmd, ch.prefix, invCmd)
		return
	}

	in := &Input{
		Origin: &event,
		Args:   args,
	}

	go cmd.Fn(client, in)
}
