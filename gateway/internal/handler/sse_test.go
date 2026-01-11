package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/logger"
)

// ============================================================
// SSE Handler Tests
// ============================================================

func createTestSSEHandler() *SSEHandler {
	broker := events.NewBroker()
	broker.Start()
	log := logger.NewDefault("test")
	return NewSSEHandler(broker, log)
}

func TestSSEHandler_HandleStats(t *testing.T) {
	handler := createTestSSEHandler()

	t.Run("returns stats with 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sse/stats", nil)
		rec := httptest.NewRecorder()

		handler.HandleStats(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		// Verify response structure
		var stats map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &stats)
		require.NoError(t, err)
		assert.Contains(t, stats, "connected_clients")
		assert.Contains(t, stats, "status")
		assert.Equal(t, "healthy", stats["status"])
	})
}

func TestSSEHandler_HandleBroadcast(t *testing.T) {
	handler := createTestSSEHandler()

	t.Run("broadcast with GET returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sse/broadcast", nil)
		rec := httptest.NewRecorder()

		handler.HandleBroadcast(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("broadcast with JSON body returns 200", func(t *testing.T) {
		body := map[string]interface{}{
			"topic": "transactions",
			"type":  "new_transaction",
			"data": map[string]interface{}{
				"amount": 1000,
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/sse/broadcast", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleBroadcast(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		// Verify response structure
		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, true, resp["success"])
		assert.Contains(t, resp, "message")
		assert.Contains(t, resp, "clients")
	})

	t.Run("broadcast with invalid JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/sse/broadcast", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleBroadcast(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("broadcast with query params returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/sse/broadcast?topic=wallet&type=balance_update&message=Balance+updated", nil)
		rec := httptest.NewRecorder()

		handler.HandleBroadcast(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, true, resp["success"])
	})

	t.Run("broadcast with defaults returns 200", func(t *testing.T) {
		// POST with empty body and no query params - should use defaults
		body := map[string]interface{}{}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/sse/broadcast", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleBroadcast(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestSSEHandler_HandleEvents(t *testing.T) {
	handler := createTestSSEHandler()

	t.Run("sets correct SSE headers", func(t *testing.T) {
		// Create a context that we can cancel
		ctx, cancel := context.WithCancel(context.Background())

		req := httptest.NewRequest(http.MethodGet, "/sse/events", nil)
		req = req.WithContext(ctx)
		req.Header.Set("X-Request-ID", "test-request-123")

		// Create a mock response recorder that implements http.Flusher
		rec := &mockFlusherRecorder{ResponseRecorder: httptest.NewRecorder()}

		// Start the handler in a goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)
			handler.HandleEvents(rec, req)
		}()

		// Give the handler time to set headers and start
		time.Sleep(50 * time.Millisecond)

		// Cancel the context to stop the handler
		cancel()

		// Wait for handler to finish
		<-done

		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
		assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("handles streaming unsupported error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sse/events", nil)
		// Use nonFlusherRecorder which doesn't implement http.Flusher
		rec := &nonFlusherRecorder{rw: httptest.NewRecorder()}

		handler.HandleEvents(rec, req)

		// Should return error for non-flusher response writer
		assert.Equal(t, http.StatusInternalServerError, rec.Code())
		assert.Contains(t, rec.Body().String(), "Streaming unsupported")
	})
}

// mockFlusherRecorder is a ResponseRecorder that implements http.Flusher.
type mockFlusherRecorder struct {
	*httptest.ResponseRecorder
}

func (m *mockFlusherRecorder) Flush() {
	// No-op for testing
}

// nonFlusherRecorder wraps ResponseRecorder without implementing http.Flusher.
// By not embedding and only composing, the Flusher interface is not promoted.
type nonFlusherRecorder struct {
	rw *httptest.ResponseRecorder
}

func (n *nonFlusherRecorder) Header() http.Header {
	return n.rw.Header()
}

func (n *nonFlusherRecorder) Write(b []byte) (int, error) {
	return n.rw.Write(b)
}

func (n *nonFlusherRecorder) WriteHeader(statusCode int) {
	n.rw.WriteHeader(statusCode)
}

// Getter for Body to access in tests
func (n *nonFlusherRecorder) Body() *bytes.Buffer {
	return n.rw.Body
}

// Getter for Code to access in tests
func (n *nonFlusherRecorder) Code() int {
	return n.rw.Code
}
