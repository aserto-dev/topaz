package assets_test

import (
	"bytes"
	_ "embed"
)

//go:embed config/config.yaml
var config []byte

func ConfigReader() *bytes.Reader {
	return bytes.NewReader(config)
}

//go:embed config/config-no-tls.yaml
var configNoTLS []byte

func ConfigNoTLSReader() *bytes.Reader {
	return bytes.NewReader(configNoTLS)
}

//go:embed config/peoplefinder.yaml
var configOnline []byte

func PeoplefinderConfigReader() *bytes.Reader {
	return bytes.NewReader(configOnline)
}

//go:embed gdrive/manifest.yaml
var manifest []byte

func ManifestReader() *bytes.Reader {
	return bytes.NewReader(manifest)
}

//go:embed db/acmecorp.db
var acmecorp []byte

func AcmecorpReader() *bytes.Reader {
	return bytes.NewReader(acmecorp)
}
