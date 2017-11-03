package controller

import (
	"log"

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

// The clusterController manages the lifecycle of a single cluster.
type clusterController struct {
	log                *log.Logger
	executor           install.Executor
	clusterStore       clusterStore
	generatedAssetsDir string
}

func (c *clusterController) run(clusterName string, watch <-chan planWrapper) {
	c.log.Println("started controller")
	for _ = range watch {
		c.log.Printf("got notification for cluster %q", clusterName)
		pw, err := c.clusterStore.Get(clusterName)
		if err != nil {
			c.log.Printf("error getting cluster from store: %v", err)
			continue
		}
		c.reconcile(clusterName, *pw)
	}
	c.log.Printf("stopping controller that was managing cluster %q", clusterName)
}

// reconcile the cluster / take it to the desired state
func (c *clusterController) reconcile(clusterName string, pw planWrapper) {
	c.log.Println("current state is:", pw.CurrentState, "desired state is:", pw.DesiredState)
	for pw.CurrentState != pw.DesiredState && pw.CanContinue {
		// transition cluster and update its state in the store
		pw.CurrentState, pw.CanContinue = c.next(pw)
		err := c.clusterStore.Put(clusterName, pw)
		if err != nil {
			c.log.Printf("error storing cluster state: %v. The cluster's current state is %q and desired state is %q", err, pw.CurrentState, pw.DesiredState)
			break
		}
	}

	if pw.CurrentState == pw.DesiredState {
		c.log.Printf("cluster %q reached desired state %q", clusterName, pw.DesiredState)
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
		c.log.Printf("the desired state is %q, but there is no transition defined for the cluster's current state %q", pw.DesiredState, pw.CurrentState)
		return pw.CurrentState, false
	}
}

func (c *clusterController) provision() (string, bool) {
	c.log.Println("provisioning infrastructure for cluster")
	return "provisioned", true
}

func (c *clusterController) destroy() (string, bool) {
	c.log.Println("destroying cluster")
	return "destroyed", true
}

func (c *clusterController) install(plan install.Plan) (string, bool) {
	c.log.Println("installing cluster")

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
