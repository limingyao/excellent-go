package net_test

import (
	"testing"

	"github.com/limingyao/excellent-go/pkg/net"
)

func TestListenAddr(t *testing.T) {
	mac, ip, err := net.ListenAddr()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(mac)
	t.Log(ip)
}
