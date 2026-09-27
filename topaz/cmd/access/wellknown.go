package access

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/configure"
	"github.com/aserto-dev/topaz/topazd/app/handlers"
	"github.com/pkg/errors"
)

type WellKnownCmd struct{}

func (cmd *WellKnownCmd) Run(ctx context.Context) error {
	info := configure.InfoConfigCmd{Var: "", Raw: false}.GetInfo()
	if info == nil {
		return cc.ErrNoConfig
	}

	ctx, cancel := context.WithTimeout(ctx, cc.Timeout())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.Access.WellKnownConfigURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if req.URL, err = wellKnownURL(req.URL, info.Directory.Plaintext); err != nil {
		return fmt.Errorf("building request URL: %w", err)
	}

	resp, err := wellKnownClient(info.Directory.Insecure).Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status: %s\n", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	var wellknown handlers.WellKnownConfig
	if err := json.Unmarshal(bodyBytes, &wellknown); err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	return enc.Encode(wellknown)
}

func wellKnownURL(reqURL *url.URL, plaintext bool) (*url.URL, error) {
	u, err := url.Parse(reqURL.String())
	if err != nil {
		return nil, err
	}

	if plaintext {
		u.Scheme = "http"
	}

	return u, nil
}

func wellKnownClient(insecure bool) *http.Client {
	if !insecure {
		return http.DefaultClient
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return &http.Client{Transport: transport}
}
