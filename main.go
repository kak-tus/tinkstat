package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
)

func main() {
	token := flag.String("token", "", "token")
	fromRaw := flag.String("from", "", "from time 2021-01-01 or 2021-01-01 11:00:00")
	toRaw := flag.String("to", "", "optional to time 2021-01-01 or 2021-01-01 11:00:00")
	ticker := flag.String("ticker", "", "ticker")

	flag.Parse()

	client := sdk.NewRestClient(*token)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	from := parseTime(*fromRaw)

	to := time.Now()

	if *toRaw != "" {
		to = parseTime(*toRaw)
	}

	operations, err := client.Operations(ctx, sdk.DefaultAccount, from, to, "")
	if err != nil {
		panic(err)
	}

	marginComission := 0.0

	for _, op := range operations {
		if op.OperationType != sdk.OperationTypeMarginCommission {
			continue
		}

		// TODO FIX (calculate only in case of margin trade)
		marginComission += op.Payment
	}

	instruments, err := client.InstrumentByTicker(ctx, *ticker)
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

	var currency string

	for _, op := range operations {
		if !figis[op.FIGI] {
			continue
		}

		if op.Status != sdk.OperationStatusDone {
			continue
		}

		if op.OperationType == sdk.BUY {
			currency = string(op.Currency)
			count += op.Quantity
			bought += op.Payment + op.Commission.Value
		} else if op.OperationType == sdk.SELL {
			currency = string(op.Currency)
			count -= op.Quantity
			sold += op.Payment + op.Commission.Value
		}
	}

	fmt.Printf("Продано на сумму (с учётом комиссий): %0.2f %s\n", sold, currency)
	fmt.Printf("Куплено на сумму (с учётом комиссий): %0.2f %s\n", bought, currency)

	// TODO FIX
	sum := bought + sold + marginComission/74

	fmt.Printf("Остаток суммы (с учётом маржинальной комиссии): %0.2f %s\n", sum, currency)
	fmt.Printf("Остаток акций: %v шт.\n", count)

	if count < 0 {
		fmt.Printf(
			"Купить акции не дороже (без учёта комиссии): %0.2f %s\n",
			sum/math.Abs(float64(count)),
			currency,
		)
	} else if count > 0 {
		fmt.Printf(
			"Продать акции не дешевле (без учёта комиссии): %0.2f %s\n",
			sum/math.Abs(float64(count)),
			currency,
		)
	}
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
