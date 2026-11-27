package topaz

import "google.golang.org/grpc"

// Test-only access to package internals for the external test package.
const DefRequestTimeout = defRequestTimeout

var (
	GetConn            = getConn
	CloseConn          = closeConn
	TimeoutInterceptor = timeoutInterceptor
)

func ResetConfig() { aCfg.Store(nil) }

func CurrentConn() *grpc.ClientConn {
	connMu.Lock()
	defer connMu.Unlock()

	return conn
}

func PluginConfig(p *Plugin) *Config { return p.config }
