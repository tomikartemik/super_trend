package trader

import (
	"log"

	"super_trend/internal/api"
	"super_trend/internal/indicator"
)

type Trader struct {
	Client      *api.BybitClient
	LastTrendUp *bool
}

func NewTrader(client *api.BybitClient) *Trader {
	return &Trader{Client: client}
}

func (t *Trader) AnalyzeAndTrade(symbol string) {
	klines, err := t.Client.GetKlines(symbol, "1", 100)
	if err != nil {
		log.Printf("ошибка получения свечей: %v", err)
		return
	}

	supertrend := indicator.CalculateSupertrend(klines, 10, 3.0)
	if len(supertrend) == 0 {
		log.Println("Supertrend пуст")
		return
	}

	latest := supertrend[len(supertrend)-1]

	// Первая итерация
	if t.LastTrendUp == nil {
		t.LastTrendUp = new(bool)
		*t.LastTrendUp = latest.TrendUp
		log.Printf("Инициализация направления: вверх? %v", latest.TrendUp)
		return
	}

	// Смена тренда
	if latest.TrendUp != *t.LastTrendUp {
		balance, err := t.Client.GetUSDTBalance()
		if err != nil {
			log.Printf("Ошибка получения баланса: %v", err)
			return
		}
		amount := balance * 0.05

		var side string
		if latest.TrendUp {
			side = "Buy"
			log.Println("Тренд вверх — открываем ЛОНГ")
		} else {
			side = "Sell"
			log.Println("Тренд вниз — открываем ШОРТ (на споте = продажа)")
		}

		if err := t.Client.PlaceOrder(symbol, side, "Market", amount); err != nil {
			log.Printf("Ошибка при выставлении ордера: %v", err)
			return
		}

		*t.LastTrendUp = latest.TrendUp
	}
}
