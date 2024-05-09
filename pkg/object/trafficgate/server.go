package trafficgate

import (
	"context"
	"encoding/json"
	"fmt"
	"loong/pkg/global"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	spec            *Spec
	server          *http.Server
	shutdownTimeOut time.Duration
}

func NewServer(rawCfg any) (*Server, error) {
	spec := &Spec{}
	err := json.Unmarshal(rawCfg.([]byte), spec)
	if err != nil {
		return nil, err
	}
	if surMap[spec.Name] {
		return nil, fmt.Errorf("the trafficGate of name %s already exists", spec.Name)
	}
	surMap[spec.Name] = true
	err = global.GlobalValidator.Struct(spec)
	if err != nil {
		return nil, err
	}
	server := &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", spec.Port),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux.NewRouter(),
		},
		shutdownTimeOut: 10 * time.Second,
		spec:            spec,
	}
	Servers = append(Servers, server)
	return server, nil
}

func (s *Server) RegisterHandler(path string, handler http.Handler) {
	s.server.Handler.(*mux.Router).Handle(path, handler)
}

func (s *Server) RegisterMiddleWare() {
	s.server.Handler.(*mux.Router).Use(s.interceptRequest)
}

func (s *Server) StartServer() error {
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) ShutdownServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeOut)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) GetPath() []Paths {
	return s.spec.Paths
}

func (s *Server) GetPort() uint16 {
	return s.spec.Port
}

func (s *Server) GetName() string {
	return s.spec.Name
}

func (s *Server) interceptRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// global ipfilter
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		// blockIPs
		if len(s.spec.IPFilter.BlockIPs) != 0 && inSlice(host, s.spec.IPFilter.BlockIPs) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// allowIPs
		if len(s.spec.IPFilter.AllowIPs) != 0 && !inSlice(host, s.spec.IPFilter.AllowIPs) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// paths ipfilter
		for _, path := range s.spec.Paths {
			if path.Path == r.RequestURI {
				if len(path.BlockIPs) != 0 && inSlice(host, path.BlockIPs) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				if len(path.AllowIPs) != 0 && !inSlice(host, path.AllowIPs) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				if len(path.Methods) != 0 && !inSlice(r.Method, path.Methods) {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func inSlice(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
