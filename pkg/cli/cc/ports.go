package cc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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
			status := PortStatus(listenAddress)
			if status != expectedStatus {
				log.Debug().Str("addr", listenAddress).Stringer("status", status).Stringer("expected", expectedStatus).Msg("WaitForPorts")
				return fmt.Errorf("%s %s", listenAddress, status)
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

// ServiceHealthStatus adopted from grpc-health-probe cli implementation
// https://github.com/grpc-ecosystem/grpc-health-probe/blob/master/main.go.
func ServiceHealthStatus(service string) bool {
	addr := "localhost:9494"
	connTimeout := time.Second * 30
	rpcTimeout := time.Second * 30

	bCtx := context.Background()
	dialCtx, dialCancel := context.WithTimeout(bCtx, connTimeout)
	defer dialCancel()

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(dialCtx, addr, dialOpts...)
	if err != nil {
		return false
	}
	defer conn.Close()

	rpcCtx, rpcCancel := context.WithTimeout(bCtx, rpcTimeout)
	defer rpcCancel()

	if err := Retry(rpcTimeout, time.Millisecond*100, func() error {
		resp, err := healthpb.NewHealthClient(conn).Check(rpcCtx, &healthpb.HealthCheckRequest{Service: service})
		if err != nil {
			return err
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			return fmt.Errorf("gRPC endpoint not SERVING")
		}

		return nil
	}); err != nil {
		return false
	}

	return true
}
