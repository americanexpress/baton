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
	"crypto/tls"
	"errors"
	"flag"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"time"
)

var (
	body             = flag.String("b", "", "Body (use instead of -f)")
	concurrency      = flag.Int("c", 1, "Number of concurrent requests")
	dataFilePath     = flag.String("f", "", "File path to file to be used as the body (use instead of -b)")
	duration         = flag.Int("t", 0, "Duration of testing in seconds (use instead of -r)")
	ignoreTLS        = flag.Bool("i", false, "Ignore TLS/SSL certificate validation ")
	method           = flag.String("m", "GET", "HTTP Method (GET,POST,PUT,DELETE)")
	numberOfRequests = flag.Int("r", 1, "Number of requests (use instead of -t)")
	requestsFromFile = flag.String("z", "", "Read requests from a file")
	suppressOutput   = flag.Bool("o", false, "Suppress output, no results will be printed to stdout")
	url              = flag.String("u", "", "URL to run against")
	wait             = flag.Int("w", 0, "Number of seconds to wait before running test")
)

// Baton implements the load tester
type Baton struct {
	configuration Configuration
	result        Result
}

type preLoadedRequest struct {
	method  string     // The HTTP method used to send the request
	url     string     // The URL to send the request at
	body    string     // The body of the request (if appropriate method is selected)
	headers [][]string // Array of two-element key/value pairs of header and value
}

type runConfiguration struct {
	preLoadedRequestsMode bool
	timedMode             bool
	preLoadedRequests     []preLoadedRequest
	client                *fasthttp.Client
	requests              chan bool
	results               chan HTTPResult
	done                  chan bool
	body                  string
}

func main() {
	flag.Parse()

	configuration := Configuration{
		*body,
		*concurrency,
		*dataFilePath,
		*duration,
		*ignoreTLS,
		*method,
		*numberOfRequests,
		*requestsFromFile,
		*suppressOutput,
		*url,
		*wait,
	}

	baton := &Baton{configuration: configuration, result: *newResult()}

	baton.run()
	baton.result.printResults()
}

func (baton *Baton) run() {

	configureLogging(baton.configuration.suppressOutput)

	err := baton.configuration.validate()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	preparedRunConfiguration, err := prepareRun(baton.configuration)
	if err != nil {
		log.Fatalf("Error during run preparation: %v", err)
	}

	if baton.configuration.wait > 0 {
		time.Sleep(time.Duration(baton.configuration.wait) * time.Second)
	}

	log.Println("Sending the requests to the server...")

	// Start the timer and kick off the workers
	start := time.Now()
	for w := 1; w <= baton.configuration.concurrency; w++ {
		var worker workable
		if preparedRunConfiguration.timedMode {
			worker = newTimedWorker(preparedRunConfiguration.requests, preparedRunConfiguration.results, preparedRunConfiguration.done, float64(baton.configuration.duration))
		} else {
			worker = newCountWorker(preparedRunConfiguration.requests, preparedRunConfiguration.results, preparedRunConfiguration.done)
		}
		worker.setCustomClient(preparedRunConfiguration.client)
		if preparedRunConfiguration.preLoadedRequestsMode {
			go worker.sendRequests(preparedRunConfiguration.preLoadedRequests)
		} else {
			request := preLoadedRequest{baton.configuration.method, baton.configuration.url, preparedRunConfiguration.body, [][]string{}}
			go worker.sendRequest(request)
		}
	}

	// Wait for all the workers to finish and then stop the timer
	for a := 1; a <= baton.configuration.concurrency; a++ {
		<-preparedRunConfiguration.done
	}
	baton.result.timeTaken = time.Since(start)

	log.Println("Finished sending the requests")
	log.Println("Processing the results...")

	processResults(baton, preparedRunConfiguration)
}

func processResults(baton *Baton, preparedRunConfiguration runConfiguration) {
	timeSum := int64(0)
	requestCount := 0
	for a := 1; a <= baton.configuration.concurrency; a++ {
		result := <-preparedRunConfiguration.results
		baton.result.httpResult.connectionErrorCount += result.connectionErrorCount
		baton.result.httpResult.status1xxCount += result.status1xxCount
		baton.result.httpResult.status2xxCount += result.status2xxCount
		baton.result.httpResult.status3xxCount += result.status3xxCount
		baton.result.httpResult.status4xxCount += result.status4xxCount
		baton.result.httpResult.status5xxCount += result.status5xxCount
		if result.minTime < baton.result.httpResult.minTime {
			baton.result.minTime = result.minTime
		}
		if result.maxTime > baton.result.httpResult.maxTime {
			baton.result.maxTime = result.maxTime
		}
		timeSum += result.timeSum
		requestCount += result.totalSuccess
	}
	baton.result.hasStats = baton.configuration.duration == 0
	baton.result.averageTime = float32(timeSum) / float32(requestCount)
	baton.result.totalRequests = baton.result.httpResult.total()
	baton.result.requestsPerSecond = int(float64(baton.result.totalRequests)/baton.result.timeTaken.Seconds() + 0.5)
}

func configureLogging(suppressOutput bool) {

	logWriter := &logWriter{true}

	if suppressOutput {
		logWriter.Disable()
	}

	log.SetFlags(0)
	log.SetOutput(logWriter)
}

func prepareRun(configuration Configuration) (runConfiguration, error) {

	preLoadedRequestsMode := false
	timedMode := false

	var preLoadedRequests []preLoadedRequest

	if configuration.requestsFromFile != "" {
		var err error
		preLoadedRequests, err = preLoadRequestsFromFile(configuration.requestsFromFile)
		preLoadedRequestsMode = true
		if err != nil {
			return runConfiguration{}, errors.New("failed to parse requests from file: " + configuration.requestsFromFile)
		}
	}

	if configuration.duration != 0 {
		timedMode = true
	}

	client := &fasthttp.Client{}
	if configuration.ignoreTLS {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		client = &fasthttp.Client{TLSConfig: tlsConfig}
	}

	body := configuration.body
	if configuration.dataFilePath != "" {
		data, err := ioutil.ReadFile(configuration.dataFilePath)
		if err != nil {
			return runConfiguration{}, err
		}
		body = string(data)
	}

	if preLoadedRequestsMode {
		log.Printf("Configuring to send requests from file. (Read %d requests)\n", len(preLoadedRequests))
	} else {
		log.Printf("Configuring to send %s requests to: %s\n", configuration.method, configuration.url)
	}

	requests := make(chan bool, configuration.numberOfRequests)
	results := make(chan HTTPResult, configuration.concurrency)
	done := make(chan bool, configuration.concurrency)

	log.Println("Generating the requests...")
	for r := 1; r <= configuration.numberOfRequests; r++ {
		requests <- true
	}
	close(requests)
	log.Println("Finished generating the requests")

	preparedRunConfiguration := runConfiguration{
		preLoadedRequestsMode,
		timedMode,
		preLoadedRequests,
		client,
		requests,
		results,
		done,
		body,
	}

	return preparedRunConfiguration, nil
}
