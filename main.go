package main

import (
	"flag"
	"time"
)

func main() {
	token := flag.String("token", "", "token")
	fromRaw := flag.String("from", "", "from time 2021-01-01 or 2021-01-01 11:00:00")
	toRaw := flag.String("to", "", "optional to time 2021-01-01 or 2021-01-01 11:00:00")
	ticker := flag.String("ticker", "", "ticker")
	cacheFile := flag.String("cache", "", "cache file")

	flag.Parse()

	opts := processorOptions{
		cacheFile: *cacheFile,
		token:     *token,
	}

	pr := newProcessor(opts)

	from := parseTime(*fromRaw)

	to := time.Now()

	if *toRaw != "" {
		to = parseTime(*toRaw)
	}

	pr.process(from, to, *ticker)
}

func parseTime(date string) time.Time {
	parsed, err := time.Parse("2006-01-02", date)
	if err == nil {
		return parsed
	}

	parsed, err = time.Parse("2006-01-02 15:04:05", date)
	if err == nil {
		return parsed
	}

	panic(err)
}
