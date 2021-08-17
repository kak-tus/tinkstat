package main

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
)

func (pr *processor) fetchCurrencies() []sdk.Instrument {
	currencies, err := pr.client.Currencies(pr.ctx)
	if err != nil {
		panic(err)
	}

	return currencies
}
