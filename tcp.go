package gost

import (
	"errors"
	"net"
	"time"

	"github.com/go-log/log"
)

// tcpTransporter is a raw TCP transporter.
type tcpTransporter struct{}

// TCPTransporter creates a raw TCP client.
func TCPTransporter() Transporter {
	return &tcpTransporter{}
}

func (tr *tcpTransporter) Dial(addr string, options ...DialOption) (net.Conn, error) {
	opts := &DialOptions{}
	for _, option := range options {
		option(opts)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DialTimeout
	}
	if opts.Chain == nil {
		rAddr, _ := net.ResolveTCPAddr("tcp", addr)

		// ipv4
		if rAddr == nil || rAddr.IP.To4() != nil {
			conn, err := net.DialTimeout("tcp", addr, timeout)
			log.Logf("[tcp]dial %s from ipv4, err: %+v", addr, err)
			return conn, err
		}

		// ipv6 使用随机出口ip
		var (
			conn *net.TCPConn
			err  error
		)

		done := make(chan bool, 1)
		go func() {
			lAddr := GetLocalAddress()
			conn, err = net.DialTCP("tcp", lAddr, rAddr)
			log.Logf("[tcp]dial %s from %s, err: %+v", addr, lAddr.IP.String(), err)
			done <- true
		}()

		select {
		case <-time.After(timeout):
			log.Logf("[ws]dial timeout %s", addr)
			return nil, errors.New("dial timeout")
		case <-done:
			//
		}

		return conn, err
	}
	return opts.Chain.Dial(addr)
}

func (tr *tcpTransporter) Handshake(conn net.Conn, options ...HandshakeOption) (net.Conn, error) {
	return conn, nil
}

func (tr *tcpTransporter) Multiplex() bool {
	return false
}

type tcpListener struct {
	net.Listener
}

// TCPListener creates a Listener for TCP proxy server.
func TCPListener(addr string) (Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &tcpListener{Listener: tcpKeepAliveListener{ln}}, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(KeepAliveTime)
	return tc, nil
}
