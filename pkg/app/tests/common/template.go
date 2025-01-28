package common_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/templates"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func InstallTemplate(dsConfig *dsc.Config, azConfig *azc.Config, tmpl string) func(*testing.T) {
	return func(t *testing.T) {
		t.Logf("addr: %s tmpl: %s", dsConfig.Host, tmpl)

		t.Setenv("TOPAZ_NO_COLOR", "true")
		c, err := cc.NewCommonContext(context.Background(), true, filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile))
		require.NoError(t, err)

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
		t.Run(tmpl.Name+"-DeleteManifest", DeleteManifest(c, dsConfig))
		t.Run(tmpl.Name+"-SetManifest", SetManifest(c, dsConfig, manifestFile))

		if len(tmpl.Assets.IdentityData) > 0 {
			idpDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.IdentityData[0]))
			t.Logf("idp_data: %s", idpDataDir)
			t.Run(tmpl.Name+"-ImportIdentityData", ImportData(c, dsConfig, idpDataDir))
		}

		if len(tmpl.Assets.DomainData) > 0 {
			domainDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.DomainData[0]))
			t.Logf("domain_data: %s", domainDataDir)
			t.Run(tmpl.Name+"-ImportDomainData", ImportData(c, dsConfig, domainDataDir))
		}

		if len(tmpl.Assets.Assertions) > 0 {
			assertionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[0])
			t.Logf("assertionsFile: %s", assertionsFile)
			t.Run(tmpl.Name+"-ExecDirectoryTest", ExecDirectoryTests(c, dsConfig, []string{assertionsFile}))
		}

		if len(tmpl.Assets.Assertions) > 1 {
			decisionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[1])
			t.Logf("decisionsFile: %s", decisionsFile)
			t.Run(tmpl.Name+"-ExecAuthorizerTest", ExecAuthorizerTests(c, azConfig, []string{decisionsFile}))
		}
	}
}

func DeleteManifest(c *cc.CommonCtx, cfg *dsc.Config) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.DeleteManifestCmd{Config: *cfg, Force: true}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func SetManifest(c *cc.CommonCtx, cfg *dsc.Config, path string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.SetManifestCmd{Config: *cfg, Path: path}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ImportData(c *cc.CommonCtx, cfg *dsc.Config, dir string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.ImportCmd{Config: *cfg, Directory: dir}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ExecDirectoryTests(c *cc.CommonCtx, cfg *dsc.Config, files []string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.TestExecCmd{Config: *cfg, TestExecCmd: common.TestExecCmd{Files: files, Summary: true, Desc: "on-error"}}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ExecAuthorizerTests(c *cc.CommonCtx, cfg *azc.Config, files []string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := authorizer.TestExecCmd{Config: *cfg, TestExecCmd: common.TestExecCmd{Files: files, Summary: true, Desc: "on-error"}}
		if err := cmd.Run(c); err != nil {
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
