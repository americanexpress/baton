package main

import (
	"github.com/valyala/fasthttp"
	"time"
)

// TimedWorker implements a worker which sends requests for a predetermined duration
type timedWorker struct {
	*worker
	durationToRun float64
}

func newTimedWorker(requests <-chan bool, results chan<- HTTPResult, done chan<- bool, durationToRun float64) *timedWorker {
	worker := newWorker(requests, results, done)
	return &timedWorker{worker, durationToRun}
}

func (worker timedWorker) sendRequest(request preLoadedRequest) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(request.url)
	req.Header.SetMethod(request.method)
	req.SetBodyString(request.body)
	resp := fasthttp.AcquireResponse()
	startTime := time.Now()

	for {
		if time.Since(startTime).Seconds() >= worker.durationToRun {
			break
		}

		worker.performRequest(req, resp)
	}

	worker.finish()
}

func (worker timedWorker) sendRequests(requests []preLoadedRequest) {
	totalPremadeRequests := len(requests)
	startTime := time.Now()

	for {
		if time.Since(startTime).Seconds() >= worker.durationToRun {
			break
		}
		req, resp := buildRequest(requests, totalPremadeRequests)
		worker.performRequest(req, resp)
	}

	worker.finish()
}
