package common_test

import (
	"context"
	"path/filepath"
	"testing"

	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
	"github.com/aserto-dev/topaz/topaz/cmd/directory"
	"github.com/aserto-dev/topaz/topaz/cmd/templates"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func InstallTemplate(ctx context.Context, dsConfig *dsc.Config, azConfig *azc.Config, tmpl string) func(*testing.T) {
	return func(t *testing.T) {
		t.Logf("addr: %s tmpl: %s", dsConfig.Host, tmpl)

		t.Setenv("TOPAZ_NO_COLOR", "true")

		absPath, err := filepath.Abs(tmpl)
		require.NoError(t, err)

		tmpl, err := templates.GetTemplateFromFile(absPath)
		require.NoError(t, err)

		t.Logf("name %s", tmpl.Name)
		t.Logf("template: %s", absPath)

		dirPath := filepath.Dir(absPath)
		t.Logf("dir %s", dirPath)

		manifestFile := filepath.Join(dirPath, tmpl.Assets.Manifest)
		t.Logf("manifestFile: %s", manifestFile)
		t.Run(tmpl.Name+"-DeleteManifest", DeleteManifest(ctx, dsConfig))
		t.Run(tmpl.Name+"-SetManifest", SetManifest(ctx, dsConfig, manifestFile))

		if len(tmpl.Assets.IdentityData) > 0 {
			idpDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.IdentityData[0]))
			t.Logf("idp_data: %s", idpDataDir)
			t.Run(tmpl.Name+"-ImportIdentityData", ImportData(ctx, dsConfig, idpDataDir))
		}

		if len(tmpl.Assets.DomainData) > 0 {
			domainDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.DomainData[0]))
			t.Logf("domain_data: %s", domainDataDir)
			t.Run(tmpl.Name+"-ImportDomainData", ImportData(ctx, dsConfig, domainDataDir))
		}

		if len(tmpl.Assets.Assertions) > 0 {
			for _, assertions := range tmpl.Assets.Assertions {
				assertionsFile := filepath.Join(dirPath, assertions)
				t.Logf("assertionsFile: %s", assertionsFile)
				t.Run(assertions, ExecTests(ctx, dsConfig, azConfig, []string{assertionsFile}))
			}
		}
	}
}

func DeleteManifest(ctx context.Context, cfg *dsc.Config) func(*testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		t.Cleanup(cancel)

		cmd := directory.DeleteManifestCmd{Config: *cfg, Force: true}
		if err := cmd.Run(ctx); err != nil {
			assert.NoError(t, err)
		}
	}
}

func SetManifest(ctx context.Context, cfg *dsc.Config, path string) func(*testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		t.Cleanup(cancel)

		cmd := directory.SetManifestCmd{Config: *cfg, Path: path}
		if err := cmd.Run(ctx); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ImportData(ctx context.Context, cfg *dsc.Config, dir string) func(*testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		t.Cleanup(cancel)

		cmd := directory.ImportCmd{Config: *cfg, Directory: dir}
		if err := cmd.Run(ctx); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ExecTests(ctx context.Context, dsConfig *dsc.Config, azConfig *azc.Config, files []string) func(*testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, dsConfig.Timeout)
		t.Cleanup(cancel)

		cmd, err := common.NewTestRunner(ctx, &common.TestExecCmd{Files: files, Summary: true, Desc: "on-error"}, azConfig, dsConfig)
		require.NoError(t, err)

		if err := cmd.Run(ctx); err != nil {
			assert.NoError(t, err)
		}
	}
}

func SetContext(k, v string) *structpb.Struct {
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			k: structpb.NewStringValue(v),
		},
	}
}
