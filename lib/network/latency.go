package network

import (
	"net"
	"time"

	"google.golang.org/grpc/benchmark/latency"
)

func WrapListener(listener net.Listener, injectLatency time.Duration) net.Listener {
	slowListener := (&latency.Network{Latency: injectLatency}).Listener(listener)
	return slowListener
}

func WrapDialFunc(dialFunc latency.Dialer, injectLatency time.Duration) latency.Dialer {
	slowDialFunc := (&latency.Network{Latency: injectLatency}).Dialer(dialFunc)
	return slowDialFunc
}
