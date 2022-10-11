package engine_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/testing"
)

var _ = Describe("Engine Runtime", func() {
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

		Context("when using the engine directly", func() {
			Context("with basic functionality", func() {
				It("successfully returns runtime info", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), "x := opa.runtime()", nil, true, true, true, "full")

					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())
					Expect(result.Result[0].Bindings).To(HaveKey("x"))

					Expect(result.Result[0].Bindings["x"]).To(HaveKey("env"))
					x := result.Result[0].Bindings["x"]
					Expect(x).ToNot(BeNil())
				})

				It("successfully returns bundle data", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), "x := data", nil, true, true, true, "full")

					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
					Expect(result.Result[0].Bindings["x"]).To(HaveKey("mycars"))
				})

				It("loads eds plugin and data", func() {
					// We have to retry getting users until it works
					// There seems to be a time delay for registering builtin functions
					Eventually(func() (interface{}, error) {
						result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.user({"id":"7cee8e85-e3cf-4b6f-84ec-fac89107e957"})`, nil, true, true, true, "full")
						if err != nil {
							return nil, err
						}

						Expect(len(result.Result)).To(BeNumerically(">", 0))
						Expect(result.Result[0].Bindings).To(HaveKey("x"))
						return result.Result[0].Bindings["x"], nil
					}).Should(HaveKeyWithValue("id", "7cee8e85-e3cf-4b6f-84ec-fac89107e957"))
				})

				It("loads eds plugin and its data", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.user({"id":"7cee8e85-e3cf-4b6f-84ec-fac89107e957"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
				})

				It("loads local bundles", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.user({"id":"7cee8e85-e3cf-4b6f-84ec-fac89107e957"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKey("x"))
				})

				It("runs queries with functions that have parameters", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, true, true, true, "full")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
				})
			})

			Context("with detailed query capabilities", func() {
				It("runs queries and returns metric data when requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, true, false, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(len(result.Metrics)).To(BeNumerically(">", 0))
				})

				It("runs queries and doesn't return metric data if not requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(len(result.Metrics)).To(Equal(0))
				})

				It("runs queries and returns an explanation when requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "full")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(string(json)).ToNot(Equal("[]"))
					Expect(string(json)).To(ContainSubstring("euang@acmecorp.com"))
					// It shouldn't be pretty-printed
					Expect(string(json)).ToNot(ContainSubstring("Enter "))
					Expect(string(json)).ToNot(ContainSubstring("Eval "))
				})

				It("runs queries and returns a pretty explanation when requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, true, false, false, "full")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(string(json)).ToNot(Equal("[]"))
					Expect(string(json)).To(ContainSubstring("euang@acmecorp.com"))
					// It should be pretty-printed
					Expect(string(json)).To(ContainSubstring("Enter "))
					Expect(string(json)).To(ContainSubstring("Eval "))
				})

				It("runs queries and returns a notes explanation when requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = data.mycars.GET`, nil, false, false, false, "notes")
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
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = data.mycars.GET`, nil, true, false, false, "notes")
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
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, false, "off")
					Expect(err).ToNot(HaveOccurred())

					json, err := result.Explanation.MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(string(json)).To(BeEmpty())
				})

				It("runs queries and returns metrics if instrumentation requested", func() {
					result, err := h.Runtime().Query(h.ContextWithTenant(), `x = ds.identity({"key":"euang@acmecorp.com"})`, nil, false, false, true, "off")
					Expect(err).ToNot(HaveOccurred())

					Expect(len(result.Result)).To(BeNumerically(">", 0))
					Expect(result.Result[0].Bindings).To(HaveKeyWithValue("x", "dfdadc39-7335-404d-af66-c77cf13a15f8"))
					Expect(len(result.Metrics)).To(BeNumerically(">", 0))
				})
			})
		})
	})
})
