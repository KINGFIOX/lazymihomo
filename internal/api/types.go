package api

import "time"

// Proxy is one proxy or policy-group object from GET /proxies or GET /group.
type Proxy struct {
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	UDP          bool           `json:"udp"`
	Alive        bool           `json:"alive"`
	History      []DelayHistory `json:"history"`
	Now          string         `json:"now"` // groups only: current selection
	All          []string       `json:"all"` // groups only: member names
	ProviderName string         `json:"provider-name"`
}

// DelayHistory is one latency sample.
type DelayHistory struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
}

// IsGroup reports whether the proxy is a policy group (has members).
func (p Proxy) IsGroup() bool { return len(p.All) > 0 }

// LastDelay returns the most recent latency sample, 0 if none.
func (p Proxy) LastDelay() uint16 {
	if len(p.History) == 0 {
		return 0
	}
	return p.History[len(p.History)-1].Delay
}

// RecordDelay appends a latency sample locally after a delay test.
func (p *Proxy) RecordDelay(ms uint16) {
	p.History = append(p.History, DelayHistory{Time: time.Now(), Delay: ms})
	if len(p.History) > 10 {
		p.History = p.History[len(p.History)-10:]
	}
}

type proxiesResp struct {
	Proxies map[string]Proxy `json:"proxies"`
}

type groupsResp struct {
	Proxies []Proxy `json:"proxies"`
}

// Connection is one live connection tracker.
type Connection struct {
	ID          string    `json:"id"`
	Metadata    Metadata  `json:"metadata"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	Start       time.Time `json:"start"`
	Chains      []string  `json:"chains"`
	Rule        string    `json:"rule"`
	RulePayload string    `json:"rulePayload"`
}

// Metadata describes the endpoints of a connection.
type Metadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	DNSMode         string `json:"dnsMode"`
	Process         string `json:"process"`
	ProcessPath     string `json:"processPath"`
}

// Hostport returns "host:port", preferring the sniffed host over the raw IP.
func (m Metadata) Hostport() string {
	host := m.Host
	if host == "" {
		host = m.DestinationIP
	}
	if host == "" {
		return ""
	}
	if m.DestinationPort == "" {
		return host
	}
	return host + ":" + m.DestinationPort
}

// ConnectionsSnapshot is the GET /connections response.
type ConnectionsSnapshot struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Memory        uint64       `json:"memory"`
	Connections   []Connection `json:"connections"`
}

// RuleExtra carries optional per-rule hit statistics.
type RuleExtra struct {
	HitCount uint64    `json:"hitCount"`
	HitAt    time.Time `json:"hitAt"`
}

// Rule is one parsed routing rule.
type Rule struct {
	Index   int        `json:"index"`
	Type    string     `json:"type"`
	Payload string     `json:"payload"`
	Proxy   string     `json:"proxy"`
	Size    int64      `json:"size"`
	Extra   *RuleExtra `json:"extra"`
}

type rulesResp struct {
	Rules []Rule `json:"rules"`
}

// SubscriptionInfo carries the quota of a subscription provider.
// Field names are capitalized on the wire (no json tags upstream).
type SubscriptionInfo struct {
	Upload   int64 `json:"Upload"`
	Download int64 `json:"Download"`
	Total    int64 `json:"Total"`
	Expire   int64 `json:"Expire"`
}

// Provider is a proxy provider (subscription source).
type Provider struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	VehicleType      string            `json:"vehicleType"`
	Proxies          []Proxy           `json:"proxies"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	SubscriptionInfo *SubscriptionInfo `json:"subscriptionInfo"`
}

type providersResp struct {
	Providers map[string]Provider `json:"providers"`
}

// Traffic is one /traffic stream sample (bytes per second and totals).
type Traffic struct {
	Up        int64 `json:"up"`
	Down      int64 `json:"down"`
	UpTotal   int64 `json:"upTotal"`
	DownTotal int64 `json:"downTotal"`
}

// Memory is one /memory stream sample.
type Memory struct {
	Inuse   uint64 `json:"inuse"`
	OSLimit uint64 `json:"oslimit"`
}

// LogLine is one /logs stream record.
type LogLine struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Version is the GET /version response.
type Version struct {
	Meta    bool   `json:"meta"`
	Version string `json:"version"`
}
