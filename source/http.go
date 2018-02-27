package source

import (
	"io/ioutil"
	"net/http"
	"time"
	//"fmt"
)

type HTTP struct {
	c    chan Result
	done chan struct{}
}

func NewHTTP(url string, interval time.Duration) HTTP {
	h := HTTP{
		c:    make(chan Result),
		done: make(chan struct{}),
	}
	go h.run(url, interval)
	return h
}

func (h HTTP) run(url string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	h.fetch(url)
	for {
		select {
		case <-t.C:
			h.fetch(url)
		case <-h.done:
			return
		}
	}
}

func (h HTTP) fetch(url string) {
	resp, err := http.Get(url)
	if err != nil {
		h.c <- Result{Err: err}
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.c <- Result{Err: err}
		return
	}
	result, err := JsonDataToResult(b)
	if err != nil || result.Err != nil {
		h.c <- Result{Err: err}
		return
	} else {
		h.c <- *result
	}
}

func (h HTTP) Get() (*Result, error) {
	res := <-h.c
	if res.Err != nil {
		return nil, res.Err
	}
	return &res, nil
}

func (h HTTP) Close() error {
	close(h.done)
	return nil
}
