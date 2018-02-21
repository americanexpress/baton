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
	"math/rand"
	"time"
)

type worker struct {
	httpResult  HTTPResult
	client      *fasthttp.Client
	requests    <-chan bool
	httpResults chan<- HTTPResult
	done        chan<- bool
}

type workable interface {
	sendRequests(requests []preLoadedRequest)
	sendRequest(request preLoadedRequest)
	setCustomClient(client *fasthttp.Client)
}

func (worker *worker) setCustomClient(client *fasthttp.Client) {
	worker.client = client
}

func newWorker(requests <-chan bool, httpResults chan<- HTTPResult, done chan<- bool) *worker {
	return &worker{*newHTTPResult(), &fasthttp.Client{}, requests, httpResults, done}
}

func (worker *worker) performRequest(req *fasthttp.Request, resp *fasthttp.Response) bool {
	if err := worker.client.Do(req, resp); err != nil {
		worker.httpResult.connectionErrorCount++
		return true
	}
	status := resp.StatusCode()

	worker.recordCount(status)
	return false
}

func (worker *worker) recordCount(status int) {
	if status >= 100 && status < 200 {
		worker.httpResult.status1xxCount++
	} else if status >= 200 && status < 300 {
		worker.httpResult.status2xxCount++
	} else if status >= 300 && status < 400 {
		worker.httpResult.status3xxCount++
	} else if status >= 400 && status < 500 {
		worker.httpResult.status4xxCount++
	} else if status >= 500 && status < 600 {
		worker.httpResult.status5xxCount++
	}
}

func (worker *worker) performRequestWithStats(req *fasthttp.Request, resp *fasthttp.Response, timings chan int) bool {
	timeNow := time.Now().UnixNano()
	if err := worker.client.Do(req, resp); err != nil {
		worker.httpResult.connectionErrorCount++
		return true
	}
	timeAfter := time.Now().UnixNano()

	i := int(timeAfter - timeNow)
	// Record the timing into a channel
	timings <- i

	status := resp.StatusCode()
	worker.recordCount(status)

	return false
}

func buildRequest(requests []preLoadedRequest, totalPremadeRequests int) (*fasthttp.Request, *fasthttp.Response) {
	var currentReq preLoadedRequest

	currentReq = requests[rand.Intn(totalPremadeRequests)]
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.SetRequestURI(currentReq.url)
	req.Header.SetMethod(currentReq.method)
	req.SetBodyString(currentReq.body)
	for i := 0; i < len(currentReq.headers); i++ {
		req.Header.Add(currentReq.headers[i][0], currentReq.headers[i][1])
	}
	return req, resp
}

func (worker *worker) finish() {
	worker.httpResults <- worker.httpResult
	worker.done <- true
}

func (worker *worker) collectStatistics(timings chan int) {
	close(timings)

	first := true
	sum, total := int64(0), 0

	for timing := range timings {
		timing = timing / 1000
		// The first request is associated with overhead
		// in setting up the client so we ignore it's result
		if first {
			first = false
			continue
		}
		if timing < worker.httpResult.minTime {
			worker.httpResult.minTime = timing
		} else if timing >= worker.httpResult.maxTime {
			worker.httpResult.maxTime = timing
		}
		sum += int64(timing)
		total++
	}

	worker.httpResult.timeSum = sum
	worker.httpResult.totalSuccess = total
}
