package agents

import (
	"testing"
	"time"
)

func TestAggregateDailyForecast(t *testing.T) {
	item := func(dtTxt, desc string, min, max float64) ForecastItem {
		var f ForecastItem
		f.DtTxt = dtTxt
		f.Weather = append(f.Weather, struct {
			Description string `json:"description"`
		}{Description: desc})
		f.Main.TempMin = min
		f.Main.TempMax = max
		return f
	}

	items := []ForecastItem{
		item("2026-11-08 09:00:00", "light rain", 10, 16),
		item("2026-11-08 12:00:00", "clear sky", 12, 18),
		item("2026-11-08 21:00:00", "heavy rain", 8, 13),
		item("2026-11-09 12:00:00", "cloudy", 7, 14),
	}

	got := aggregateDailyForecast(items)
	if len(got) != 2 {
		t.Fatalf("expected 2 days, got %d", len(got))
	}

	if got[0].Date != "2026-11-08" || got[0].Description != "clear sky" {
		t.Errorf("day 1 = %+v, want date 2026-11-08 with noon description 'clear sky'", got[0])
	}
	if got[0].Min != 8 || got[0].Max != 18 {
		t.Errorf("day 1 min/max = %v/%v, want 8/18", got[0].Min, got[0].Max)
	}

	if got[1].Date != "2026-11-09" || got[1].Description != "cloudy" {
		t.Errorf("day 2 = %+v, want date 2026-11-09 with description 'cloudy'", got[1])
	}
}

func TestAggregateDailyForecastSkipsInvalidItems(t *testing.T) {
	items := []ForecastItem{
		{DtTxt: ""},
		{DtTxt: "short"},
	}

	if got := aggregateDailyForecast(items); len(got) != 0 {
		t.Fatalf("expected no days, got %+v", got)
	}
}

func TestDistanceFromNoon(t *testing.T) {
	tests := []struct {
		dtTxt string
		want  time.Duration
	}{
		{"2026-11-08 12:00:00", 0},
		{"2026-11-08 09:00:00", 3 * time.Hour},
		{"2026-11-08 15:00:00", 3 * time.Hour},
		{"not-a-date", 24 * time.Hour},
	}

	for _, tt := range tests {
		if got := distanceFromNoon(tt.dtTxt); got != tt.want {
			t.Errorf("distanceFromNoon(%q) = %v, want %v", tt.dtTxt, got, tt.want)
		}
	}
}
