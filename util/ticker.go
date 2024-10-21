package util

import "time"

type FetchTicker struct {
	Ticker    *time.Ticker
	CloseChan chan struct{}
}

func NewFetchTicker(interval int) *FetchTicker {
	return &FetchTicker{getTicker(interval), make(chan struct{})}
}

func (f *FetchTicker) Stop() {
	f.Ticker.Stop()
	close(f.CloseChan)
}

func getTicker(interval int) *time.Ticker {
	d := time.Minute
	if IsDebug() {
		d = time.Second
	}
	return time.NewTicker(d * time.Duration(interval))
}
