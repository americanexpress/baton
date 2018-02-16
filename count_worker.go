package main

import (
	"github.com/valyala/fasthttp"
)

// CountWorker implements a worker which sends a fixed number of requests
type countWorker struct {
	*worker
}

func newCountWorker(requests <-chan bool, results chan<- HTTPResult, done chan<- bool) *countWorker {
	worker := newWorker(requests, results, done)
	return &countWorker{worker}
}

func (worker *countWorker) sendRequest(request preloadedRequest) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(request.url)
	req.Header.SetMethod(request.method)
	req.SetBodyString(request.body)
	resp := fasthttp.AcquireResponse()

	for range worker.requests {
		worker.performRequest(req, resp)
	}

	worker.finish()
}
func (worker *countWorker) sendRequests(requests []preloadedRequest) {
	totalPremadeRequests := len(requests)

	for range worker.requests {
		req, resp := buildRequest(requests, totalPremadeRequests)
		worker.performRequest(req, resp)
	}

	worker.finish()
}
