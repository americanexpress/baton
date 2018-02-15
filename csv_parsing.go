package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"
)

func extractHeaders(rawHeaders string) []string {
	headerParts := strings.Split(rawHeaders, ":")
	if len(headerParts) == 2 {
		return []string{headerParts[0], headerParts[1]}
	}
	return nil
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

		if noFields >= 2 {
			method = record[0]
			url = record[1]
		}

		if noFields >= 3 {
			body = record[2]
		}

		if noFields >= 4 {
			for i := 3; i < noFields; i++ {
				extractedHeaders := extractHeaders(record[i])
				if extractedHeaders != nil {
					headers = append(headers, extractedHeaders)
				}
			}
		}

		requests = append(requests, preloadedRequest{method, url, body, headers})
	}

	return requests, nil
}
