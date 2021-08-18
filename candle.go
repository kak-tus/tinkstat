package main

import (
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/ssgreg/repeat"
)

func (pr *processor) searchCandle(figi string, date time.Time) sdk.Candle {
	date = date.Truncate(time.Minute)

	dateLimit := date.AddDate(0, 0, -7)

	for date.After(dateLimit) {
		candle, ok := pr.getCandle(figi, date)
		if !ok {
			if pr.needFetchCandle(figi, date) {
				candles, full := pr.fetchCandles(figi, date)

				_, ok := pr.cache.FIGIIndex[figi]
				if !ok {
					pr.cache.FIGIIndex[figi] = make(byHourIndex)
				}

				hourKey := date.UTC().Truncate(time.Hour).Format(time.RFC3339)

				_, ok = pr.cache.FIGIIndex[figi][hourKey]
				if !ok {
					pr.cache.FIGIIndex[figi][hourKey] = byMinuteIndex{
						Full:  full,
						Items: make(map[string]sdk.Candle),
					}
				}

				for _, candle := range candles {
					minuteKey := candle.TS.UTC().Truncate(time.Minute).Format(time.RFC3339)

					pr.cache.FIGIIndex[figi][hourKey].Items[minuteKey] = candle
				}

				pr.saveCache()

				candle, _ = pr.getCandle(figi, date)
			}
		}

		if candle.FIGI != "" {
			return candle
		}

		date = date.Add(-time.Minute)
	}

	panic("no cache found")
}

func (pr *processor) needFetchCandle(figi string, date time.Time) bool {
	if date.After(time.Now()) {
		return false
	}

	hourIdx, ok := pr.cache.FIGIIndex[figi]
	if !ok {
		return true
	}

	hourKey := date.UTC().Truncate(time.Hour).Format(time.RFC3339)

	_, ok = hourIdx[hourKey]

	return !ok
}

func (pr *processor) getCandle(figi string, date time.Time) (sdk.Candle, bool) {
	hourIdx, ok := pr.cache.FIGIIndex[figi]
	if !ok {
		return sdk.Candle{}, false
	}

	hourKey := date.UTC().Truncate(time.Hour).Format(time.RFC3339)

	minuteIdx, ok := hourIdx[hourKey]
	if !ok {
		return sdk.Candle{}, false
	}

	minuteKey := date.UTC().Truncate(time.Minute).Format(time.RFC3339)

	candle, ok := minuteIdx.Items[minuteKey]
	if !ok {
		if minuteIdx.Full {
			// We have full hour index, so no candle for this minute anyway
			// Return ok but empty candle
			return sdk.Candle{}, true
		} else {
			return sdk.Candle{}, false
		}
	}

	return candle, true
}

func (pr *processor) fetchCandles(figi string, date time.Time) ([]sdk.Candle, bool) {
	from := date.Truncate(time.Hour)
	to := from.Add(time.Hour)

	full := false

	if time.Now().After(to) {
		full = true
	}

	var candles []sdk.Candle

	err := repeat.Repeat(
		repeat.Fn(func() error {
			var err error

			candles, err = pr.client.Candles(pr.ctx, from, to, sdk.CandleInterval1Min, figi)
			if err != nil {
				return repeat.HintTemporary(err)
			}

			return nil
		}),
		repeat.StopOnSuccess(),
		repeat.LimitMaxTries(10),
		repeat.WithDelay(repeat.FixedBackoff(5*time.Second).Set()),
	)
	if err != nil {
		panic(err)
	}

	return candles, full
}

func (pr *processor) searchCurrencyCandle(currency string, date time.Time) sdk.Candle {
	date = date.Truncate(time.Minute)

	figi, ok := pr.cache.CurrencyIndex[currency]
	if !ok {
		currencies := pr.fetchCurrencies()

		for _, curr := range currencies {
			ticker := curr.Ticker[0:3]
			pr.cache.CurrencyIndex[ticker] = curr.FIGI
		}

		pr.saveCache()

		figi, ok = pr.cache.CurrencyIndex[currency]
		if !ok {
			panic("no currency found")
		}
	}

	return pr.searchCandle(figi, date)
}
