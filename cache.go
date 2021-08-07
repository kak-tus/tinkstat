package main

import (
	"io/ioutil"
	"os"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/goccy/go-json"
)

type cache struct {
	CurrencyIndex currencyIndex
	FIGIIndex     figiIndex
}

type figiIndex map[string]byHourIndex

type byHourIndex map[string]byMinuteIndex

type byMinuteIndex map[string]sdk.Candle

type currencyIndex map[string]string

func (pr *processor) loadCache() {
	if _, err := os.Stat(pr.cacheFile); os.IsNotExist(err) {
		pr.cache.FIGIIndex = make(figiIndex)
		pr.cache.CurrencyIndex = make(currencyIndex)

		return
	}

	data, err := ioutil.ReadFile(pr.cacheFile)
	if err != nil {
		panic(err)
	}

	var newCache cache
	if err := json.Unmarshal(data, &newCache); err != nil {
		panic(err)
	}

	pr.cache = newCache
}

func (pr *processor) getCandle(figi string, date time.Time) sdk.Candle {
	date = date.Truncate(time.Minute)

	for date.After(date.AddDate(0, 0, -7)) {
		pr.updateCandleCache(figi, date)

		hourIdx, ok := pr.cache.FIGIIndex[figi]
		if !ok {
			panic("no figi cache found")
		}

		minuteIdx, ok := hourIdx[date.UTC().Truncate(time.Hour).Format(time.RFC3339)]
		if !ok {
			panic("no hour cache found")
		}

		candle, ok := minuteIdx[date.UTC().Format(time.RFC3339)]
		if !ok {
			date = date.Add(-time.Minute)
			continue
		}

		return candle
	}

	panic("no cache found")
}

func (pr *processor) updateCandleCache(figi string, date time.Time) {
	hourKey := date.UTC().Truncate(time.Hour).Format(time.RFC3339)

	if hourIdx, ok := pr.cache.FIGIIndex[figi]; ok {
		if _, ok := hourIdx[hourKey]; ok {
			return
		}
	} else {
		pr.cache.FIGIIndex[figi] = make(byHourIndex)
	}

	pr.cache.FIGIIndex[figi][hourKey] = make(byMinuteIndex)

	from := date.Truncate(time.Hour)
	to := from.Add(time.Hour)

	candles, err := pr.client.Candles(pr.ctx, from, to, sdk.CandleInterval1Min, figi)
	if err != nil {
		panic(err)
	}

	for _, candle := range candles {
		pr.cache.FIGIIndex[figi][hourKey][candle.TS.UTC().Format(time.RFC3339)] = candle
	}

	pr.saveCache()
}

func (pr *processor) updateCurrencyCache(currency string, date time.Time) {
	if _, ok := pr.cache.CurrencyIndex[currency]; ok {
		return
	}

	currencies, err := pr.client.Currencies(pr.ctx)
	if err != nil {
		panic(err)
	}

	for _, curr := range currencies {
		ticker := curr.Ticker[0:3]
		pr.cache.CurrencyIndex[ticker] = curr.FIGI
	}

	pr.saveCache()
}

func (pr *processor) saveCache() {
	encoded, err := json.Marshal(pr.cache)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(pr.cacheFile, encoded, 0644); err != nil {
		panic(err)
	}
}

func (pr *processor) getCurrencyCandle(currency string, date time.Time) sdk.Candle {
	date = date.Truncate(time.Minute)

	pr.updateCurrencyCache(currency, date)

	figi, ok := pr.cache.CurrencyIndex[currency]
	if !ok {
		panic("no currency cache found")
	}

	return pr.getCandle(figi, date)
}
