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

func preLoadRequestsFromFile(filename string) ([]preLoadedRequest, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(file))
	var requests []preLoadedRequest

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

		requests = append(requests, preLoadedRequest{method, url, body, headers})
	}

	return requests, nil
}
