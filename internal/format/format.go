// Package format renders human-readable values for the UI.
package format

import (
	"fmt"
	"time"
)

// Bytes renders a byte count in binary units, e.g. "1.5 KiB".
func Bytes(n int64) string {
	if n < 0 {
		n = 0
	}
	return bytes2(float64(n), "B")
}

func bytes2(v float64, suffix string) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d %s", int64(v), suffix)
	}
	for _, u := range []string{"Ki", "Mi", "Gi", "Ti", "Pi"} {
		v /= unit
		if v < unit {
			return fmt.Sprintf("%.1f %s%s", v, u, suffix)
		}
	}
	return fmt.Sprintf("%.1f E%s", v, suffix)
}

// Bandwidth renders a per-second byte rate, e.g. "1.2 KiB/s".
func Bandwidth(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B/s", n)
	}
	return Bytes(n) + "/s"
}

// Age renders how long ago t was, e.g. "3m12s", "2h05m", "4d07h".
func Age(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd%02dh", int(d.Hours())/24, int(d.Hours())%24)
	}
}

// Delay renders a latency value; 0 means failed or never tested.
func Delay(ms uint16) string {
	if ms == 0 {
		return "timeout"
	}
	return fmt.Sprintf("%dms", ms)
}

// Percent renders n/d as an integer percentage, "-" when d is unknown.
func Percent(n, d int64) string {
	if d <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d%%", n*100/d)
}

// Date renders a unix timestamp as YYYY-MM-DD, "-" when zero.
func Date(unix int64) string {
	if unix <= 0 {
		return "-"
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}
