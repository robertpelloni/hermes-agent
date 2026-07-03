package agent

import (
	"context"
	"fmt"
)

// BasicEventLoop is a simple message processing loop designed to
// accept user inputs, process them, and output responses as a stub
// for integration with the Python backend.
type BasicEventLoop struct {
	InputChan  chan string
	OutputChan chan string
}

// NewBasicEventLoop creates a new BasicEventLoop.
func NewBasicEventLoop() *BasicEventLoop {
	return &BasicEventLoop{
		InputChan:  make(chan string, 10),
		OutputChan: make(chan string, 10),
	}
}

// Start begins the event loop.
func (l *BasicEventLoop) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-l.InputChan:
				// Stub: process message through agent logic and produce response
				response := fmt.Sprintf("Processed via Go agent stub: %s", msg)
				l.OutputChan <- response
			}
		}
	}()
}
