package update

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// freeLock returns a buffered channel pre-loaded with a token, simulating
// "no update currently running".
func freeLock() chan bool {
	ch := make(chan bool, 1)
	ch <- true
	return ch
}

// busyLock returns a buffered channel with no token, simulating
// "an update is already running".
func busyLock() chan bool {
	return make(chan bool, 1)
}

func TestHandle_DoesNotWriteRequestBodyToStdout(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	captured := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		captured <- buf.Bytes()
	}()

	called := false
	h := New(func(images []string) { called = true }, freeLock())

	body := strings.NewReader("this body must not appear on stdout")
	req := httptest.NewRequest(http.MethodPost, "/v1/update", body)
	rr := httptest.NewRecorder()
	h.Handle(rr, req)

	_ = w.Close()
	got := <-captured

	if bytes.Contains(got, []byte("this body must not appear on stdout")) {
		t.Fatalf("request body was written to stdout: %q", string(got))
	}
	if !called {
		t.Fatalf("update function was not invoked")
	}
}

func TestHandle_ImageQuery_NonBlockingWhenBusy(t *testing.T) {
	called := false
	h := New(func(images []string) { called = true }, busyLock())

	req := httptest.NewRequest(http.MethodGet, "/v1/update?image=foo", nil)
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.Handle(rr, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Handle blocked on busy lock for image query")
	}

	if got, want := rr.Code, http.StatusServiceUnavailable; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: got %q, want application/json", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "update already running") {
		t.Fatalf("body: got %q, want it to mention 'update already running'", body)
	}
	if called {
		t.Fatalf("update function was invoked while busy")
	}
}

func TestHandle_NoImage_NonBlockingWhenBusy(t *testing.T) {
	called := false
	h := New(func(images []string) { called = true }, busyLock())

	req := httptest.NewRequest(http.MethodGet, "/v1/update", nil)
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.Handle(rr, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Handle blocked on busy lock for no-image request")
	}

	if got, want := rr.Code, http.StatusServiceUnavailable; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}
	if called {
		t.Fatalf("update function was invoked while busy")
	}
}

func TestHandle_FreeLock_InvokesUpdate(t *testing.T) {
	var gotImages []string
	h := New(func(images []string) { gotImages = images }, freeLock())

	req := httptest.NewRequest(http.MethodGet, "/v1/update?image=alpha,beta&image=gamma", nil)
	rr := httptest.NewRecorder()
	h.Handle(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}
	want := []string{"alpha", "beta", "gamma"}
	if len(gotImages) != len(want) {
		t.Fatalf("images: got %v, want %v", gotImages, want)
	}
	for i := range want {
		if gotImages[i] != want[i] {
			t.Fatalf("images[%d]: got %q, want %q", i, gotImages[i], want[i])
		}
	}
}
