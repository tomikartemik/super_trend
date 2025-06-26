package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Kline struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type BybitClient struct {
	Key    string
	Secret string
}

func NewBybitClient() *BybitClient {
	return &BybitClient{
		Key:    os.Getenv("BYBIT_API_KEY"),
		Secret: os.Getenv("BYBIT_API_SECRET"),
	}
}

// ====== KLINES ======
func (c *BybitClient) GetKlines(symbol, interval string, limit int) ([]Kline, error) {
	url := fmt.Sprintf("https://api.bybit.com/v5/market/kline?category=spot&symbol=%s&interval=%s&limit=%d", symbol, interval, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			List [][]interface{} `json:"list"`
		} `json:"result"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	klines := make([]Kline, 0, len(result.Result.List))
	for _, item := range result.Result.List {
		ts, _ := strconv.ParseInt(item[0].(string), 10, 64)
		open, _ := strconv.ParseFloat(item[1].(string), 64)
		high, _ := strconv.ParseFloat(item[2].(string), 64)
		low, _ := strconv.ParseFloat(item[3].(string), 64)
		close, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[5].(string), 64)

		klines = append(klines, Kline{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return klines, nil
}

// ====== SIGN ======
func (c *BybitClient) sign(payload string, timestamp string) string {
	message := timestamp + c.Key + "5000" + payload
	mac := hmac.New(sha256.New, []byte(c.Secret))
	mac.Write([]byte(message))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// ====== BALANCE ======
func (c *BybitClient) GetUSDTBalance() (float64, error) {
	url := "https://api.bybit.com/v5/account/wallet-balance?accountType=SPOT"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := c.sign("", timestamp)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-BYBIT-API-KEY", c.Key)
	req.Header.Set("X-BYBIT-SIGN", signature)
	req.Header.Set("X-BYBIT-TIMESTAMP", timestamp)
	req.Header.Set("X-BYBIT-RECV-WINDOW", "5000")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var resp struct {
		Result struct {
			List []struct {
				Coin []struct {
					Coin string `json:"coin"`
					Free string `json:"availableToWithdraw"`
				} `json:"coin"`
			} `json:"list"`
		} `json:"result"`
	}
	body, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}

	for _, coin := range resp.Result.List[0].Coin {
		if coin.Coin == "USDT" {
			return strconv.ParseFloat(coin.Free, 64)
		}
	}

	return 0, fmt.Errorf("USDT balance not found")
}

// ====== ORDER ======
func (c *BybitClient) PlaceOrder(symbol, side, orderType string, qty float64) error {
	url := "https://api.bybit.com/v5/order/create"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	bodyMap := map[string]interface{}{
		"category":  "spot",
		"symbol":    symbol,
		"side":      side,
		"orderType": orderType,
		"qty":       fmt.Sprintf("%.6f", qty),
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	signature := c.sign(string(bodyBytes), timestamp)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BYBIT-API-KEY", c.Key)
	req.Header.Set("X-BYBIT-SIGN", signature)
	req.Header.Set("X-BYBIT-TIMESTAMP", timestamp)
	req.Header.Set("X-BYBIT-RECV-WINDOW", "5000")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	body, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["retCode"].(float64) != 0 {
		return fmt.Errorf("order error: %v", result)
	}

	return nil
}
