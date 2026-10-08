package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// streamLines connects to a chunked JSON-lines endpoint and calls fn for each
// line until ctx is canceled; on connection errors it retries with backoff.
// It is meant to run in its own goroutine.
func (c *Client) streamLines(ctx context.Context, path string, fn func(line []byte)) {
	backoff := c.retryBackoff
	for ctx.Err() == nil {
		c.streamOnce(ctx, path, fn)
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 8*time.Second {
			backoff *= 2
		}
	}
}

func (c *Client) streamOnce(ctx context.Context, path string, fn func([]byte)) error {
	req, err := c.newReq(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.streamHC.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &APIError{Status: resp.StatusCode, Message: decodeMessage(data, resp.StatusCode)}
	}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		fn(line)
	}
	return sc.Err()
}

// StreamTraffic subscribes to /traffic until ctx is canceled, retrying on
// errors. Runs in the caller's goroutine; intended use: `go client.StreamTraffic(...)`.
func (c *Client) StreamTraffic(ctx context.Context, fn func(Traffic, error)) {
	c.streamLines(ctx, "/traffic", func(line []byte) {
		var t Traffic
		if err := json.Unmarshal(line, &t); err == nil {
			fn(t, nil)
		}
	})
}

// StreamMemory subscribes to /memory until ctx is canceled.
func (c *Client) StreamMemory(ctx context.Context, fn func(Memory, error)) {
	c.streamLines(ctx, "/memory", func(line []byte) {
		var m Memory
		if err := json.Unmarshal(line, &m); err == nil {
			fn(m, nil)
		}
	})
}

// StreamLogs subscribes to /logs at the given level until ctx is canceled.
func (c *Client) StreamLogs(ctx context.Context, level string, fn func(LogLine, error)) {
	path := "/logs?level=" + url.QueryEscape(strings.TrimSpace(level))
	c.streamLines(ctx, path, func(line []byte) {
		var l LogLine
		if err := json.Unmarshal(line, &l); err == nil {
			fn(l, nil)
		}
	})
}
