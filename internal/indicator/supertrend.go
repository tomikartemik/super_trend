package indicator

import (
	"math"

	"super_trend/internal/api"
)

type Supertrend struct {
	Value       float64
	TrendUp     bool
	PrevUp      float64
	PrevDown    float64
	PrevTrendUp bool
}

func CalculateSupertrend(klines []api.Kline, period int, multiplier float64) []Supertrend {
	st := make([]Supertrend, len(klines))
	atr := calculateATR(klines, period)

	for i := range klines {
		if i < period {
			continue
		}
		hl2 := (klines[i].High + klines[i].Low) / 2
		upperBand := hl2 + multiplier*atr[i]
		lowerBand := hl2 - multiplier*atr[i]

		if i == period {
			st[i].Value = upperBand
			st[i].TrendUp = true
			st[i].PrevUp = upperBand
			st[i].PrevDown = lowerBand
			continue
		}

		prev := st[i-1]
		if klines[i].Close > prev.PrevDown {
			st[i].TrendUp = true
		} else if klines[i].Close < prev.PrevUp {
			st[i].TrendUp = false
		} else {
			st[i].TrendUp = prev.PrevTrendUp
		}

		if st[i].TrendUp {
			if lowerBand > prev.PrevDown {
				st[i].Value = lowerBand
			} else {
				st[i].Value = prev.PrevDown
			}
		} else {
			if upperBand < prev.PrevUp {
				st[i].Value = upperBand
			} else {
				st[i].Value = prev.PrevUp
			}
		}

		st[i].PrevUp = upperBand
		st[i].PrevDown = lowerBand
		st[i].PrevTrendUp = st[i].TrendUp
	}

	return st
}

func calculateATR(klines []api.Kline, period int) []float64 {
	atr := make([]float64, len(klines))
	tr := make([]float64, len(klines))

	for i := 1; i < len(klines); i++ {
		high := klines[i].High
		low := klines[i].Low
		prevClose := klines[i-1].Close

		tr[i] = math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
	}

	var sum float64
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	atr[period] = sum / float64(period)

	for i := period + 1; i < len(klines); i++ {
		atr[i] = (atr[i-1]*(float64(period-1)) + tr[i]) / float64(period)
	}

	return atr
}
