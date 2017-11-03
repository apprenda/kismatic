package controller

import (
	"log"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/store"
)

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

func (c *clusterController) run(clusterName string, watch <-chan struct{}) {
	c.log.Println("started controller")
	for _ = range watch {
		c.log.Printf("got notification for cluster %q", clusterName)
		cluster, err := c.clusterStore.Get(clusterName)
		if err != nil {
			c.log.Printf("error getting cluster from store: %v", err)
			continue
		}
		c.reconcile(clusterName, *cluster)
	}
	c.log.Printf("stopping controller that was managing cluster %q", clusterName)
}

// reconcile the cluster / take it to the desired state
func (c *clusterController) reconcile(clusterName string, cluster store.Cluster) {
	c.log.Println("current state is:", cluster.CurrentState, "desired state is:", cluster.DesiredState)
	for cluster.CurrentState != cluster.DesiredState && cluster.CanContinue {
		// transition cluster and update its state in the store
		cluster.CurrentState, cluster.CanContinue = c.next(cluster)
		err := c.clusterStore.Put(clusterName, cluster)
		if err != nil {
			c.log.Printf("error storing cluster state: %v. The cluster's current state is %q and desired state is %q", err, cluster.CurrentState, cluster.DesiredState)
			break
		}
	}

	if cluster.CurrentState == cluster.DesiredState {
		c.log.Printf("cluster %q reached desired state %q", clusterName, cluster.DesiredState)
	}
}

// next transitions the cluster into the next state according to the desired
// state. It returns the cluster's state after the transition, and whether it
// can continue transitioning the cluster to the desired state. In the case of
// an error, it will return false, as we do not currently support retries.
func (c *clusterController) next(cluster store.Cluster) (string, bool) {
	switch cluster.CurrentState {
	case planned:
		return provisioning, true
	case provisioning:
		return c.provision()
	case provisioned:
		if cluster.DesiredState == destroyed {
			return destroying, true
		}
		return installing, true
	case provisionFailed:
		if cluster.DesiredState == destroyed {
			return destroying, true
		}
		return provisioning, true
	case destroying:
		return c.destroy()
	case installing:
		return c.install(cluster.Plan)
	case installFailed:
		if cluster.DesiredState == destroyed {
			return destroying, true
		}
		return installing, true
	default:
		// Log a message, and return false so that we don't get stuck in an
		// infinte loop. The only thing the user can do in this case is delete
		// the cluster and file a bug, as this scenario should not happen.
		c.log.Printf("the desired state is %q, but there is no transition defined for the cluster's current state %q", cluster.DesiredState, cluster.CurrentState)
		return cluster.CurrentState, false
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
