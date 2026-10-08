package format

import (
	"testing"
	"time"
)

func TestBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{2 << 20, "2.0 MiB"},
		{3<<30 + 500<<20, "3.5 GiB"},
	}
	for _, c := range cases {
		if got := Bytes(c.in); got != c.want {
			t.Errorf("Bytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBandwidth(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B/s"},
		{999, "999 B/s"},
		{1536, "1.5 KiB/s"},
	}
	for _, c := range cases {
		if got := Bandwidth(c.in); got != c.want {
			t.Errorf("Bandwidth(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAge(t *testing.T) {
	now := time.Now()
	cases := []struct {
		ago  time.Duration
		want string
	}{
		{5 * time.Second, "5s"},
		{90 * time.Second, "1m30s"},
		{2*time.Hour + 5*time.Minute, "2h05m"},
		{26 * time.Hour, "1d02h"},
	}
	for _, c := range cases {
		if got := Age(now.Add(-c.ago)); got != c.want {
			t.Errorf("Age(-%v) = %q, want %q", c.ago, got, c.want)
		}
	}
}

func TestDelay(t *testing.T) {
	if Delay(0) != "timeout" {
		t.Errorf("Delay(0) = %q", Delay(0))
	}
	if Delay(123) != "123ms" {
		t.Errorf("Delay(123) = %q", Delay(123))
	}
}

func TestPercent(t *testing.T) {
	if got := Percent(1, 4); got != "25%" {
		t.Errorf("Percent(1,4) = %q", got)
	}
	if got := Percent(1, 0); got != "-" {
		t.Errorf("Percent(1,0) = %q", got)
	}
}

func TestDate(t *testing.T) {
	if got := Date(0); got != "-" {
		t.Errorf("Date(0) = %q", got)
	}
	if got := Date(1762000000); got != "2025-11-01" {
		t.Errorf("Date(1762000000) = %q", got)
	}
}
