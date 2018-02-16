package main

import "fmt"

type logWriter struct {
	enabled bool
}

func (writer *logWriter) Disable() {
	writer.enabled = false
}

func (writer *logWriter) Enable() {
	writer.enabled = true
}

func (writer *logWriter) Write(bytes []byte) (int, error) {
	if writer.enabled {
		return fmt.Print(string(bytes))
	}
	return 0, nil
}
