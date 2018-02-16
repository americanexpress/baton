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
}

func (result *Result) printResults() {
	fmt.Println()
	fmt.Println()
	fmt.Printf("====================== Results ======================\n")
	fmt.Printf("Total requests:                            %10d\n", result.totalRequests)
	fmt.Printf("Time taken to complete requests:      %15s\n", result.timeTaken.String())
	fmt.Printf("Requests per second:                       %10d\n", result.requestsPerSecond)
	fmt.Printf("===================== Breakdown =====================\n")
	fmt.Printf("Number of connection errors:               %10d\n", result.httpResult.connectionErrorCount)
	fmt.Printf("Number of 1xx responses:                   %10d\n", result.httpResult.status1xxCount)
	fmt.Printf("Number of 2xx responses:                   %10d\n", result.httpResult.status2xxCount)
	fmt.Printf("Number of 3xx responses:                   %10d\n", result.httpResult.status3xxCount)
	fmt.Printf("Number of 4xx responses:                   %10d\n", result.httpResult.status4xxCount)
	fmt.Printf("Number of 5xx responses:                   %10d\n", result.httpResult.status5xxCount)
	fmt.Printf("=====================================================\n")
}
