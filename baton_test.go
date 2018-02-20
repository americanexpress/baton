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
	"encoding/hex"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

type HTTPTestHandler struct {
	noRequestsReceived  uint32
	lastBodyReceived    string
	lastMethodReceived  string
	lastURIReceived     string
	lastHeadersReceived *fasthttp.RequestHeader
	lastTimestamp       int64
}

func (h *HTTPTestHandler) HandleRequest(ctx *fasthttp.RequestCtx) {
	atomic.AddUint32(&h.noRequestsReceived, 1)
	atomic.StoreInt64(&h.lastTimestamp, time.Now().Unix())
	h.lastBodyReceived = hex.EncodeToString(ctx.Request.Body())
	h.lastMethodReceived = string(ctx.Request.Header.Method())
	h.lastURIReceived = ctx.Request.URI().String()
	newHeader := fasthttp.RequestHeader{}
	ctx.Request.Header.CopyTo(&newHeader)
	h.lastHeadersReceived = &newHeader

}

func (h *HTTPTestHandler) reset() {
	h.noRequestsReceived = 0
	h.lastBodyReceived = ""
	h.lastMethodReceived = ""
	h.lastURIReceived = ""
	h.lastTimestamp = 0
}

var serverRunning = false
var internalHandlerRef *HTTPTestHandler
var port = "8888"

func startServer() *HTTPTestHandler {
	if !serverRunning {
		internalHandlerRef = &HTTPTestHandler{0, "", "", "", &fasthttp.RequestHeader{}, 0}
		serverRunning = true
		go func() {
			err := fasthttp.ListenAndServe(":"+port, internalHandlerRef.HandleRequest)
			if err != nil {
				fmt.Printf("Failed to bind on port %s. Make sure no other service is running on that port and restart the test\n", port)
				os.Exit(1)
			}
		}()
		return internalHandlerRef
	}
	internalHandlerRef.reset()
	return internalHandlerRef
}

func defaultConfig() Configuration {
	return Configuration{
		"",
		1,
		"",
		0,
		false,
		"GET",
		1,
		"",
		true,
		"http://localhost:" + port,
		0,
	}
}

func setupAndListen(config Configuration) *HTTPTestHandler {
	testHandler := startServer()
	// Give server enough time to spin
	time.Sleep(time.Duration(500) * time.Millisecond)
	// Create a baton instance with given config
	baton := &Baton{configuration: config, result: Result{}}
	baton.run()
	// Give server enough time to receive requests
	time.Sleep(time.Duration(500) * time.Millisecond)
	// Collect results from handler
	return testHandler
}

func TestRequestCount(t *testing.T) {
	noRequestsToSend := 10000

	config := defaultConfig()
	config.numberOfRequests = noRequestsToSend
	testHandler := setupAndListen(config)

	reqsReceived := int(testHandler.noRequestsReceived)
	if reqsReceived != noRequestsToSend {
		t.Errorf("Wrong number of requests sent. Expected %d, got %d", noRequestsToSend, reqsReceived)
	}
}

func TestRequestCountWithMoreWorkers(t *testing.T) {
	noRequestsToSend := 100000

	config := defaultConfig()
	config.concurrency = 10
	config.numberOfRequests = noRequestsToSend
	testHandler := setupAndListen(config)

	reqsReceived := int(testHandler.noRequestsReceived)
	if reqsReceived != noRequestsToSend {
		t.Errorf("Wrong number of requests sent. Expected %d, got %d", noRequestsToSend, reqsReceived)
	}
}

func TestThatBodyHasCorrectValue(t *testing.T) {
	body := "Hello World"
	bodyBytes := hex.EncodeToString([]byte(body))

	config := defaultConfig()
	config.body = body
	config.method = "POST"
	testHandler := setupAndListen(config)

	bytesReceived := testHandler.lastBodyReceived
	if bytesReceived != bodyBytes {
		t.Errorf("The body received by the server didn't match. Expected %s, got %s", bodyBytes, bytesReceived)
	}
}

func TestThatTheCorrectHTTPMethodIsUsed(t *testing.T) {
	method := "DELETE"

	config := defaultConfig()
	config.method = method
	testHandler := setupAndListen(config)

	methodReceived := testHandler.lastMethodReceived
	if methodReceived != method {
		t.Errorf("The HTTP method received by the server didn't match. Expected %s, got %s", method, methodReceived)
	}
}

