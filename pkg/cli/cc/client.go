package cc

import (
	"os"
	"strconv"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/x"
)

const (
	defaultDirectorySvc    = "localhost:9292"
	defaultDirectoryKey    = ""
	defaultDirectoryToken  = ""
	defaultAuthorizerSvc   = "localhost:8282"
	defaultAuthorizerKey   = ""
	defaultAuthorizerToken = ""
	defaultTenantID        = ""
	defaultInsecure        = false
	defaultPlaintext       = false
	defaultTimeout         = 5 * time.Second
	defaultNoCheck         = false
	defaultNoColor         = false
)

func DirectorySvc() string {
	if directorySvc := os.Getenv(x.EnvTopazDirectorySvc); directorySvc != "" {
		return directorySvc
	}

	return defaultDirectorySvc
}

func DirectoryKey() string {
	if directoryKey := os.Getenv(x.EnvTopazDirectoryKey); directoryKey != "" {
		return directoryKey
	}

	return defaultDirectoryKey
}

func DirectoryToken() string {
	if directoryToken := os.Getenv(x.EnvTopazDirectoryToken); directoryToken != "" {
		return directoryToken
	}

	return defaultDirectoryToken
}

func AuthorizerSvc() string {
	if authorizerSvc := os.Getenv(x.EnvTopazAuthorizerSvc); authorizerSvc != "" {
		return authorizerSvc
	}

	return defaultAuthorizerSvc
}

func AuthorizerKey() string {
	if authorizerKey := os.Getenv(x.EnvTopazAuthorizerKey); authorizerKey != "" {
		return authorizerKey
	}

	return defaultAuthorizerKey
}

func AuthorizerToken() string {
	if authorizerToken := os.Getenv(x.EnvTopazAuthorizerToken); authorizerToken != "" {
		return authorizerToken
	}

	return defaultAuthorizerToken
}

func TenantID() string {
	if tenantID := os.Getenv(x.EnvAsertoTenantID); tenantID != "" {
		return tenantID
	}

	return defaultTenantID
}

func Insecure() bool {
	if insecure := os.Getenv(x.EnvTopazInsecure); insecure != "" {
		if b, err := strconv.ParseBool(insecure); err == nil {
			return b
		}
	}

	return defaultInsecure
}

func Plaintext() bool {
	if plaintext := os.Getenv(x.EnvTopazPlaintext); plaintext != "" {
		if b, err := strconv.ParseBool(plaintext); err == nil {
			return b
		}
	}

	return defaultPlaintext
}

func Timeout() time.Duration {
	if timeout := os.Getenv(x.EnvTopazTimeout); timeout != "" {
		if dur, err := time.ParseDuration(timeout); err == nil {
			return dur
		}
	}

	return defaultTimeout
}

func NoCheck() bool {
	if noCheck := os.Getenv(x.EnvTopazNoCheck); noCheck != "" {
		if b, err := strconv.ParseBool(noCheck); err == nil {
			return b
		}
	}

	return defaults.NoCheck
}

func NoColor() bool {
	if noColor := os.Getenv(x.EnvTopazNoColor); noColor != "" {
		if b, err := strconv.ParseBool(noColor); err == nil {
			return b
		}
	}

	return defaults.NoColor
}
