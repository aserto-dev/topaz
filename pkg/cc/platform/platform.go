package platform

import "runtime"

func IsArm64() bool {
	return runtime.GOARCH == "arm64"
}

func IsAmd64() bool {
	return runtime.GOARCH == "amd64"
}
