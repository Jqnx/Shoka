package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JSON sets the JSON content type, writes status, then encodes data as the
// body. The status is already committed, so an encoding failure cannot be
// reported to the client and is deliberately ignored.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// SSEHeaders sets the headers required for a Server-Sent Events stream and
// writes the 200 status. Call once, before writing any events.
func SSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

// SSEEvent writes a single named Server-Sent Event with a JSON-encoded
// payload. Does not flush — call http.Flusher.Flush() after (or batch a
// few events before flushing).
func SSEEvent(w http.ResponseWriter, event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)

	return err
}

// NoContent replies 204 with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// File replies 200 with data under contentType, marked immutable so clients
// may cache covers and page images for a year.
func File(w http.ResponseWriter, data []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
