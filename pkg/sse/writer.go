package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Writer handles Server-Sent Events writing
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewWriter creates a new SSE writer
func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	return &Writer{
		w:       w,
		flusher: flusher,
	}, nil
}

// WriteEvent writes a data event
func (w *Writer) WriteEvent(data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w.w, "data: %s\n\n", jsonData)
	if err != nil {
		return err
	}

	w.flusher.Flush()
	return nil
}

// WriteDone writes the final [DONE] event
func (w *Writer) WriteDone() error {
	_, err := fmt.Fprint(w.w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}

	w.flusher.Flush()
	return nil
}
