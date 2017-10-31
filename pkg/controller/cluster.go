package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/apprenda/kismatic/pkg/install"
)

type planWrapper struct {
	DesiredState string
	CurrentState string
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
	retry        bool
	log          *log.Logger
	executor     install.Executor
	clusterStore store
}

// New returns a controller that manages the lifecycle of the clusters that are
// defined in the cluster store
func New(l *log.Logger, e install.Executor, s store) ClusterController {
	return &clusterController{
		log:          l,
		executor:     e,
		clusterStore: s,
	}
}

// Run starts the controller.
func (c *clusterController) Run(ctx context.Context) error {
	c.log.Println("started controller")
	watch, err := c.clusterStore.Watch(context.Background(), []byte("clusters"))
	if err != nil {
		return fmt.Errorf("error creating watch on 'clusters': %v", err)
	}
	done := ctx.Done()
	for {
		select {
		case resp := <-watch:
			c.log.Printf("Got a watch notification for key: %s\n", string(resp.key))
			c.retry = true // we got a notification, so this should be set to true

			var pw planWrapper
			err := json.Unmarshal(resp.value, &pw)
			if err != nil {
				// TODO: need to think about this... what does it mean for the
				// client if we return the error here?
				panic("unexpected value found in store")
			}

			c.log.Println("Current state is:", pw.CurrentState, "Desired state is:", pw.DesiredState)
			for pw.CurrentState != pw.DesiredState && c.retry {
				// take the cluster to the next state, and update the store
				pw.CurrentState = c.next(pw)

				// TODO: Handle errors here
				b, _ := json.Marshal(pw)
				c.clusterStore.Put(resp.key, b)
			}
			if pw.CurrentState == pw.DesiredState {
				c.log.Println("reached desired state:", pw.DesiredState)
			}

		case <-done:
			c.log.Println("stopping the controller")
			return nil
		}
	}
}

// next transitions the cluster into the next state according to the desired
// state.
func (c *clusterController) next(pw planWrapper) string {
	switch pw.CurrentState {
	case planned:
		return provisioning
	case provisioning:
		return c.provision()
	case provisioned:
		if pw.DesiredState == destroyed {
			return destroying
		}
		return installing
	case provisionFailed:
		if pw.DesiredState == destroyed {
			return destroying
		}
		return provisioning
	case destroying:
		return c.destroy()
	case installing:
		return c.install(pw.Plan)
	case installFailed:
		if pw.DesiredState == destroyed {
			return destroying
		}
		return installing
	default:
		// TODO: Is there something else we can do here?
		panic(fmt.Sprintf("no next transition defined for %s", pw.CurrentState))
	}
}

func (c *clusterController) provision() string {
	c.log.Println("provisioning")
	return "provisioned"
}

func (c *clusterController) destroy() string {
	c.log.Println("destroying")
	return "destroyed"
}

func (c *clusterController) install(plan install.Plan) string {
	c.log.Println("installing")
	err := c.executor.Install(&plan)
	if err != nil {
		c.log.Printf("error installing cluster: %v\n", err)
		c.retry = false
		return installFailed
	}
	return installed
}
