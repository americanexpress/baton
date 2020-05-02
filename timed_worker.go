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
