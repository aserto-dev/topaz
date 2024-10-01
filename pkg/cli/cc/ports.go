package cc

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var errTimeout = errors.New("timeout")

type PortStatus int

const (
	PortOpened PortStatus = iota
	PortClosed
)

func (p PortStatus) String() string {
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

func WaitForPorts(ports []string, expectedStatus PortStatus) error {
	t0 := time.Now().UTC()
	log.Debug().Time("started", t0).Msg("WaitForPorts")

	defer func() {
		t1 := time.Now().UTC()
		log.Debug().Time("stopped", t1).Msg("WaitForPorts")

		diff := t1.Sub(t0)
		log.Debug().Str("elapsed", diff.String()).Msg("WaitForPorts")
	}()

	for _, port := range ports {
		listenAddress := fmt.Sprintf("%s:%s", address, port)

		if err := Retry(timeout, interval, func() error {
			status := portStatus(listenAddress)
			if status != expectedStatus {
				log.Debug().Str("addr", listenAddress).Stringer("status", status).Stringer("expected", expectedStatus).Msg("WaitForPorts")
				return errors.Errorf("%s %s", listenAddress, status)
			}
			log.Debug().Str("addr", listenAddress).Stringer("status", status).Stringer("expected", expectedStatus).Msg("WaitForPorts")
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// PortStatus returns true if there's a socket listening on the specified listenAddress.
func portStatus(listenAddress string) PortStatus {
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
