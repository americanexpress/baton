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
	"fmt"
	"time"
)

// Result contains the final output of a Baton execution
type Result struct {
	httpResult        HTTPResult
	totalRequests     int
	timeTaken         time.Duration
	requestsPerSecond int
	hasStats          bool
	averageTime       float32
	minTime           int
	maxTime           int
}

func newResult() *Result {
	return &Result{*newHTTPResult(), 0, 0, 0, false, 0, 0, 0}
}

func (result *Result) printResults() {
	fmt.Println()
	fmt.Println()
	fmt.Printf("====================== Results ======================\n")
	fmt.Printf("Total requests:                            %10d\n", result.totalRequests)
	fmt.Printf("Time taken to complete requests:      %15s\n", result.timeTaken.String())
	fmt.Printf("Requests per second:                       %10d\n", result.requestsPerSecond)
	if result.hasStats {
		fmt.Printf("Max response time (ms):                    %10d\n", result.maxTime)
		fmt.Printf("Min response time (ms):                    %10d\n", result.minTime)
		fmt.Printf("Avg response time (ms):                        %6.2f\n", result.averageTime)
	}
	fmt.Printf("===================== Breakdown =====================\n")
	fmt.Printf("Number of connection errors:               %10d\n", result.httpResult.connectionErrorCount)
	fmt.Printf("Number of 1xx responses:                   %10d\n", result.httpResult.status1xxCount)
	fmt.Printf("Number of 2xx responses:                   %10d\n", result.httpResult.status2xxCount)
	fmt.Printf("Number of 3xx responses:                   %10d\n", result.httpResult.status3xxCount)
	fmt.Printf("Number of 4xx responses:                   %10d\n", result.httpResult.status4xxCount)
	fmt.Printf("Number of 5xx responses:                   %10d\n", result.httpResult.status5xxCount)
	fmt.Printf("=====================================================\n")
}
