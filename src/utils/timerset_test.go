package utils

import (
	"net"
	"testing"
	"time"
)

func TestTimersetString(t *testing.T) {
	timer := NewTimerSet(time.Minute * 3)

	if timer.String() != "[]" {
		t.Error("invalid timer.String value, current", timer.String())
	}

	if timer.Front() != nil {
		t.Error("Front should be nul")
	}

	timer.AddIP(net.ParseIP("10.0.0.0"), time.Now())

	t.Log(timer.Front().String())
	if timer.String() != "[10.0.0.0,]" {
		t.Error("invalid timer.String value, current", timer.String())
	}

	if timer.Front() == nil {
		t.Error("Front should be nul")
	}

	if timer.Front().String() != "10.0.0.0" {
		t.Error("invalid timer.Front().String() value, current", timer.Front().String())
	}
}
