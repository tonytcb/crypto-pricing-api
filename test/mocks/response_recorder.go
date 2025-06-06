package mocks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
)

// ThreadSafeRecorder is a thread-safe ResponseRecorder
type ThreadSafeRecorder struct {
	mu sync.Mutex
	w  *httptest.ResponseRecorder
}

func NewThreadSafeRecorder() *ThreadSafeRecorder {
	return &ThreadSafeRecorder{
		w: httptest.NewRecorder(),
	}
}

func (t *ThreadSafeRecorder) Recorder() *httptest.ResponseRecorder {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w
}

func (t *ThreadSafeRecorder) Header() http.Header {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Header()
}

func (t *ThreadSafeRecorder) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Write(b)
}

func (t *ThreadSafeRecorder) WriteHeader(statusCode int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.w.WriteHeader(statusCode)
}

func (t *ThreadSafeRecorder) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.w.Flush()
}

func (t *ThreadSafeRecorder) Result() *http.Response {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Result()
}

func (t *ThreadSafeRecorder) Body() *bytes.Buffer {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Body
}

func (t *ThreadSafeRecorder) Code() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Code
}

func (t *ThreadSafeRecorder) BodyString() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Body.String()
}

func (t *ThreadSafeRecorder) GetAllHeaders() http.Header {
	t.mu.Lock()
	defer t.mu.Unlock()

	headers := make(http.Header)
	for k, vv := range t.w.Header() {
		values := make([]string, len(vv))
		copy(values, vv)
		headers[k] = values
	}

	return headers
}

func (t *ThreadSafeRecorder) GetHeader(key string) string {
	headers := t.GetAllHeaders()
	return headers.Get(key)
}

func (t *ThreadSafeRecorder) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.w.Body.Reset()
}
