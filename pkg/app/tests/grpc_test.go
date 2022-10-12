package engine_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/testing"
)

var _ = Describe("Engine GRPC Server", func() {
	Context("with offline bundle file", func() {

		var (
			h *testing.EngineHarness
		)

		BeforeEach(func() {
			h = testing.SetupOffline(ginkgoT, func(cfg *config.Config) {
				cfg.Directory.EdgeConfig.DBPath = testing.AssetAcmeEBBFilePath()
				cfg.OPA.LocalBundles.Paths = []string{testing.AssetLocalBundle()}
			})
		})

		AfterEach(func() {
			h.Cleanup()
		})

		Context("when using the grpc server", func() {
			It("works", func() {
				Expect(true).To(BeTrue())
			})
		})
	})
})
