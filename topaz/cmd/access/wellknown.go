package access

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/configure"
	"github.com/aserto-dev/topaz/topazd/app/handlers"
	"github.com/olekukonko/errors"
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

	resp, err := http.DefaultClient.Do(req)
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
