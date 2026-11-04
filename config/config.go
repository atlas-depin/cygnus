package config

import (
	"errors"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func (c BaseConfig) Validate() error {
	if c.DataDirectory == "" {
		return errors.New("invalid data directory")
	}

	switch strings.ToLower(strings.TrimSpace(c.APICfg.UploadBufferMode)) {
	case "", UploadBufferAuto, UploadBufferMemory, UploadBufferDisk:
	default:
		return fmt.Errorf("invalid api_config.upload_buffer_mode %q (want %q, %q or %q)",
			c.APICfg.UploadBufferMode, UploadBufferAuto, UploadBufferMemory, UploadBufferDisk)
	}

	return nil
}

// ReadConfig parses data and returns Config.
// Error during parsing or an invalid configuration in the Config will return an error.
func ReadConfig(data []byte) (*Config, error) {
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, config.Validate()
}

func (c BaseConfig) Export() ([]byte, error) {
	sb := strings.Builder{}
	d, err := yaml.Marshal(&c)
	if err != nil {
		return nil, err
	}

	sb.Write(d)

	return []byte(sb.String()), nil
}
