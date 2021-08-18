package main

import (
	"log"
	"strings"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/ssgreg/repeat"
)

func (pr *processor) searchCandle(figi string, date time.Time) sdk.Candle {
	date = date.Truncate(time.Minute)

	dateLimit := date.AddDate(0, 0, -7)

	for date.After(dateLimit) {
		candle, ok := pr.getCandle(figi, date)
		if ok {
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
	if !ok {
		return true
	}

	return false
}

func (pr *processor) getCandle(figi string, date time.Time) (sdk.Candle, bool) {
	candle, ok := pr.getCandleFromCache(figi, date)
	if ok {
		return candle, ok
	}

	if !pr.needFetchCandle(figi, date) {
		return sdk.Candle{}, false
	}

	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.AddDate(0, 0, 1)

	fullLimit := time.Now()

	candles := pr.fetchCandles(figi, from, to)

	_, ok = pr.cache.FIGIIndex[figi]
	if !ok {
		pr.cache.FIGIIndex[figi] = make(byHourIndex)
	}

	for from.Before(to) {
		hourKey := from.UTC().Format(time.RFC3339)

		var full bool
		if from.Add(time.Hour).Before(fullLimit) {
			full = true
		}

		_, ok = pr.cache.FIGIIndex[figi][hourKey]
		if !ok {
			pr.cache.FIGIIndex[figi][hourKey] = byMinuteIndex{
				Full:  full,
				Items: make(map[string]sdk.Candle),
			}
		}

		from = from.Add(time.Hour)
	}

	for _, candle := range candles {
		hourKey := candle.TS.UTC().Truncate(time.Hour).Format(time.RFC3339)

		var full bool
		if candle.TS.UTC().Truncate(time.Hour).Add(time.Hour).Before(fullLimit) {
			full = true
		}

		_, ok = pr.cache.FIGIIndex[figi][hourKey]
		if !ok {
			pr.cache.FIGIIndex[figi][hourKey] = byMinuteIndex{
				Full:  full,
				Items: make(map[string]sdk.Candle),
			}
		}

		minuteKey := candle.TS.UTC().Truncate(time.Minute).Format(time.RFC3339)

		pr.cache.FIGIIndex[figi][hourKey].Items[minuteKey] = candle
	}

	pr.saveCache()

	return pr.getCandleFromCache(figi, date)
}

func (pr *processor) getCandleFromCache(figi string, date time.Time) (sdk.Candle, bool) {
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
		return sdk.Candle{}, false
	}

	return candle, true
}

func (pr *processor) fetchCandles(figi string, from, to time.Time) []sdk.Candle {
	var candles []sdk.Candle

	err := repeat.Repeat(
		repeat.Fn(func() error {
			var err error

			candles, err = pr.client.Candles(pr.ctx, from, to, sdk.CandleInterval1Min, figi)
			// No correct way now to check
			if err != nil && strings.Contains(err.Error(), "code=429") {
				log.Println(err)
				return repeat.HintTemporary(err)
			}

			return nil
		}),
		repeat.StopOnSuccess(),
		repeat.LimitMaxTries(10),
		repeat.WithDelay(repeat.FixedBackoff(time.Minute).Set()),
	)
	if err != nil {
		panic(err)
	}

	return candles
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
