package api

import ()

const ()

type HistoricalDataPoint struct {
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
}

type HistoricalDataResponse struct {
	Pair string                `json:"pair"`
	Data []HistoricalDataPoint `json:"ohlc"`
}

type HistoricalDataResponseWrapper struct {
	Inner HistoricalDataResponse `json:"data"`
}
