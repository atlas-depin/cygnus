package config

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Seed struct {
	SeedPhrase     string `json:"seed_phrase"`
	DerivationPath string `json:"derivation_path"`
}

type ChainConfig struct {
	ChainId        string  `yaml:"chain_id" mapstructure:"chain_id"`
	KeyringBackend string  `yaml:"keyring_backend" mapstructure:"keyring_backend"`
	RPCAddr        string  `yaml:"rpc_addr" mapstructure:"rpc_addr"`
	GRPCAddr       string  `yaml:"grpc_addr" mapstructure:"grpc_addr"`
	GasPrice       string  `yaml:"gas_price" mapstructure:"gas_price"`
	GasAdjustment  float64 `yaml:"gas_adjustment" mapstructure:"gas_adjustment"`
}

// Upload buffer modes for APIConfig.UploadBufferMode.
const (
	// UploadBufferAuto streams to temp files only when max_upload_size is 0
	// (i.e. unlimited). Bounded uploads are buffered in RAM.
	UploadBufferAuto = "auto"
	// UploadBufferMemory always buffers the entire request body in RAM.
	// Fastest, but a single upload cannot exceed available memory.
	UploadBufferMemory = "memory"
	// UploadBufferDisk always streams, spilling multipart parts larger than
	// 16MB to temp files. Lower memory, at the cost of extra disk I/O.
	UploadBufferDisk = "disk"
)

type APIConfig struct {
	Port             int64  `yaml:"port" mapstructure:"port"`
	MaxUploadSize    int64  `yaml:"max_upload_size" mapstructure:"max_upload_size"`
	PrioritizeProofs bool   `yaml:"prioritize_proofs" mapstructure:"prioritize_proofs"`
	FsyncUploads     bool   `yaml:"fsync_uploads" mapstructure:"fsync_uploads"`
	UploadBufferMode string `yaml:"upload_buffer_mode" mapstructure:"upload_buffer_mode"`
	TimeoutSeconds   int    `yaml:"timeout_seconds" mapstructure:"timeout_seconds"`
}

// Timeout resolves the configured timeout.
func (c APIConfig) Timeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return 0
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// StreamRequestBody reports whether the HTTP server should stream request
// bodies to temp files instead of buffering them fully in RAM. An unset or
// unrecognized UploadBufferMode falls back to "auto".
func (c APIConfig) StreamRequestBody() bool {
	switch strings.ToLower(strings.TrimSpace(c.UploadBufferMode)) {
	case UploadBufferMemory:
		return false
	case UploadBufferDisk:
		return true
	default: // "", UploadBufferAuto
		return c.MaxUploadSize <= 0
	}
}

type Config struct {
	BaseConfig    `mapstructure:",squash"`
	HomeDirectory string
}

type BaseConfig struct {
	DataDirectory string      `yaml:"data_directory" mapstructure:"data_directory"`
	ChainCfg      ChainConfig `yaml:"chain_config" mapstructure:"chain_config"`
	APICfg        APIConfig   `yaml:"api_config" mapstructure:"api_config"`

	ProviderName     string           `yaml:"provider_name" mapstructure:"provider_name"`
	Ip               string           `yaml:"hostname" mapstructure:"hostname"`
	TotalSpace       int64            `yaml:"total_bytes_offered" mapstructure:"total_bytes_offered"`
	CacheMerkleTrees bool             `yaml:"cache_merkle_trees" mapstructure:"cache_merkle_trees"`
	StraySweep       StraySweepConfig `yaml:"stray_sweep" mapstructure:"stray_sweep"`
}

// DefaultAPIConfig returns the default APIConfig with preset ports & advanced options focused on efficiency.
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		Port:             3333,
		MaxUploadSize:    0,
		FsyncUploads:     true,
		PrioritizeProofs: true,
		UploadBufferMode: UploadBufferAuto,
		TimeoutSeconds:   1800,
	}
}

// DefaultChainConfig returns the default ChainConfig with preset RPC & GRPC endpoints.
func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		ChainId:        "itineraria-1",
		RPCAddr:        "https://rpc.atlasdepin.com",
		GRPCAddr:       "grpc.atlasdepin.com:443",
		GasPrice:       "0.03uatl",
		GasAdjustment:  2.0,
		KeyringBackend: "test",
	}
}

func DefaultConfig(home string) *BaseConfig {
	return &BaseConfig{
		ProviderName:     "My First Provider",
		Ip:               "localhost",
		TotalSpace:       10 * 1000 * 1000 * 1000,
		DataDirectory:    home + "/data",
		CacheMerkleTrees: true,

		ChainCfg:   DefaultChainConfig(),
		APICfg:     DefaultAPIConfig(),
		StraySweep: DefaultStraySweepConfig(),
	}
}

func (c Config) MarshalZerologObject(e *zerolog.Event) {
	e.Str("ChainRPCAddr", c.ChainCfg.RPCAddr).
		Str("ChainGRPCAddr", c.ChainCfg.GRPCAddr).
		Str("ChainGasPrice", c.ChainCfg.GasPrice).
		Float64("ChainGasAdjustment", c.ChainCfg.GasAdjustment).
		Str("IP", c.Ip).
		Int64("TotalSpace", c.TotalSpace).
		Str("DataDirectory", c.DataDirectory).
		Int64("API_Port", c.APICfg.Port).
		Bool("API_FsyncUploads", c.APICfg.FsyncUploads).
		Bool("API_PrioritizeProofs", c.APICfg.PrioritizeProofs).
		Int64("API_MaxUploadSize", c.APICfg.MaxUploadSize).
		Bool("API_StreamRequestBody", c.APICfg.StreamRequestBody()).
		Dur("API_Timeout", c.APICfg.Timeout()).
		Bool("CacheMerkleTrees", c.CacheMerkleTrees).
		Bool("StraySweep_Enabled", c.StraySweep.Enabled)
}

// StraySweepConfig controls the background loop that discovers and claims STRAY files.
type StraySweepConfig struct {
	Enabled             bool `yaml:"enabled" mapstructure:"enabled"`
	IntervalSeconds     int  `yaml:"interval_seconds" mapstructure:"interval_seconds"`
	MaxClaimsPerSweep   int  `yaml:"max_claims_per_sweep" mapstructure:"max_claims_per_sweep"`
	MaxConcurrentClaims int  `yaml:"max_concurrent_claims" mapstructure:"max_concurrent_claims"`
}

// DefaultStraySweepConfig returns sensible defaults for the stray-file sweeper.
func DefaultStraySweepConfig() StraySweepConfig {
	return StraySweepConfig{
		Enabled:             true,
		IntervalSeconds:     60,
		MaxClaimsPerSweep:   25,
		MaxConcurrentClaims: 5,
	}
}
