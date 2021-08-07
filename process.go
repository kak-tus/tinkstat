package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
)

type processorOptions struct {
	cacheFile string
	token     string
}

type processor struct {
	cache     cache
	cacheFile string
	cancel    context.CancelFunc
	client    *sdk.RestClient
	ctx       context.Context
}

func newProcessor(opts processorOptions) *processor {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	pr := &processor{
		cacheFile: opts.cacheFile,
		cancel:    cancel,
		client:    sdk.NewRestClient(opts.token),
		ctx:       ctx,
	}

	pr.loadCache()

	return pr
}

func (pr *processor) process(from, to time.Time, ticker string) {
	operations, err := pr.client.Operations(pr.ctx, sdk.DefaultAccount, from, to, "")
	if err != nil {
		panic(err)
	}

	sort.Slice(operations, func(i, j int) bool {
		return operations[i].DateTime.Before(operations[j].DateTime)
	})

	instruments, err := pr.client.InstrumentByTicker(pr.ctx, ticker)
	if err != nil {
		panic(err)
	}

	figis := make(map[string]bool)

	for _, instr := range instruments {
		figis[instr.FIGI] = true
	}

	sold := 0.0
	bought := 0.0
	count := 0
	marginFullComission := 0.0
	marginPredictedComission := 0.0

	var currency string

	for _, op := range operations {
		switch op.OperationType {
		case sdk.BUY, sdk.SELL:
			if !figis[op.FIGI] {
				continue
			}

			if op.Status != sdk.OperationStatusDone {
				continue
			}

			currency = string(op.Currency)

			if op.OperationType == sdk.BUY {
				count += op.Quantity
				bought += op.Payment + op.Commission.Value
			} else if op.OperationType == sdk.SELL {
				count -= op.Quantity
				sold += op.Payment + op.Commission.Value
			}
		case sdk.OperationTypeMarginCommission:
			// First operation from previous day - skip it, it is not from our operations
			if currency == "" {
				continue
			}

			// Only short support now
			if count >= 0 {
				continue
			}

			candle := pr.getCurrencyCandle(currency, op.DateTime)

			sum := bought + sold

			comission := marginComission(sum * candle.ClosePrice)

			if !(comission < 0) {
				continue
			}

			marginFullComission += op.Payment
			marginPredictedComission += comission
		}
	}

	fmt.Printf("Продано на сумму (с учётом комиссий): %0.2f %s\n", sold, currency)
	fmt.Printf("Куплено на сумму (с учётом комиссий): %0.2f %s\n", bought, currency)

	candle := pr.getCurrencyCandle(currency, to)

	sumFull := bought + sold + marginFullComission/candle.ClosePrice
	sumPredicted := bought + sold + marginPredictedComission/candle.ClosePrice

	fmt.Printf("Остаток суммы (с учётом полной маржинальной комиссии): %0.2f %s\n", sumFull, currency)
	fmt.Printf(
		"Остаток суммы (с учётом прогнозной [только за эту акцию] маржинальной комиссии): %0.2f %s\n",
		sumPredicted,
		currency,
	)

	fmt.Printf("Остаток акций: %v шт.\n", count)

	if count < 0 {
		fmt.Printf(
			"Купить акции не дороже (без учёта комиссии): %0.2f %s\n",
			sumFull/math.Abs(float64(count)),
			currency,
		)
		fmt.Printf(
			"Купить акции не дороже (без учёта комиссии, маржинальная прогнозная): %0.2f %s\n",
			sumPredicted/math.Abs(float64(count)),
			currency,
		)
	} else if count > 0 {
		fmt.Printf(
			"Продать акции не дешевле (без учёта комиссии): %0.2f %s\n",
			sumFull/math.Abs(float64(count)),
			currency,
		)
		fmt.Printf(
			"Продать акции не дешевле (без учёта комиссии, маржинальная прогнозная): %0.2f %s\n",
			sumPredicted/math.Abs(float64(count)),
			currency,
		)
	}

	pr.cancel()
}

func marginComission(val float64) float64 {
	switch v := val; {
	case v < 3_000:
		return 0
	case v < 50_000:
		return -25
	case v < 100_000:
		return -45
	case v < 200_000:
		return -85
	case v < 300_000:
		return -115
	case v < 500_000:
		return -185
	case v < 1_000_000:
		return -365
	case v < 2_000_000:
		return -700
	case v < 5_000_000:
		return -1700
	default:
		return -val * 0.033 / 100
	}
}
