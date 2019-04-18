package cmd

import (
	"log"
	"os"
	"time"

	"github.com/DE-labtory/swim/cmd/common"
	"github.com/DE-labtory/swim/conf"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "swim"
	app.Version = ""
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "",
			Email: "",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "",
			Usage: "name for config",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "set debug mode",
		},
	}
	app.Commands = cli.Commands{}
	app.Commands = append(app.Commands, Cmd()...)

	app.Before = func(c *cli.Context) error {
		if configPath := c.String("config"); configPath != "" {
			absPath, err := common.RelativeToAbsolutePath(configPath)
			if err != nil {
				return err
			}
			conf.SetConfigPath(absPath)
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
