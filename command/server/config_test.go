package server

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigDefault(t *testing.T) {
	// the default config should work out of the box
	config := DefaultConfig()
	assert.NoError(t, config.loadChain())

	_, err := config.buildNode()
	assert.NoError(t, err)

	_, err = config.buildEth()
	assert.NoError(t, err)
}

func TestConfigMerge(t *testing.T) {
	c0 := &Config{
		Chain: "0",
		Debug: true,
		Whitelist: map[string]string{
			"a": "b",
		},
		TxPool: &TxPoolConfig{
			LifeTime: 5 * time.Second,
		},
		P2P: &P2PConfig{
			Discovery: &P2PDiscovery{
				StaticNodes: []string{
					"a",
				},
			},
		},
	}
	c1 := &Config{
		Chain: "1",
		Whitelist: map[string]string{
			"b": "c",
		},
		P2P: &P2PConfig{
			MaxPeers: 10,
			Discovery: &P2PDiscovery{
				StaticNodes: []string{
					"b",
				},
			},
		},
	}
	expected := &Config{
		Chain: "1",
		Debug: true,
		Whitelist: map[string]string{
			"a": "b",
			"b": "c",
		},
		TxPool: &TxPoolConfig{
			LifeTime: 5 * time.Second,
		},
		P2P: &P2PConfig{
			MaxPeers: 10,
			Discovery: &P2PDiscovery{
				StaticNodes: []string{
					"a",
					"b",
				},
			},
		},
	}
	assert.NoError(t, c0.Merge(c1))
	assert.Equal(t, c0, expected)
}

func TestConfigHcl(t *testing.T) {
	readConfig := func(data string, format string) *Config {
		tmpDir, err := ioutil.TempDir("/tmp", "test-config")
		assert.NoError(t, err)

		filename := filepath.Join(tmpDir, "config."+format)
		assert.NoError(t, ioutil.WriteFile(filename, []byte(data), 0755))

		config, err := readConfigFile(filename)
		assert.NoError(t, err)
		return config
	}

	cfg := `{
		"datadir": "datadir",
		"p2p": {
			"max_peers": 30
		}
	}`
	config := readConfig(cfg, "json")
	assert.Equal(t, config, &Config{
		DataDir: "datadir",
		P2P: &P2PConfig{
			MaxPeers: 30,
		},
	})
}

var dummyEnodeAddr = "enode://0cb82b395094ee4a2915e9714894627de9ed8498fb881cec6db7c65e8b9a5bd7f2f25cc84e71e89d0947e51c76e85d0847de848c7782b13c0255247a6758178c@44.232.55.71:30303"

func TestConfigBootnodesDefault(t *testing.T) {
	t.Run("EmptyBootnodes", func(t *testing.T) {
		// if no bootnodes are specific, we use the ones from the genesis chain
		config := DefaultConfig()
		assert.NoError(t, config.loadChain())

		cfg, err := config.buildNode()
		assert.NoError(t, err)
		assert.NotEmpty(t, cfg.P2P.BootstrapNodes)
	})
	t.Run("NotEmptyBootnodes", func(t *testing.T) {
		// if bootnodes specific, DO NOT load the genesis bootnodes
		config := DefaultConfig()
		config.P2P.Discovery.Bootnodes = []string{dummyEnodeAddr}

		cfg, err := config.buildNode()
		assert.NoError(t, err)
		assert.Len(t, cfg.P2P.BootstrapNodes, 1)
	})
}
