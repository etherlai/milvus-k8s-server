package kv

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cockroachdb/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"milvus-k8s-server/pkg/configs"
	"os"
	"time"
)

// MetaKV contains base operations of kv. Include save, load and remove etc.
type MetaKV interface {
	Load(ctx context.Context, key string) (string, error)
	LoadWithPrefix(ctx context.Context, key string) ([]string, []string, error)
	Close()
}

// implementation assertion
var _ MetaKV = (*etcdKV)(nil)

// etcdKV implements TxnKV interface, it supports to process multiple kvs in a transaction.
type etcdKV struct {
	client   *clientv3.Client
	rootPath string
	metaPath string
}

// NewEtcdKV creates a new etcd kv.
func NewEtcdKV(client *clientv3.Client, rootPath string, metaPath string) *etcdKV {
	kv := &etcdKV{
		client:   client,
		rootPath: rootPath,
		metaPath: metaPath,
	}
	return kv
}

// Load returns value of the key.
func (kv *etcdKV) Load(ctx context.Context, key string) (string, error) {
	key = joinPath(kv.rootPath, kv.metaPath, key)
	resp, err := kv.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if resp.Count <= 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return string(resp.Kvs[0].Value), nil
}

// LoadWithPrefix returns all the keys and values with the given key prefix.
func (kv *etcdKV) LoadWithPrefix(ctx context.Context, key string) ([]string, []string, error) {
	key = joinPath(kv.rootPath, kv.metaPath, key)
	resp, err := kv.client.Get(ctx, key, clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil, nil, err
	}
	keys := make([]string, 0, resp.Count)
	values := make([]string, 0, resp.Count)
	for _, kv := range resp.Kvs {
		keys = append(keys, string(kv.Key))
		values = append(values, string(kv.Value))
	}
	return keys, values, nil
}

// Close closes the connection to etcd.
func (kv *etcdKV) Close() {
	kv.client.Close()
}

func ConnectEtcd(c *configs.Config) (*etcdKV, error) {
	tls, err := getTLSConfig(c)
	if err != nil {
		return nil, err
	}
	cfg := clientv3.Config{
		Endpoints:   []string{c.EtcdAddr},
		DialTimeout: time.Second * 10,

		TLS: tls,
		// disable grpc logging
		Logger: zap.NewNop(),
	}
	etcdCli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	cli := NewEtcdKV(etcdCli, c.RootPath, c.MetaPath)
	return cli, nil
}

func getTLSConfig(c *configs.Config) (*tls.Config, error) {
	if !c.EnableTLS {
		return nil, nil
	}

	var tlsMinVersion uint16
	switch c.TLSMinVersion {
	case "1.0":
		tlsMinVersion = tls.VersionTLS10
	case "1.1":
		tlsMinVersion = tls.VersionTLS11
	case "1.2":
		tlsMinVersion = tls.VersionTLS12
	case "1.3":
		tlsMinVersion = tls.VersionTLS13
	default:
		return nil, errors.New("invalid min tls version, only 1.0, 1.1, 1.2 and 1.3 is supported")
	}

	rootCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if c.RootCA != "" {
		bs, err := os.ReadFile(c.RootCA)
		if err != nil {
			return nil, err
		}

		ok := rootCertPool.AppendCertsFromPEM(bs)
		if !ok {
			return nil, errors.New("Root CA PEM cannot be parsed")
		}
	}

	cert, err := tls.LoadX509KeyPair(c.ETCDPem, c.ETCDKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load etcd cert/key pair")
	}

	// #nosec G402
	return &tls.Config{
		RootCAs: rootCertPool,
		Certificates: []tls.Certificate{
			cert,
		},
		MinVersion: tlsMinVersion,
	}, nil
}
