package main

import (
	"os"
	"testing"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestCurrency(t *testing.T) {
	if os.Getenv("GODEBUG") == "" {
		t.SkipNow()
	}

	opts := processorOptions{
		token: os.Getenv("TOKEN"),
	}

	pr := newProcessor(opts)

	currencies, err := pr.client.Currencies(pr.ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, curr := range currencies {
		ticker := curr.Ticker[0:3]

		if ticker != "USD" {
			continue
		}

		from, err := time.Parse(time.RFC3339, "2021-08-17T22:07:00+03:00")
		assert.NoError(t, err)

		to, err := time.Parse(time.RFC3339, "2021-08-17T22:09:00+03:00")
		assert.NoError(t, err)

		candles, err := pr.client.Candles(pr.ctx, from, to, sdk.CandleInterval1Min, curr.FIGI)
		assert.NoError(t, err)

		spew.Dump(candles)
	}
}
