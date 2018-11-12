package cmd

import "github.com/urfave/cli"

var startCmd = cli.Command{
	Name:  "start",
	Usage: "swim start [--address <io:port>]",
	Action: func(c *cli.Context) error {
		return nil
	},
}

var joinCmd = cli.Command{
	Name:        "join",
	Aliases:     []string{"a"},
	Usage:       "option for join",
	Subcommands: cli.Commands{},
}

func Cmd() cli.Commands {
	joinCmd.Subcommands = append(joinCmd.Subcommands, member())

	return cli.Commands{joinCmd, startCmd}
}

func member() cli.Command {
	return cli.Command{
		Name:  "member",
		Usage: "swim join member [--address <ip:port>]",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}
