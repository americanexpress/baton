package main

// HTTPResult contains counters for the responses to the HTTP requests
type HTTPResult struct {
	connectionErrorCount int
	status1xxCount       int
	status2xxCount       int
	status3xxCount       int
	status4xxCount       int
	status5xxCount       int
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
