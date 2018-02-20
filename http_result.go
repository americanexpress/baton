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

import "math"

// HTTPResult contains counters for the responses to the HTTP requests
type HTTPResult struct {
	connectionErrorCount int
	status1xxCount       int
	status2xxCount       int
	status3xxCount       int
	status4xxCount       int
	status5xxCount       int
	maxTime              int
	minTime              int
	timeSum              int64
	totalSuccess         int
}

func newHTTPResult() *HTTPResult {
	return &HTTPResult{0, 0, 0, 0, 0, 0, 0, math.MaxInt64, 0, 0}
}

func (httpResult HTTPResult) total() int {
	totalRequestsCounter := 0
	totalRequestsCounter += httpResult.connectionErrorCount
	totalRequestsCounter += httpResult.status1xxCount
	totalRequestsCounter += httpResult.status2xxCount
	totalRequestsCounter += httpResult.status3xxCount
	totalRequestsCounter += httpResult.status4xxCount
	totalRequestsCounter += httpResult.status5xxCount

	return totalRequestsCounter
}
