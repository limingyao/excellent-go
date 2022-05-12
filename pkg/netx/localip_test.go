package netx_test

import (
	"testing"

	"github.com/limingyao/excellent-go/pkg/netx"
)

func TestListenAddr(t *testing.T) {
	mac, ip, err := netx.ListenAddr()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(mac)
	t.Log(ip)
}
