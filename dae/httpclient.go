package dae

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/daeuniverse/dae/common/consts"
	"github.com/daeuniverse/dae/control"
	"github.com/mzz2017/softwind/netproxy"
	"github.com/sirupsen/logrus"
)

var HttpTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("no dns record: %v", host)
		}
		port, err := strconv.ParseUint(_port, 10, 16)
		if err != nil {
			return nil, err
		}
		logrus.Warnln("!!!")
		conn, err := ControlPlane.RouteDialTcp(&control.RouteDialParam{Outbound: consts.OutboundControlPlaneRouting, Domain: host, Mac: [6]uint8{}, ProcessName: [16]uint8{}, Src: netip.MustParseAddrPort("0.0.0.0:0"), Dest: netip.AddrPortFrom(addrs[0], uint16(port)), Mark: 0})
		if err != nil {
			return nil, err
		}
		return &netproxy.FakeNetConn{Conn: conn, LAddr: nil, RAddr: nil}, nil
	},
	TLSHandshakeTimeout:   10 * time.Second,
	DisableKeepAlives:     true,
	DisableCompression:    false,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ForceAttemptHTTP2:     true,
}
