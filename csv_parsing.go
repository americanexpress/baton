package main

import (
	"os"
	"encoding/csv"
	"bufio"
	"io"
	"strings"
	"runtime"
	"errors"
)

func extractHeaders(rawHeaders string) [][]string {
	var headers [][]string
	if rawHeaders != "" {
		header := strings.Split(rawHeaders, "\n")
		for i := 0; i < len(header); i++ {
			headerParts := strings.Split(header[i], ":")
			if len(headerParts) == 2 {
				headers = append(headers, []string{headerParts[0], headerParts[1]})
			}
		}
	}
	return headers
}

func preloadRequestsFromFile(filename string) ([]preloadedRequest, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(file))
	var requests []preloadedRequest

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var method = ""
		var url = ""
		var body = ""
		var headers [][]string
		noFields := len(record)

		if noFields < 2 {
			return nil, errors.New("invalid number of fields")
		}

		if noFields >= 2{
			method = record[0]
			url = record[1]
		}

		if noFields >= 3 {
			body = record[2]
		}

		if noFields >= 4 {
			headers = extractHeaders(record[3])
		}

		requests = append(requests, preloadedRequest{method, url, body, headers})
	}

	return requests, nil
}
