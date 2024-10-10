package templates

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/topaz/pkg/cc/config"
)

type TemplateCmd struct {
	List    ListTemplatesCmd   `cmd:"" help:"list template"`
	Install InstallTemplateCmd `cmd:"" help:"install template"`
	Verify  VerifyTemplateCmd  `cmd:"" help:"verify template content links" hidden:""`
}

const (
	colName          string = "name"
	colDescription   string = "description"
	colDocumentation string = "documentation"
)

type tmplCatalog map[string]*tmplRef

type tmplRef struct {
	Title            string `json:"title,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Description      string `json:"description,omitempty"`
	URL              string `json:"topaz_url,omitempty"`
	DocumentationURL string `json:"docs_url,omitempty"`

	absURL *url.URL
}

type template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Assets      struct {
		Manifest string `json:"manifest"`
		Policy   struct {
			Name     string `json:"name"`
			Resource string `json:"resource"`
			Local    bool   `json:"local"`
		} `json:"policy,omitempty"`
		IdentityData []string `json:"idp_data,omitempty"`
		DomainData   []string `json:"domain_data,omitempty"`
		Assertions   []string `json:"assertions,omitempty"`
	} `json:"assets"`

	base *url.URL
}

func (t *template) AbsURL(relative string) string {
	abs := *t.base
	abs.Path = path.Join(abs.Path, relative)
	return abs.String()
}

func download(src, dir string) (string, error) {
	buf, err := getBytes(src)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	f, err := os.Create(filepath.Join(dir, filepath.Base(src)))
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(buf); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func getBytes(fileURL string) ([]byte, error) {
	if exists, _ := config.FileExists(fileURL); exists {
		return os.ReadFile(fileURL)
	}

	resp, err := http.Get(fileURL) //nolint used to download the template files
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getTemplate(name, templatesURL string) (*template, error) {
	if exists, _ := config.FileExists(name); exists {
		return getTemplateFromFile(name)
	}

	tmplURL, err := url.Parse(templatesURL)
	if err != nil {
		return nil, err
	}

	ref, err := getTemplateRef(name, tmplURL)
	if err != nil {
		return nil, err
	}

	tmpl, err := ref.getTemplate()
	if err != nil {
		return nil, err
	}
	tmpl.Name = name

	return tmpl, nil
}

func getCatalog(templatesURL string) (tmplCatalog, error) {
	buf, err := getBytes(templatesURL)
	if err != nil {
		return nil, err
	}

	var ctlg tmplCatalog
	if err := json.Unmarshal(buf, &ctlg); err != nil {
		return nil, err
	}

	return ctlg, nil
}

func getTemplateRef(name string, templatesURL *url.URL) (*tmplRef, error) {
	list, err := getCatalog(templatesURL.String())
	if err != nil {
		return nil, err
	}

	if ref, ok := list[name]; ok {
		parsed, err := url.Parse(ref.URL)
		if err != nil {
			return nil, err
		}
		if parsed.Scheme == "" {
			absURL := *templatesURL
			absURL.Path = path.Join(path.Dir(absURL.Path), parsed.Path)
			ref.absURL = &absURL
		}

		return ref, nil
	}

	return nil, derr.ErrNotFound.Msgf("template %s", name)
}

func getTemplateFromFile(file string) (*template, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tmpl := template{base: &url.URL{Path: path.Dir(file)}}
	if err := json.Unmarshal(b, &tmpl); err != nil {
		return nil, err
	}

	return &tmpl, nil
}

func (i *tmplRef) getTemplate() (*template, error) {
	buf, err := getBytes(i.absURL.String())
	if err != nil {
		return nil, err
	}

	base := *i.absURL
	base.Path = path.Dir(base.Path)
	tmpl := template{base: &base}
	if err := json.Unmarshal(buf, &tmpl); err != nil {
		return nil, err
	}

	// TODO: remove after updating all templates to URLs relative to the template definition.
	tmpl.Assets.Manifest = strings.TrimPrefix(tmpl.Assets.Manifest, base.Path)
	for i, v := range tmpl.Assets.IdentityData {
		tmpl.Assets.IdentityData[i] = strings.TrimPrefix(v, base.Path)
	}
	for i, v := range tmpl.Assets.DomainData {
		tmpl.Assets.DomainData[i] = strings.TrimPrefix(v, base.Path)
	}

	return &tmpl, nil
}

func GetTemplateFromFile(file string) (*template, error) {
	return getTemplateFromFile(file)
}
