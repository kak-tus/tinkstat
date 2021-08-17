package main

import (
	"io/ioutil"
	"os"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/goccy/go-json"
)

type cache struct {
	CurrencyIndex currencyIndex
	FIGIIndex     figiIndex
}

type figiIndex map[string]byHourIndex

type byHourIndex map[string]byMinuteIndex

type byMinuteIndex struct {
	Full  bool
	Items map[string]sdk.Candle
}

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

func (pr *processor) saveCache() {
	idx := make(figiIndex)

	for figi := range pr.cache.FIGIIndex {
		idx[figi] = make(byHourIndex)

		for hourKey := range pr.cache.FIGIIndex[figi] {
			if !pr.cache.FIGIIndex[figi][hourKey].Full {
				// Don't save not full hour to cache
				continue
			}

			idx[figi][hourKey] = pr.cache.FIGIIndex[figi][hourKey]
		}
	}

	toSave := cache{
		CurrencyIndex: pr.cache.CurrencyIndex,
		FIGIIndex:     idx,
	}

	encoded, err := json.Marshal(toSave)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(pr.cacheFile, encoded, 0644); err != nil {
		panic(err)
	}
}
