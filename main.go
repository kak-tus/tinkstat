package main

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "token",
			Required: true,
			Usage:    "token",
		},
		&cli.TimestampFlag{
			Layout:   "2006-01-02",
			Name:     "from",
			Required: true,
			Usage:    "from time",
		},
		&cli.TimestampFlag{
			Layout: "2006-01-02",
			Name:   "to",
			Usage:  "to time",
			Value:  cli.NewTimestamp(time.Now()),
		},
		&cli.StringFlag{
			Name:     "ticker",
			Required: true,
			Usage:    "ticker",
		},
		&cli.PathFlag{
			Name:     "cache",
			Required: true,
			Usage:    "cache file",
		},
	}

	balance := func(c *cli.Context) error {
		opts := processorOptions{
			cacheFile: c.Path("cache"),
			token:     c.String("token"),
		}

		pr := newProcessor(opts)

		from := c.Timestamp("from").Local()
		to := c.Timestamp("to").Local()

		pr.balance(from, to, c.String("ticker"))

		return nil
	}

	stat := func(c *cli.Context) error {
		opts := processorOptions{
			cacheFile: c.Path("cache"),
			token:     c.String("token"),
		}

		pr := newProcessor(opts)

		from := c.Timestamp("from").Local()
		to := c.Timestamp("to").Local()

		return pr.stat(from, to, c.String("ticker"))
	}

	commands := []*cli.Command{
		{
			Name:   "balance",
			Action: balance,
		},
		{
			Name:   "stat",
			Action: stat,
		},
	}

	app := &cli.App{
		Commands: commands,
		Flags:    flags,
		Name:     "tinkstat",
		Usage:    "Get stat from tinkoff API",
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
