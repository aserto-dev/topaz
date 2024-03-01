package cc

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type portStatus int

const (
	PortOpened portStatus = iota
	PortClosed
)

func (p portStatus) String() string {
	switch p {
	case PortClosed:
		return "closed"
	case PortOpened:
		return "opened"
	default:
		return ""
	}
}

const (
	timeout  = 10 * time.Second
	interval = 10 * time.Millisecond
	address  = "0.0.0.0"
)

func WaitForPorts(ports []string, expectedStatus portStatus) error {
	for _, port := range ports {
		listenAddress := fmt.Sprintf("%s:%s", address, port)

		if err := Retry(timeout, interval, func() error {
			if status := PortStatus(listenAddress); status != expectedStatus {
				return fmt.Errorf("%s %s", listenAddress, status)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// PortStatus returns true if there's a socket listening on the specified listenAddress.
func PortStatus(listenAddress string) portStatus {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", listenAddress, timeout)
	if err != nil {
		return PortClosed
	}
	if conn != nil {
		defer conn.Close()
		return PortOpened
	}
	return PortClosed
}

var errTimeout = errors.New("timeout")

func Retry(timeout, interval time.Duration, f func() error) (err error) {
	err = errTimeout
	for t := time.After(timeout); ; {
		select {
		case <-t:
			return
		default:
		}

		err = f()
		if err == nil {
			return
		}

		if timeout > 0 {
			time.Sleep(interval)
		}
	}
}
