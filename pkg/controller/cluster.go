package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/apprenda/kismatic/pkg/install"
)

type planWrapper struct {
	DesiredState string
	CurrentState string
	CanContinue  bool
	Plan         install.Plan
}

const (
	planned         = "planned"
	provisioning    = "provisioning"
	provisionFailed = "provisionFailed"
	provisioned     = "provisioned"
	installing      = "installing"
	installFailed   = "installFailed"
	installed       = "installed"
	modifying       = "modifying"
	modifyFailed    = "modifyFailed"
	destroying      = "destroying"
	destroyed       = "destroyed"
)

// The ClusterController manages the lifecycle of a given cluster
type ClusterController interface {
	Run(ctx context.Context) error
}

type watchChan <-chan watchResponse

type watchResponse struct {
	key   []byte
	value []byte
}

type store interface {
	Watch(ctx context.Context, bucket []byte) (watchChan, error)
	Put(key []byte, value []byte) error
}

type clusterController struct {
	log                *log.Logger
	executor           install.Executor
	clusterStore       store
	reconcileFreq      time.Duration
	generatedAssetsDir string
}

// New returns a controller that manages the lifecycle of the clusters that are
// defined in the cluster store
func New(l *log.Logger, e install.Executor, s store) ClusterController {
	return &clusterController{
		log:           l,
		executor:      e,
		clusterStore:  s,
		reconcileFreq: 10 * time.Minute,
	}
}

// Run starts the controller. If there is an issue starting the controller, an
// error is returned. Otherwise, the controller will run until the context is
// cancelled.
func (c *clusterController) Run(ctx context.Context) error {
	c.log.Println("started controller")
	watch, err := c.clusterStore.Watch(context.Background(), []byte("clusters"))
	if err != nil {
		return fmt.Errorf("error creating watch on 'clusters': %v", err)
	}
	ticker := time.Tick(c.reconcileFreq)
	for {
		select {
		case resp := <-watch:
			c.log.Printf("Got a watch event for key: %s", string(resp.key))

			var pw planWrapper
			err := json.Unmarshal(resp.value, &pw)
			if err != nil {
				c.log.Printf("error unmarshaling watch event's value: %v", err)
				continue
			}

			c.log.Println("Current state is:", pw.CurrentState, "Desired state is:", pw.DesiredState)
			for pw.CurrentState != pw.DesiredState && pw.CanContinue {
				// take the cluster to the next state, and update the store
				pw.CurrentState, pw.CanContinue = c.next(pw)

				// Update the store with the latest state
				b, err := json.Marshal(pw)
				if err != nil {
					c.log.Printf("error marshaling cluster resource: %v. The cluster's current state is %q and desired state is %q", err, pw.CurrentState, pw.DesiredState)
					break
				}
				err = c.clusterStore.Put(resp.key, b)
				if err != nil {
					c.log.Printf("error storing cluster state: %v. The cluster's current state is %q and desired state is %q", err, pw.CurrentState, pw.DesiredState)
					break
				}
			}

			if pw.CurrentState == pw.DesiredState {
				c.log.Println("reached desired state:", pw.DesiredState)
			}
		case <-ticker:
			c.log.Println("tick")
			// TODO: Check if action must be taken
		case <-ctx.Done():
			c.log.Println("stopping the controller")
			return nil
		}
	}
}

// next transitions the cluster into the next state according to the desired
// state. It returns the cluster's state after the transition, and whether it
// can continue transitioning the cluster to the desired state. In the case of
// an error, it will return false, as we do not currently support retries.
func (c *clusterController) next(pw planWrapper) (string, bool) {
	switch pw.CurrentState {
	case planned:
		return provisioning, true
	case provisioning:
		return c.provision()
	case provisioned:
		if pw.DesiredState == destroyed {
			return destroying, true
		}
		return installing, true
	case provisionFailed:
		if pw.DesiredState == destroyed {
			return destroying, true
		}
		return provisioning, true
	case destroying:
		return c.destroy()
	case installing:
		return c.install(pw.Plan)
	case installFailed:
		if pw.DesiredState == destroyed {
			return destroying, true
		}
		return installing, true
	default:
		// Log a message, and return false so that we don't get stuck in an
		// infinte loop. The only thing the user can do in this case is delete
		// the cluster and file a bug, as this scenario should not happen.
		c.log.Printf("The desired state is %q, but there is no transition defined for the cluster's current state %q", pw.DesiredState, pw.CurrentState)
		return pw.CurrentState, false
	}
}

func (c *clusterController) provision() (string, bool) {
	c.log.Println("provisioning")
	return "provisioned", true
}

func (c *clusterController) destroy() (string, bool) {
	c.log.Println("destroying")
	return "destroyed", true
}

func (c *clusterController) install(plan install.Plan) (string, bool) {
	c.log.Println("installing")

	// TODO: Run validation here, or create "validating", "validationFailed" states?

	err := c.executor.GenerateCertificates(&plan, false)
	if err != nil {
		c.log.Printf("error generating certificates: %v", err)
		return installFailed, false
	}

	err = c.executor.GenerateKubeconfig(plan, c.generatedAssetsDir)
	if err != nil {
		c.log.Printf("error generating kubeconfig file: %v", err)
		return installFailed, false
	}

	err = c.executor.Install(&plan)
	if err != nil {
		c.log.Printf("error installing the cluster: %v", err)
		return installFailed, false
	}

	// Skip the smoketest if the user asked us to skip the installation of a
	// networking stack
	if !plan.NetworkConfigured() {
		return installed, true
	}

	err = c.executor.RunSmokeTest(&plan)
	if err != nil {
		c.log.Printf("error running smoke test against the cluster: %v", err)
		return installFailed, false
	}

	return installed, true
}
