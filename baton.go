package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	supressOutput    = flag.Bool("o", false, "Supress output, no results will be printed to stdout")
	url              = flag.String("u", "", "URL to run against")
	wait             = flag.Int("w", 0, "Number of seconds to wait before running test")
)

// Configuration represents the Baton configuration
type Configuration struct {
	body             string
	concurrency      int
	dataFilePath     string
	duration         int
	ignoreTLS        bool
	method           string
	numberOfRequests int
	requestsFromFile string
	supressOutput    bool
	url              string
	wait             int
}

// Baton implements the load tester
type Baton struct {
	configuration Configuration
	result        Result
}

type preloadedRequest struct {
	method string
	url    string
	body   string
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
		*supressOutput,
		*url,
		*wait,
	}

	baton := &Baton{configuration: configuration, result: Result{}}

	baton.run()
	baton.result.printResults()
}

func preloadRequestsFromFile(filename string) ([]preloadedRequest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var requests []preloadedRequest
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ">", 3)
		method := parts[0]
		url := parts[1]
		body := parts[2]
		requests = append(requests, preloadedRequest{method, url, body})
	}
	return requests, scanner.Err()
}

func (baton *Baton) run() {

	logWriter := &logWriter{true}

	if baton.configuration.supressOutput {
		logWriter.Disable()
	}

	log.SetFlags(0)
	log.SetOutput(logWriter)

	preloadedRequestsMode := false
	timedMode := false
	var preloadedRequests []preloadedRequest
	var err error

	if baton.configuration.requestsFromFile != "" {
		preloadedRequests, err = preloadRequestsFromFile(baton.configuration.requestsFromFile)
		preloadedRequestsMode = true
		if err != nil {
			validationError("Failed to parse requests from file: " + baton.configuration.requestsFromFile)
		}
	} else if baton.configuration.url == "" {
		validationError("")
	}

	if baton.configuration.duration != 0 {
		timedMode = true
	}

	if baton.configuration.concurrency == 0 || baton.configuration.numberOfRequests == 0 {
		validationError("Invalid concurrency level or number of requests")
	}

	client := &fasthttp.Client{}
	if baton.configuration.ignoreTLS {
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		client = &fasthttp.Client{TLSConfig: tlsConfig}
	}

	switch baton.configuration.method {
	case "GET", "POST", "PUT", "DELETE":
		if baton.configuration.dataFilePath != "" {
			data, err := ioutil.ReadFile(baton.configuration.dataFilePath)
			if err != nil {
				validationError(err.Error())
			}
			baton.configuration.body = string(data)
		}
	default:
		validationError("Invalid method specified")
	}

	if preloadedRequestsMode {
		log.Printf("Configuring to send requests from file. (Read %d requests)\n", len(preloadedRequests))
	} else {
		log.Printf("Configuring to send %s requests to: %s\n", baton.configuration.method, baton.configuration.url)
	}

	if baton.configuration.wait > 0 {
		time.Sleep(time.Duration(baton.configuration.wait) * time.Second)
	}

	requests := make(chan bool, baton.configuration.numberOfRequests)
	results := make(chan HTTPResult, baton.configuration.concurrency)
	done := make(chan bool, baton.configuration.concurrency)

	log.Println("Generating the requests...")
	for r := 1; r <= baton.configuration.numberOfRequests; r++ {
		requests <- true
	}
	close(requests)
	log.Println("Finished generating the requests")
	log.Println("Sending the requests to the server...")

	// Start the timer and kick off the workers
	start := time.Now()
	for w := 1; w <= baton.configuration.concurrency; w++ {
		var worker workable
		if timedMode {
			worker = newTimedWorker(requests, results, done, float64(baton.configuration.duration))
		} else {
			worker = newCountWorker(requests, results, done)
		}
		worker.setCustomClient(client)
		if preloadedRequestsMode {
			go worker.sendRequests(preloadedRequests)
		} else {
			go worker.sendRequest(preloadedRequest{baton.configuration.method, baton.configuration.url, baton.configuration.body})
		}
	}

	// Wait for all the workers to finish and then stop the timer
	for a := 1; a <= baton.configuration.concurrency; a++ {
		<-done
	}
	baton.result.timeTaken = time.Since(start)

	log.Println("Finished sending the requests")
	log.Println("Processing the results...")

	for a := 1; a <= baton.configuration.concurrency; a++ {
		result := <-results
		baton.result.httpResult.connectionErrorCount += result.connectionErrorCount
		baton.result.httpResult.status1xxCount += result.status1xxCount
		baton.result.httpResult.status2xxCount += result.status2xxCount
		baton.result.httpResult.status3xxCount += result.status3xxCount
		baton.result.httpResult.status4xxCount += result.status4xxCount
		baton.result.httpResult.status5xxCount += result.status5xxCount
	}

	baton.result.totalRequests = baton.result.httpResult.total()

	baton.result.requestsPerSecond = int(float64(baton.result.totalRequests)/baton.result.timeTaken.Seconds() + 0.5)
}

func validationError(msg string) {
	if msg != "" {
		log.Printf("\n%s\n\n", msg)
	}
	flag.PrintDefaults()
	os.Exit(2)
}
