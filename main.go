package main

import (
	"log"
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
			Layout:   "2006-01-02 15:04",
			Name:     "from",
			Required: true,
			Usage:    "from time (parsed in local timezone)",
		},
		&cli.TimestampFlag{
			Layout: "2006-01-02 15:04",
			Name:   "to",
			Usage:  "to time (parsed in local timezone)",
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

		loc := time.Now().Location()

		from := time.Date(
			c.Timestamp("from").Year(), c.Timestamp("from").Month(), c.Timestamp("from").Day(),
			c.Timestamp("from").Hour(), c.Timestamp("from").Minute(), 0, 0, loc,
		)

		to := time.Date(
			c.Timestamp("to").Year(), c.Timestamp("to").Month(), c.Timestamp("to").Day(),
			c.Timestamp("to").Hour(), c.Timestamp("to").Minute(), 0, 0, loc,
		)

		pr.balance(from, to, c.String("ticker"))

		return nil
	}

	stat := func(c *cli.Context) error {
		opts := processorOptions{
			cacheFile: c.Path("cache"),
			token:     c.String("token"),
		}

		pr := newProcessor(opts)

		loc := time.Now().Location()

		from := time.Date(
			c.Timestamp("from").Year(), c.Timestamp("from").Month(), c.Timestamp("from").Day(),
			c.Timestamp("from").Hour(), c.Timestamp("from").Minute(), 0, 0, loc,
		)

		to := time.Date(
			c.Timestamp("to").Year(), c.Timestamp("to").Month(), c.Timestamp("to").Day(),
			c.Timestamp("to").Hour(), c.Timestamp("to").Minute(), 0, 0, loc,
		)

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
		log.Fatal(err)
	}
}
