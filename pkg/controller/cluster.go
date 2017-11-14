package controller

import (
	"log"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/provision"
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
	log            *log.Logger
	executor       install.Executor
	newProvisioner func(store.Cluster) provision.Provisioner
	clusterStore   store.ClusterStore
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
		cluster = c.transition(cluster)
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

// transition performs an action to take the cluster to the next state. The
// action to be performed depends on the current state and the desired state.
// Once the action is done, an updated cluster spec is returned that reflects
// the outcome of the action.
func (c *clusterController) transition(cluster store.Cluster) store.Cluster {
	if cluster.CurrentState == cluster.DesiredState {
		panic("cannot transition a cluster that is already at it's desired state")
	}
	switch cluster.CurrentState {
	case planned:
		cluster.CurrentState = provisioning
		return cluster
	case provisioning:
		return c.provision(cluster)
	case provisioned:
		if cluster.DesiredState == destroyed {
			cluster.CurrentState = destroying
			return cluster
		}
		cluster.CurrentState = installing
		return cluster
	case provisionFailed:
		if cluster.DesiredState == destroyed {
			cluster.CurrentState = destroying
			return cluster
		}
		cluster.CurrentState = provisioning
		return cluster
	case destroying:
		return c.destroy(cluster)
	case installing:
		return c.install(cluster)
	case installFailed:
		if cluster.DesiredState == destroyed {
			cluster.CurrentState = destroying
			return cluster
		}
		cluster.CurrentState = installing
		return cluster
	default:
		// Log a message, and set CanContinue to false so that we don't get
		// stuck in an infinte loop. The only thing the user can do in this case
		// is delete the cluster and file a bug, as this scenario should not
		// happen.
		c.log.Printf("the desired state is %q, but there is no transition defined for the cluster's current state %q", cluster.DesiredState, cluster.CurrentState)
		cluster.CanContinue = false
		return cluster
	}
}

func (c *clusterController) provision(cluster store.Cluster) store.Cluster {
	c.log.Println("provisioning infrastructure for cluster")
	provisioner := c.newProvisioner(cluster)
	updatedPlan, err := provisioner.Provision(cluster.Plan)
	if err != nil {
		c.log.Printf("error provisioning: %v", err)
		cluster.CurrentState = provisionFailed
		cluster.CanContinue = false
		return cluster
	}
	cluster.Plan = *updatedPlan
	cluster.CurrentState = provisioned
	return cluster
}

func (c *clusterController) destroy(cluster store.Cluster) store.Cluster {
	c.log.Println("destroying cluster")
	return cluster
}

func (c *clusterController) install(cluster store.Cluster) store.Cluster {
	c.log.Println("installing cluster")
	plan := cluster.Plan

	err := c.executor.RunPreFlightCheck(&plan)
	if err != nil {
		c.log.Printf("error running preflight checks: %v", err)
		cluster.CurrentState = installFailed
		cluster.CanContinue = false
		return cluster
	}

	err = c.executor.GenerateCertificates(&plan, false)
	if err != nil {
		c.log.Printf("error generating certificates: %v", err)
		cluster.CurrentState = installFailed
		cluster.CanContinue = false
		return cluster
	}

	err = c.executor.GenerateKubeconfig(plan)
	if err != nil {
		c.log.Printf("error generating kubeconfig file: %v", err)
		cluster.CurrentState = installFailed
		cluster.CanContinue = false
		return cluster
	}

	err = c.executor.Install(&plan, true)
	if err != nil {
		c.log.Printf("error installing the cluster: %v", err)
		cluster.CurrentState = installFailed
		cluster.CanContinue = false
		return cluster
	}

	// Skip the smoketest if the user asked us to skip the installation of a
	// networking stack
	if !plan.NetworkConfigured() {
		cluster.CurrentState = installed
		return cluster
	}

	err = c.executor.RunSmokeTest(&plan)
	if err != nil {
		c.log.Printf("error running smoke test against the cluster: %v", err)
		cluster.CurrentState = installFailed
		return cluster
	}

	cluster.CurrentState = installed
	return cluster
}
