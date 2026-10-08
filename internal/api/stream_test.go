package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStreamTraffic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/traffic" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fl := w.(http.Flusher)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, `{"up":%d,"down":%d,"upTotal":%d,"downTotal":%d}`+"\n", i, i*2, i*10, i*20)
			fl.Flush()
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	c.retryBackoff = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan Traffic, 8)
	done := make(chan struct{})
	go func() {
		n := 0
		c.StreamTraffic(ctx, func(tr Traffic, err error) {
			if err != nil {
				return
			}
			ch <- tr
			n++
			if n == 3 {
				close(done)
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for samples")
	}
	cancel()
	got := []Traffic{<-ch, <-ch, <-ch}
	if got[2].Down != 4 || got[2].DownTotal != 40 {
		t.Errorf("last sample = %+v", got[2])
	}
}

func TestStreamReconnects(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) == 1 {
			// First connection: deliver one line then kill the connection.
			fmt.Fprintln(w, `{"inuse":10,"oslimit":0}`)
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		}
		fmt.Fprintln(w, `{"inuse":20,"oslimit":0}`)
		w.(http.Flusher).Flush()
		// Hold the second connection open until the client goes away.
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	c.retryBackoff = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		n := 0
		c.StreamMemory(ctx, func(m Memory, err error) {
			if err != nil {
				return
			}
			n++
			if n == 2 {
				close(done)
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stream did not reconnect")
	}
	cancel()
}

func TestStreamLogsLevelParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("level"); got != "warning" {
			t.Errorf("level = %q", got)
		}
		fmt.Fprintln(w, `{"type":"warning","payload":"boom"}`)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go c.StreamLogs(ctx, "warning", func(l LogLine, err error) {
		if err != nil {
			return
		}
		if l.Type == "warning" {
			close(done)
		}
	})

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("no log line received")
	}
	cancel()
}
