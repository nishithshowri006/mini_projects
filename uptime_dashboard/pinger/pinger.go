package pinger

import (
	"context"
	
	"iter"
	"net/http"
	"sync"
	"time"

	
)

const TIMEOUT = 10

type URLMetadata struct {
	Interval int      `yaml:"Interval,omitempty"`
	URLs     []string `yaml:"URLs,omitempty"`
}
type IntervalList struct {
	URLMetadata URLMetadata `yaml:"UrlMetadata"`
}

type PingResponse struct {
	StatusCode   int
	PingedAt     time.Time
	URL          string
	Error        error
	ResponseTime time.Duration
}

type Pinger struct {
	Pings   chan PingResponse
	Inteval time.Duration
}

func NewPinger(buffCount int, interval time.Duration) *Pinger {
	c := make(chan PingResponse, buffCount)
	return &Pinger{Pings: c, Inteval: interval}
}

func (p *Pinger) ping(ctx context.Context, url string) {
	ctx, cancel := context.WithTimeout(ctx, TIMEOUT*time.Second)
	defer cancel()
	s := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		p.Pings <- PingResponse{
			URL:        url,
			PingedAt:   time.Now(),
			StatusCode: -1,
			Error:      err,
		}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.Pings <- PingResponse{
			URL:        url,
			PingedAt:   time.Now(),
			StatusCode: -1,
			Error:      err,
		}
		return
	}
	defer resp.Body.Close()
	e := time.Since(s)
	p.Pings <- PingResponse{
		URL:          url,
		PingedAt:     time.Now(),
		StatusCode:   resp.StatusCode,
		Error:        nil,
		ResponseTime: e,
	}
}

func (p *Pinger) pingLoop(ctx context.Context, url string) {
	tick := time.NewTicker(p.Inteval)
	defer tick.Stop()
	p.ping(ctx, url)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			p.ping(ctx, url)
		}
	}
}
func (p *Pinger) StartLoop(ctx context.Context, urls iter.Seq[string]) {
	var wg sync.WaitGroup
	for url := range urls {
		wg.Go(func() {
			p.pingLoop(ctx, url)
		})
	}
	wg.Wait()
	close(p.Pings)
}

