package runtime_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
)

var (
	ginkgoT *testing.T
)

func TestEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	ginkgoT = t
	RunSpecs(t, "Engine Suite")
}

var _ = Describe("Engine Runtime", func() {
	Context("with offline bundle file", func() {

		var (
			h *atesting.EngineHarness
		)

		BeforeSuite(func() {
			h = atesting.SetupOffline(ginkgoT, func(cfg *config.Config) {
				cfg.Edge.DBPath = atesting.AssetAcmeDBFilePath()
				cfg.OPA.LocalBundles.Paths = []string{atesting.AssetLocalBundle()}
			})
		})

		AfterSuite(func() {
			h.Cleanup()
		})

		Context("when using the engine directly", func() {
			Context("with basic functionality", func() {
				It("successfully returns runtime info", func() {
					result, err := h.Runtime().Query(h.Context(), "x := opa.runtime()", nil, true, true, true, "full")

					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())
					Expect(result.Result[0].Bindings).To(HaveKey("x"))

					Expect(result.Result[0].Bindings["x"]).To(HaveKey("env"))
					x := result.Result[0].Bindings["x"]
					Expect(x).ToNot(BeNil())
				})

				It("successfully returns bundle data", func() {
					result, err := h.Runtime().Query(h.Context(), "x := data", nil, true, true, true, "full")

					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
					Expect(result.Result[0].Bindings["x"]).To(HaveKey("mycars"))
				})

				It("loads eds plugin and data", func() {
					// We have to retry getting users until it works
					// There seems to be a time delay for registering builtin functions
					Eventually(func() (interface{}, error) {
						result, err := h.Runtime().Query(h.Context(), `x = ds.user({"key":"CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs"})`, nil, true, true, true, "full")
						if err != nil {
							return nil, err
						}

						Expect(len(result.Result)).To(BeNumerically(">", 0))
						Expect(result.Result[0].Bindings).To(HaveKey("x"))
						return result.Result[0].Bindings["x"], nil
					}).Should(HaveKeyWithValue("key", "CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs"))
				})

				It("loads eds plugin and its data", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.user({"id":"CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
				})

				It("loads local bundles", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.user({"id":"CiQ3Y2VlOGU4NS1lM2NmLTRiNmYtODRlYy1mYWM4OTEwN2U5NTcSBWxvY2Fs"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
				})

				It("runs queries with functions that have parameters", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
				})
			})

			Context("with detailed query capabilities", func() {
				It("runs queries and returns metric data when requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, true, false, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					Expect(len(result.Metrics)).To(BeNumerically(">", 0))
				})

				It("runs queries and doesn't return metric data if not requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					Expect(len(result.Metrics)).To(Equal(0))
				})

				It("runs queries and returns an explanation when requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "full")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					Expect(string(json)).ToNot(Equal("[]"))
					Expect(string(json)).To(ContainSubstring("euang@acmecorp.com"))
					// It shouldn't be pretty-printed
					Expect(string(json)).ToNot(ContainSubstring("Enter "))
					Expect(string(json)).ToNot(ContainSubstring("Eval "))
				})

				It("runs queries and returns a pretty explanation when requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, true, false, false, "full")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					Expect(string(json)).ToNot(Equal("[]"))
					Expect(string(json)).To(ContainSubstring("euang@acmecorp.com"))
					// It should be pretty-printed
					Expect(string(json)).To(ContainSubstring("Enter "))
					Expect(string(json)).To(ContainSubstring("Eval "))
				})

				It("runs queries and returns a notes explanation when requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = data.mycars.GET`, nil, false, false, false, "notes")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))

					Expect(string(json)).ToNot(Equal("[]"))
					// It shouldn't be pretty-printed
					Expect(string(json)).To(ContainSubstring("visible block"))
					Expect(string(json)).To(ContainSubstring("enabled block"))
					Expect(string(json)).ToNot(ContainSubstring("Enter "))
				})

				It("runs queries and returns a pretty notes explanation when requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = data.mycars.GET`, nil, true, false, false, "notes")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))

					Expect(string(json)).ToNot(Equal("[]"))
					// It should be pretty-printed
					Expect(string(json)).To(ContainSubstring("visible block"))
					Expect(string(json)).To(ContainSubstring("enabled block"))
					Expect(string(json)).To(ContainSubstring("Enter "))
				})

				It("runs queries and doesn't return an explanation if not requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "off")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					Expect(string(json)).To(BeEmpty())
				})

				It("runs queries and returns metrics if instrumentation requested", func() {
					result, err := h.Runtime().Query(h.Context(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, true, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "CiRkZmRhZGMzOS03MzM1LTQwNGQtYWY2Ni1jNzdjZjEzYTE1ZjgSBWxvY2Fs"))
					// Expect(len(result.Metrics)).To(BeNumerically(">", 0))
				})
			})
		})
	})
})
