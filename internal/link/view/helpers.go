package view

import "url_shortener/internal/link/model"

type ChartBar struct {
	Day    string
	Label  string
	Clicks int
	Height float64
	Y      float64
	X      int
}

type ChartData struct {
	Bars     []ChartBar
	SvgWidth int
	HasData  bool
}

func referrerName(ref string) string {
	if ref == "" {
		return "Прямой переход"
	}
	return ref
}

func buildChart(daily []model.DailyClick) ChartData {
	if len(daily) == 0 {
		return ChartData{HasData: false}
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
	barH := 180.0
	bars := make([]ChartBar, len(daily))
	for i, d := range daily {
		label := d.Day
		if len(d.Day) >= 5 {
			label = d.Day[5:]
		}
		h := float64(d.Clicks) / float64(max) * barH
		bars[i] = ChartBar{
			Day:    d.Day,
			Label:  label,
			Clicks: d.Clicks,
			Height: h,
			Y:      210 - h,
			X:      10 + i*30,
		}
	}
	return ChartData{Bars: bars, SvgWidth: len(daily)*30 + 40, HasData: true}
}
