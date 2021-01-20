package config

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	f, err := ioutil.TempFile(t.TempDir(), "config")
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	_, err = Load(filepath.Join(f.Name()))
	require.NoError(t, err)
}

func TestFileNotFound(t *testing.T) {
	_, err := Load("/nope/no-way.yml")
	require.Error(t, err)
}

func TestInvalidFields(t *testing.T) {
	_, err := Load("testdata/invalid_config.yml")
	require.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: field foo not found in type config.Config")
}

func TestInvalidYaml(t *testing.T) {
	_, err := Load("testdata/invalid.yml")
	require.EqualError(t, err, "yaml: line 1: did not find expected node content")
}

func TestValidConfig(t *testing.T) {
	config, err := Load("testdata/valid.yml")
	require.NoError(t, err)
	require.Equal(t, config.AdditionalFeatureTypes, []string{"my-feature"})
	require.Equal(t, config.AdditionalFixTypes, []string{"docs", "refactor"})
}
