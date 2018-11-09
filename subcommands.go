package swim

import "github.com/urfave/cli"

var startCmd = cli.Command{
	Name:  "start",
	Usage: "swim start [--address <io:port>]",
	Action: func(c *cli.Context) error {
		return nil
	},
}

var addCmd = cli.Command{
	Name:        "add",
	Aliases:     []string{"a"},
	Usage:       "option for add",
	Subcommands: cli.Commands{},
}

func Cmd() cli.Commands {
	addCmd.Subcommands = append(addCmd.Subcommands, member())

	return cli.Commands{addCmd, startCmd}
}

func member() cli.Command {
	return cli.Command{
		Name:  "member",
		Usage: "swim add member [--address <ip:port>]",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}
