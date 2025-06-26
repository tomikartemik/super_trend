package main

import (
	"log"
	"time"

	"super_trend/internal/api"
	"super_trend/internal/trader"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	client := api.NewBybitClient()
	tr := trader.NewTrader(client)

	for {
		tr.AnalyzeAndTrade("SOLUSDT")
		time.Sleep(time.Minute)
	}
}
