package inspector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/apprenda/kismatic/pkg/inspector/check"
	"github.com/apprenda/kismatic/pkg/inspector/rule"
)

// Server supports the execution of inspector rules from a remote node
type Server struct {
	// The Port the server will listen on
	Port int
	// NodeFacts are the facts that apply to the node where the server is running
	NodeFacts []string
	// RulesEngine for running inspector rules
	rulesEngine *rule.Engine
}

type serverError struct {
	Error string
}

var executeEndpoint = "/execute"
var closeEndpoint = "/close"

// NewServer returns an inspector server that has been initialized
// with the default rules engine
func NewServer(nodeFacts []string, port int, packageInstallationDisabled bool, dockerInstallationDisabled bool) (*Server, error) {
	s := &Server{
		Port: port,
	}
	distro, err := check.DetectDistro()
	if err != nil {
		return nil, fmt.Errorf("error building server: %v", err)
	}
	s.NodeFacts = append(nodeFacts, string(distro))
	pkgMgr, err := check.NewPackageManager(distro)
	if err != nil {
		return nil, fmt.Errorf("error building server: %v", err)
	}
	engine := &rule.Engine{
		RuleCheckMapper: rule.DefaultCheckMapper{
			PackageManager:              pkgMgr,
			PackageInstallationDisabled: packageInstallationDisabled,
			DockerInstallationDisabled:  dockerInstallationDisabled,
		},
	}
	s.rulesEngine = engine
	return s, nil
}

// Start the server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	// Execute endpoint
	mux.HandleFunc(executeEndpoint, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Decode rules
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error decoding rules when processing request: %v", err)
			return
		}
		defer req.Body.Close()
		rules, err := rule.UnmarshalRulesJSON(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error unmarshaling rules from JSON: %v", err)
			return
		}
		// Run the rules that we received
		results, err := s.rulesEngine.ExecuteRules(rules, s.NodeFacts)
		if err != nil {
			err = json.NewEncoder(w).Encode(serverError{Error: err.Error()})
			if err != nil {
				log.Printf("error writing server response: %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			log.Printf("error writing server response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	// Close endpoint
	mux.HandleFunc(closeEndpoint, func(w http.ResponseWriter, req *http.Request) {
		err := s.rulesEngine.CloseChecks()
		if err != nil {
			log.Printf("error closing checks: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
