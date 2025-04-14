package directory

import (
	"io"
	"text/template"
	"time"

	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type BoltDBStore struct {
	directory.Config `json:"config,squash"` //nolint:staticcheck  //squash accepted by mapstructure
}

const BoltDBDefaultRequestTimeout = time.Second * 5

const BoltDBStorePlugin string = "boltdb"

func (c *BoltDBStore) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("db_path", "${TOPAZ_DB_DIR}/directory.db")
	v.SetDefault("request_timeout", BoltDBDefaultRequestTimeout.String())
}

func (c *BoltDBStore) Validate() (bool, error) {
	return true, nil
}

func (c *BoltDBStore) Generate(w io.Writer) error {
	tmpl, err := template.New("STORE").Parse(boltDBStoreTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func (c BoltDBStore) Map() map[string]interface{} {
	var m map[string]interface{}
	if err := mapstructure.Decode(c, &m); err != nil {
		return nil
	}

	return m
}

func BoltDBStoreFromMap(m map[string]interface{}) *BoltDBStore {
	var cfg BoltDBStore
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

func BoltDBStoreMap(cfg *BoltDBStore) map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(cfg, &result); err != nil {
		return nil
	}

	return result
}

func (c *BoltDBStore) ToMap() map[string]interface{} {
	var result map[string]interface{}
	if err := mapstructure.Decode(c, &result); err != nil {
		return nil
	}

	return result
}

const boltDBStoreTemplate = `      db_path: '{{ .DBPath }}'
`