func TestThatServerReceivesCorrectURI(t *testing.T) {
	uri := "http://localhost:" + port + "/path/to/complex?uri&with=html&entities=&#ef"

	config := defaultConfig()
	config.url = uri
	testHandler := setupAndListen(config)

	uriReceived := testHandler.lastURIReceived
	if uriReceived != uri {
		t.Errorf("The URI received by the server does not match. Expected %s, got %s.", uri, uriReceived)
	}
}

func TestLoadPostFromTextFile(t *testing.T) {
	body := "Hello World"
	bodyBytes := hex.EncodeToString([]byte(body))

	config := defaultConfig()
	config.dataFilePath = "test-resources/post-body.txt"
	config.method = "POST"
	testHandler := setupAndListen(config)

	bytesReceived := testHandler.lastBodyReceived
	if bytesReceived != bodyBytes {
		t.Errorf("The body received by the server didn't match. Expected %s, got %s", bodyBytes, bytesReceived)
	}
}

func TestPostRequestLoadedFromFile(t *testing.T) {
	uri := "http://localhost:" + port
	method := "POST"
	fileContents := method + "," + uri + "," + "Data"
	fileInBytes := []byte(fileContents)

	fileDir := "test-resources/requests-from-file.txt"
	if ioutil.WriteFile(fileDir, fileInBytes, 0644) != nil {
		t.Errorf("Failed to write a required test case file. Check the directory permissions.")
	}
	defer os.Remove(fileDir)

	config := defaultConfig()
	config.requestsFromFile = fileDir
	config.numberOfRequests = 2
	testHandler := setupAndListen(config)

	postBodyInBytes := hex.EncodeToString([]byte("Data"))
	methodReceived := testHandler.lastMethodReceived
	bodyReceived := testHandler.lastBodyReceived

	if methodReceived != method {
		t.Errorf("The HTTP method received by the server didn't match. Expected %s, got %s", method, methodReceived)
	}

	if bodyReceived != postBodyInBytes {
		t.Errorf("The body received by the server didn't match. Expected %s, got %s", postBodyInBytes, bodyReceived)
	}
}

func TestThatHeadersAreSetWhenSendingFromFile(t *testing.T) {
	uri := "http://localhost:" + port
	method := "GET"
	fileContents := method + "," + uri + "," + "" + "," + "Content-Type: Hello, Secret: World"
	fileInBytes := []byte(fileContents)

	fileDir := "test-resources/requests-from-file.txt"
	if ioutil.WriteFile(fileDir, fileInBytes, 0644) != nil {
		t.Errorf("Failed to write a required test case file. Check the directory permissions.")
	}
	defer os.Remove(fileDir)

	config := defaultConfig()
	config.requestsFromFile = fileDir
	config.numberOfRequests = 1
	testHandler := setupAndListen(config)

	headerActual := hex.EncodeToString(testHandler.lastHeadersReceived.Peek("Content-Type"))
	headerExpected := hex.EncodeToString([]byte("Hello"))
	if headerExpected != headerActual {
		t.Errorf("Header not found or improperly set, Expected %s, got %s", headerExpected, headerActual)
	}

	headerActual2 := hex.EncodeToString(testHandler.lastHeadersReceived.Peek("Secret"))
	headerExpected2 := hex.EncodeToString([]byte("World"))
	if headerExpected != headerActual {
		t.Errorf("Header not found or improperly set, Expected %s, got %s", headerExpected2, headerActual2)
	}
}

func TestThatTimeOptionRunsForCorrectAmountOfTime(t *testing.T) {
	duration := 10
	testHandler := startServer()
	timeNow := time.Now().Unix()

	config := defaultConfig()
	config.duration = duration

	baton := &Baton{configuration: config, result: Result{}}
	baton.run()

	time.Sleep(time.Duration(15) * time.Second)
	lastTimeStamp := testHandler.lastTimestamp
	epsilonLow := int64(duration - 1)
	epsilonHigh := int64(duration + 1)
	diff := lastTimeStamp - timeNow
	if diff < epsilonLow || diff > epsilonHigh {
		t.Errorf("Requests sent for longer/shorter than expected. Expected %d, got %d)", duration, diff)
	}
}
