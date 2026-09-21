package access

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aserto-dev/topaz/topaz/cmd/configure"
)

type WellKnownCmd struct{}

func (cmd *WellKnownCmd) Run(ctx context.Context) error {
	info := configure.InfoConfigCmd{Var: "", Raw: false}.GetInfo()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
		return fmt.Errorf("unexpected status: %s\n", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	bodyReader := bytes.NewReader(bodyBytes)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	return enc.Encode(bodyReader)
}
