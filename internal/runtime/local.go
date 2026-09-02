package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/loader"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type bundleIndex struct {
	Version   int                  `json:"schemaVersion"` //nolint:tagliatelle
	Manifests []ocispec.Descriptor `json:"manifests"`
}

func (r *Runtime) loadPaths(paths []string) (map[string]*bundle.Bundle, error) {
	if len(paths) == 0 {
		paths = r.Config.LocalBundles.Paths
	}

	if r.Config.LocalBundles.LocalPolicyImage != "" {
		tarballPath, err := r.getPolicyTarballPath(r.Config.LocalBundles.LocalPolicyImage)
		if err != nil {
			r.Logger.Warn().Err(err).Msg("Could not load configured local policy image")
		} else if tarballPath != "" {
			paths = append(paths, tarballPath)
		}
	}

	result := make(map[string]*bundle.Bundle, len(paths))

	skipVerify := r.Config.LocalBundles.SkipVerification
	verificationConfig := r.Config.LocalBundles.VerificationConfig

	var err error

	for _, path := range paths {
		r.Logger.Info().Str("path", path).Msg("Loading local bundle")

		result[path], err = loader.NewFileLoader().
			WithBundleVerificationConfig(verificationConfig).
			WithSkipBundleVerification(skipVerify).
			AsBundle(path)
		if err != nil {
			return nil, errors.Wrapf(err, "load bundle from local path '%s'", path)
		}
	}

	return result, nil
}

func (r *Runtime) getPolicyTarballPath(policyImageRef string) (string, error) {
	storeRoot, err := r.fileStoreRoot()
	if err != nil {
		return "", err
	}

	time.Sleep(1 * time.Second) // wait until index.json is updated

	localIndex, err := r.loadBundleIndex(storeRoot)
	if err != nil {
		return "", err
	}

	// load manifest for policyImageRef
	manifest, found := localIndex.findManifest(policyImageRef)

	if found && manifest.MediaType == ocispec.MediaTypeImageLayerGzip {
		return filepath.Join(storeRoot, "policies-root", "blobs", "sha256", manifest.Digest.Hex()), nil
	}

	if !found || manifest.Digest == "" {
		return "", errors.Errorf(
			"could not find policy image %s with a supported media type ('%s' or '%s')",
			policyImageRef, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageLayerGzip,
		)
	}

	manifestFile := filepath.Join(storeRoot, "policies-root", "blobs", "sha256", manifest.Digest.Hex())

	manifestBytes, err := os.ReadFile(manifestFile)
	if err != nil {
		return "", err
	}

	var searchedManifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &searchedManifest); err != nil {
		return "", err
	}

	if len(searchedManifest.Layers) != 1 {
		return "", errors.New("unknown image type - incorrect number of layers")
	}

	tarballPath := filepath.Join(
		r.Config.LocalBundles.FileStoreRoot,
		"policies-root",
		"blobs",
		"sha256",
		searchedManifest.Layers[0].Digest.Hex(),
	)

	return tarballPath, nil
}

func (i *bundleIndex) findManifest(policyImageRef string) (*ocispec.Descriptor, bool) {
	for _, manifest := range i.Manifests {
		refName := manifest.Annotations[ocispec.AnnotationRefName]
		if strings.Contains(refName, policyImageRef) && (manifest.MediaType == ocispec.MediaTypeImageLayerGzip ||
			manifest.MediaType == ocispec.MediaTypeImageManifest) {
			return &manifest, true
		}
	}

	return nil, false
}

func (r *Runtime) loadBundleIndex(storeRoot string) (*bundleIndex, error) {
	indexPath := filepath.Join(storeRoot, "policies-root", "index.json")

	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	if len(indexBytes) == 0 {
		return nil, errors.Errorf("empty index.json file")
	}

	// load index.json from root oci path
	var index bundleIndex
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

func (r *Runtime) fileStoreRoot() (string, error) {
	if r.Config.LocalBundles.FileStoreRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "failed to determine user home directory")
		}

		r.Config.LocalBundles.FileStoreRoot = filepath.Join(home, ".policy")
	}

	return r.Config.LocalBundles.FileStoreRoot, nil
}
