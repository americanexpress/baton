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
