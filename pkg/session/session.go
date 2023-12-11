package session

import (
	"context"
	"encoding/json"
	"fmt"
	"milvus-k8s-server/pkg/common/kv"
	"net"
	"path"
	"time"
)

const (
	sessionPrefix = `session`
)

// ListSessions returns all session.
func ListSessions(cli kv.MetaKV) ([]*Session, error) {
	return ListSessionsByPrefix(cli, "")
}

// ListSessionsByPrefix returns all session with provided prefix.
func ListSessionsByPrefix(cli kv.MetaKV, prefix string) ([]*Session, error) {
	prefix = path.Join(sessionPrefix, prefix)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, vals, err := cli.LoadWithPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0, len(vals))
	for _, val := range vals {
		session := &Session{}
		err := json.Unmarshal([]byte(val), session)
		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}

// Session is the json model for milvus session struct in etcd.
type Session struct {
	ServerID   int64  `json:"ServerID,omitempty"`
	ServerName string `json:"ServerName,omitempty"`
	Address    string `json:"Address,omitempty"`
	Exclusive  bool   `json:"Exclusive,omitempty"`
	Version    string `json:"Version,omitempty"`
}

func (s Session) String() string {
	return fmt.Sprintf("Session:%s, ServerID: %d, Version: %s, Address: %s", s.ServerName, s.ServerID, s.Version, s.Address)
}

func (s Session) IP() string {
	addr, err := net.ResolveTCPAddr("tcp", s.Address)
	if err != nil {
		return ""
	}
	return addr.IP.To4().String()
}
