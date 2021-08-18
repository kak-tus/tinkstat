package main

import (
	"os"
	"text/template"
	"time"
)

type chartTemplate struct {
	From   string
	Labels []int64
	Points []float64
	Ticker string
}

func (pr *processor) stat(from, to time.Time, ticker string) error {
	templ, err := template.New("main").Parse(templateData)
	if err != nil {
		return err
	}

	wr, err := os.OpenFile("stat.html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer wr.Close()

	instruments, err := pr.client.InstrumentByTicker(pr.ctx, ticker)
	if err != nil {
		return err
	}

	var figi string

	// Only one first FIGI supported
	// TODO FIX
	for _, instr := range instruments {
		figi = instr.FIGI
		break
	}

	// Week start
	for from.Weekday() != time.Monday {
		from = from.AddDate(0, 0, -1)
	}

	// Session open
	from = time.Date(from.Year(), from.Month(), from.Day(), 7, 0, 0, 0, from.Location())

	// Week stop
	for to.Weekday() != time.Friday {
		to = to.AddDate(0, 0, 1)
	}

	// Session close
	to = to.AddDate(0, 0, 1)
	to = time.Date(to.Year(), to.Month(), to.Day(), 1, 45, 0, 0, to.Location())

	charts := make([]chartTemplate, 0)

	for from.Before(to) {
		localClose := from.AddDate(0, 0, 1)
		localClose = time.Date(localClose.Year(), localClose.Month(), localClose.Day(), 1, 45, 0, 0, localClose.Location())

		chart := chartTemplate{
			From:   from.Format("2006-01-02"),
			Ticker: ticker,
		}

		labels := make([]int64, 0)
		points := make([]float64, 0)

		for from.Before(localClose) {
			candle, ok := pr.getCandle(figi, from)

			if ok {
				labels = append(labels, candle.TS.Unix()*1000)
				points = append(points, candle.ClosePrice)
			}

			from = from.Add(time.Minute)
		}

		chart.Labels = labels
		chart.Points = points

		charts = append(charts, chart)

		// Go to next week start after last chart at this week
		if from.Weekday() == time.Saturday {
			for from.Weekday() != time.Monday {
				from = from.AddDate(0, 0, 1)
			}
		}

		// Session open
		from = time.Date(from.Year(), from.Month(), from.Day(), 7, 0, 0, 0, from.Location())
	}

	data := struct {
		Charts []chartTemplate
	}{
		Charts: charts,
	}

	return templ.Execute(wr, data)
}
