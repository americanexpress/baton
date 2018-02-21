/*
 * Copyright 2018 American Express
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package main

import (
	"github.com/valyala/fasthttp"
)

// CountWorker implements a worker which sends a fixed number of requests
type countWorker struct {
	*worker
	timings chan int
}

func newCountWorker(requests <-chan bool, results chan<- HTTPResult, done chan<- bool) *countWorker {
	worker := newWorker(requests, results, done)
	timings := make(chan int, len(requests))
	return &countWorker{worker, timings}
}

func (worker *countWorker) sendRequest(request preLoadedRequest) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(request.url)
	req.Header.SetMethod(request.method)
	req.SetBodyString(request.body)
	resp := fasthttp.AcquireResponse()

	for range worker.requests {
		worker.performRequestWithStats(req, resp, worker.timings)
	}

	worker.collectStatistics(worker.timings)
	worker.finish()
}
func (worker *countWorker) sendRequests(requests []preLoadedRequest) {
	totalPremadeRequests := len(requests)

	for range worker.requests {
		req, resp := buildRequest(requests, totalPremadeRequests)
		worker.performRequestWithStats(req, resp, worker.timings)
	}

	worker.collectStatistics(worker.timings)
	worker.finish()
}
