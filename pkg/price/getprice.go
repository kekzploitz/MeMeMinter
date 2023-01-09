package price

import (
	"encoding/json"
	"io"
	"net/http"
)

type Prices struct {
	GBP float32 `json:"gbp"`
	USD float32 `json:"usd"`
	EUR float32 `json:"eur"`
}

type Price struct {
	Beam Prices
}

func GetPrice(coingeckoUrl string) (bool, float32, float32, float32) {
	resp, err := http.Get(coingeckoUrl)

	if err != nil {
		return false, 0, 0, 0
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, 0, 0, 0
	}

	data := Price{}

	_ = json.Unmarshal([]byte(body), &data)

	return true, data.Beam.GBP, data.Beam.USD, data.Beam.EUR
}
