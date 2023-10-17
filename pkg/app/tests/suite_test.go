package engine_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	ginkgoT *testing.T
)

func TestEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	ginkgoT = t
	RunSpecs(t, "Engine Suite")
}
