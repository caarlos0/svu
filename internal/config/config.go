package config

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AdditionalFeatureTypes []string `yaml:"additionalFeatureTypes,omitempty"`
	AdditionalFixTypes     []string `yaml:"additionalFixTypes,omitempty"`
}

// Load config file.
func Load(file string) (config Config, err error) {
	f, err := os.Open(file) // #nosec
	if err != nil {
		return
	}
	defer f.Close()
	log.WithField("file", file).Info("loading config file")
	return LoadReader(f)
}

// LoadReader config via io.Reader.
func LoadReader(fd io.Reader) (config Config, err error) {
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return config, err
	}
	err = yaml.UnmarshalStrict(data, &config)
	log.WithField("config", config).Debug("loaded config file")
	return config, err
}
