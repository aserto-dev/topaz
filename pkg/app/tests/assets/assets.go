package assets_test

import (
	"bytes"
	_ "embed"
	"io"
)

//go:embed config/config.yaml
var config []byte

func ConfigReader() io.Reader {
	return bytes.NewReader(config)
}

//go:embed config/config-no-tls.yaml
var configNoTLS []byte

func ConfigNoTLSReader() io.Reader {
	return bytes.NewReader(configNoTLS)
}

//go:embed config/peoplefinder.yaml
var configOnline []byte

func PeoplefinderConfigReader() io.Reader {
	return bytes.NewReader(configOnline)
}

//go:embed gdrive/manifest.yaml
var manifest []byte

func ManifestReader() io.Reader {
	return bytes.NewReader(manifest)
}

//go:embed db/acmecorp.db
var acmecorp []byte

func AcmecorpReader() io.Reader {
	return bytes.NewReader(acmecorp)
}
