package configs

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
)

const (
	configFileName = `config.yaml`
)

var (
	errConfigPathNotExist = errors.New("config path not exist")
	errConfigPathIsDir    = errors.New("config path is dir")
)

type Config struct {
	ServerPort    int    `yaml:"serverPort" default:"8000" desc:"the port milvus-k8s-server is listening"`
	EtcdAddr      string `yaml:"etcd" default:"127.0.0.1:2379" desc:"the etcd endpoint to connect"`
	RootPath      string `yaml:"rootPath" default:"by-dev" desc:"meta root path milvus is using"`
	MetaPath      string `yaml:"metaPath" default:"meta" desc:"meta path prefix"`
	EnableTLS     bool   `yaml:"enableTLS" default:"false" desc:"use TLS"`
	RootCA        string `yaml:"rootCAPem" default:"" desc:"root CA pem file path"`
	ETCDPem       string `yaml:"etcdCert" default:"" desc:"etcd tls cert file path"`
	ETCDKey       string `yaml:"etcdKey" default:"" desc:"etcd tls key file path"`
	TLSMinVersion string `yaml:"min_version" default:"1.2" desc:"TLS min version"`
}

func load(configPath string) (*Config, error) {
	err := checkConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(getConfigPath(configPath))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(bs, c)
	return c, err
}

func getConfigPath(configPath string) string {
	return path.Join(configPath, configFileName)
}

// checkConfigPath exists and is a directory.
func checkConfigPath(configPath string) error {
	info, err := os.Stat(configPath)
	if err != nil {
		// not exist, return specified type to handle
		if os.IsNotExist(err) {
			return errConfigPathNotExist
		}
		return err
	}
	if info.IsDir() {
		fmt.Printf("%s is not a directory\n", configPath)
		return fmt.Errorf("%w(%s)", errConfigPathIsDir, configPath)
	}

	return nil
}

func NewConfig(configPath string) (*Config, error) {
	config, err := load(configPath)

	return config, err
}
