package yookassa

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

type socks5Dialer struct {
	address  string
	username string
	password string
	dialer   *net.Dialer
}

func newSocks5Dialer(address, username, password string) *socks5Dialer {
	return &socks5Dialer{
		address:  address,
		username: username,
		password: password,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
}

func (d *socks5Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.New("proxy: no support for SOCKS5 proxy connections of type " + network)
	}

	base := d.dialer
	if base == nil {
		base = &net.Dialer{}
	}

	conn, err := base.DialContext(ctx, "tcp", d.address)
	if err != nil {
		return nil, err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}

	if err := d.connect(conn, addr); err != nil {
		conn.Close()
		return nil, err
	}

	conn.SetDeadline(time.Time{})
	return conn, nil
}

const (
	socks5Version      = 5
	socks5AuthNone     = 0
	socks5AuthPassword = 2
	socks5Connect      = 1
	socks5IP4          = 1
	socks5Domain       = 3
	socks5IP6          = 4
)

var socks5Errors = []string{
	"",
	"general failure",
	"connection forbidden",
	"network unreachable",
	"host unreachable",
	"connection refused",
	"TTL expired",
	"command not supported",
	"address type not supported",
}

func (d *socks5Dialer) connect(conn net.Conn, target string) error {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("proxy: failed to parse port number: " + portStr)
	}
	if port < 1 || port > 0xffff {
		return errors.New("proxy: port number out of range: " + portStr)
	}

	buf := make([]byte, 0, 6+len(host))
	buf = append(buf, socks5Version)
	if len(d.username) > 0 && len(d.username) < 256 && len(d.password) < 256 {
		buf = append(buf, 2, socks5AuthNone, socks5AuthPassword)
	} else {
		buf = append(buf, 1, socks5AuthNone)
	}

	if _, err := conn.Write(buf); err != nil {
		return errors.New("proxy: failed to write greeting to SOCKS5 proxy at " + d.address + ": " + err.Error())
	}

	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return errors.New("proxy: failed to read greeting from SOCKS5 proxy at " + d.address + ": " + err.Error())
	}
	if buf[0] != socks5Version {
		return errors.New("proxy: SOCKS5 proxy at " + d.address + " has unexpected version " + strconv.Itoa(int(buf[0])))
	}
	if buf[1] == 0xff {
		return errors.New("proxy: SOCKS5 proxy at " + d.address + " requires authentication")
	}

	if buf[1] == socks5AuthPassword {
		if len(d.username) == 0 || len(d.username) >= 256 || len(d.password) >= 256 {
			return errors.New("proxy: SOCKS5 proxy at " + d.address + " requires username/password")
		}
		buf = buf[:0]
		buf = append(buf, 1)
		buf = append(buf, byte(len(d.username)))
		buf = append(buf, d.username...)
		buf = append(buf, byte(len(d.password)))
		buf = append(buf, d.password...)
		if _, err := conn.Write(buf); err != nil {
			return errors.New("proxy: failed to write authentication request to SOCKS5 proxy at " + d.address + ": " + err.Error())
		}
		if _, err := io.ReadFull(conn, buf[:2]); err != nil {
			return errors.New("proxy: failed to read authentication reply from SOCKS5 proxy at " + d.address + ": " + err.Error())
		}
		if buf[1] != 0 {
			return errors.New("proxy: SOCKS5 proxy at " + d.address + " rejected username/password")
		}
	}

	buf = buf[:0]
	buf = append(buf, socks5Version, socks5Connect, 0)
	ip := net.ParseIP(host)
	if ip4 := ip.To4(); ip4 != nil {
		buf = append(buf, socks5IP4)
		buf = append(buf, ip4...)
	} else if ip6 := ip.To16(); ip6 != nil {
		buf = append(buf, socks5IP6)
		buf = append(buf, ip6...)
	} else {
		if len(host) > 255 {
			return errors.New("proxy: destination host name too long: " + host)
		}
		buf = append(buf, socks5Domain)
		buf = append(buf, byte(len(host)))
		buf = append(buf, host...)
	}
	buf = append(buf, byte(port>>8), byte(port))

	if _, err := conn.Write(buf); err != nil {
		return errors.New("proxy: failed to write connect request to SOCKS5 proxy at " + d.address + ": " + err.Error())
	}

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return errors.New("proxy: failed to read connect reply from SOCKS5 proxy at " + d.address + ": " + err.Error())
	}
	if int(buf[1]) < len(socks5Errors) {
		if failure := socks5Errors[buf[1]]; failure != "" {
			return errors.New("proxy: SOCKS5 proxy at " + d.address + " failed to connect: " + failure)
		}
	}

	switch buf[3] {
	case socks5IP4:
		buf = buf[:4]
	case socks5IP6:
		buf = buf[:16]
	case socks5Domain:
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return errors.New("proxy: failed to read domain length from SOCKS5 proxy at " + d.address + ": " + err.Error())
		}
		buf = buf[:int(buf[0])]
	default:
		return errors.New("proxy: got unknown address type " + strconv.Itoa(int(buf[3])) + " from SOCKS5 proxy at " + d.address)
	}

	if _, err := io.ReadFull(conn, buf); err != nil {
		return errors.New("proxy: failed to read address from SOCKS5 proxy at " + d.address + ": " + err.Error())
	}
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return errors.New("proxy: failed to read port from SOCKS5 proxy at " + d.address + ": " + err.Error())
	}

	return nil
}
