package workers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// APIOptions contains the set of configuration options for the global api
type APIOptions struct {
	Logger *log.Logger
}

type apiServer struct {
	lock     sync.Mutex
	managers map[string]*Manager
	logger   *log.Logger
}

func (s *apiServer) registerManager(m *Manager) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.managers[m.uuid] = m
}

func (s *apiServer) deregisterManager(m *Manager) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.managers, m.uuid)
}

var globalHTTPServer *http.Server

var globalAPIServer = &apiServer{
	managers: map[string]*Manager{},
	logger:   log.New(os.Stdout, "go-workers2: ", log.Ldate|log.Lmicroseconds),
}

// ConfigureAPIServer allows global API server configuration with the given options
func ConfigureAPIServer(options APIOptions) {
	if options.Logger != nil {
		globalAPIServer.logger = options.Logger
	}
}

// RegisterAPIEndpoints sets up API server endpoints
func RegisterAPIEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/stats", globalAPIServer.Stats)
	mux.HandleFunc("/retries", globalAPIServer.Retries)
}

// StartAPIServer starts the API server
func StartAPIServer(port int) {
	mux := http.NewServeMux()
	RegisterAPIEndpoints(mux)

	globalAPIServer.logger.Println("APIs are available at", fmt.Sprintf("http://localhost:%v/", port))

	globalHTTPServer = &http.Server{Addr: fmt.Sprint(":", port), Handler: mux}
	if err := globalHTTPServer.ListenAndServe(); err != nil {
		globalAPIServer.logger.Println(err)
	}
}

// StopAPIServer stops the API server
func StopAPIServer() {
	if globalHTTPServer != nil {
		globalHTTPServer.Shutdown(context.Background())
	}
}
