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
	"errors"
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
	suppressOutput   bool
	url              string
	wait             int
}

func (configuration *Configuration) validate() error {

	if configuration.concurrency < 1 || configuration.numberOfRequests == 0 {
		return errors.New("invalid concurrency level or number of requests")
	}

	return nil
}
