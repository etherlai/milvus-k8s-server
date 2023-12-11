package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"log"
	"milvus-k8s-server/pkg/configs"
	"milvus-k8s-server/pkg/querynode"
	"net/http"
	"strconv"
)

// Server contains the runtime configuration for the Pilot discovery service.
type Server struct {
	//httpServer   *http.Server // debug
	mux *mux.Router // debug
	//HTTPListener net.Listener

	basePort int
	qnMgr    *querynode.QueryNodeManager
}

// NewServer creates a new Server instance based on the provided arguments.
func NewServer(cfg *configs.Config) (*Server, error) {
	qnMgr, err := querynode.NewQueryNodeManager(cfg)
	if err != nil {
		log.Fatal("querynode manager init failed", err.Error())
	}
	s := &Server{
		basePort: cfg.ServerPort,
		qnMgr:    qnMgr,
	}

	if err := s.initService(); err != nil {
		return nil, fmt.Errorf("service: %v", err)
	}

	return s, nil
}

func (s *Server) Start() error {
	// At this point we are ready - start Http Listener so that it can respond to readiness events.
	log.Printf("starting Http service at :%s\n", s.basePort)
	if err := http.ListenAndServe(":"+strconv.Itoa(s.basePort), s.mux); err != nil {
		return err
	}

	return nil
}

func (s *Server) initService() error {
	s.mux = mux.NewRouter()
	s.mux.HandleFunc("/health", s.httpServerReadyHandler)
	s.mux.HandleFunc("/querynodes", s.getAllQueryNodes)

	return nil
}

func (s *Server) httpServerReadyHandler(w http.ResponseWriter, _ *http.Request) {
	// TODO check readiness of other secure gRPC and HTTP servers.
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getAllQueryNodes(w http.ResponseWriter, _ *http.Request) {
	qns, err := s.qnMgr.GetAllQueryNodes()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(qns)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
