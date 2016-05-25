package utils

import (
	"net"
	"time"
)

// TimerSet contains IPs add with by timestamp
type TimerSet struct {
	maxDuration time.Duration
	ips         []ipTime
}

type ipTime struct {
	Timestamp time.Time
	IP        net.IP
}

// NewTimerSet return new instance of a TimeSet
func NewTimerSet(maxDuration time.Duration) *TimerSet {
	return &TimerSet{maxDuration: maxDuration, ips: []ipTime{}}
}

// AddIP add and IP with is timestamp
func (t *TimerSet) AddIP(ip net.IP, now time.Time) {
	t.ips = append(t.ips, ipTime{Timestamp: now, IP: ip})
	t.cleanSet(now)
}

func (t *TimerSet) cleanSet(now time.Time) {
	for i := len(t.ips) - 1; i >= 0; i-- {
		if now.Sub(t.ips[i].Timestamp) > t.maxDuration {
			t.ips = append(t.ips[:0], t.ips[i+1:]...)
			return
		}
	}
}

// FindIP returns true if the IP is present in the list
func (t *TimerSet) FindIP(ip net.IP) bool {
	for _, ipTime := range t.ips {
		if ip.Equal(ipTime.IP) {
			return true
		}
	}
	return false
}

// Front return last IP
func (t TimerSet) Front() net.IP {
	var output net.IP

	if len(t.ips) > 0 {
		output = t.ips[len(t.ips)-1].IP
	}

	return output
}

func (t TimerSet) String() string {
	output := "["

	for _, timer := range t.ips {
		output += timer.IP.String()
		output += ","
	}

	output += "]"

	return output
}

// Debug print debug information
func (t TimerSet) Debug() string {
	output := "["

	for _, timer := range t.ips {
		output += timer.IP.String()
		output += "("
		output += timer.Timestamp.String()
		output += ")"
		output += ","
	}

	output += "]"

	return output
}
