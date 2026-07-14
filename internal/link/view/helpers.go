package view

import (
	"fmt"
	"url_shortener/internal/link/model"
)

type ChartBar struct {
	Day    string
	Label  string
	Height string
	Clicks int
}

func referrerName(ref string) string {
	if ref == "" {
		return "Прямой переход"
	}
	return ref
}

func chartBars(daily []model.DailyClick, maxHeight float64) []ChartBar {
	if len(daily) == 0 {
		return nil
	}
	max := 0
	for _, d := range daily {
		if d.Clicks > max {
			max = d.Clicks
		}
	}
	if max == 0 {
		max = 1
	}
	bars := make([]ChartBar, len(daily))
	for i, d := range daily {
		label := d.Day
		if len(d.Day) >= 5 {
			label = d.Day[5:]
		}
		bars[i] = ChartBar{
			Day:    d.Day,
			Label:  label,
			Clicks: d.Clicks,
			Height: fmt.Sprintf("%.0f", float64(d.Clicks)/float64(max)*maxHeight),
		}
	}
	return bars
}
