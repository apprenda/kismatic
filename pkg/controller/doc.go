/*
Package controller implements a controller that manages the lifecycle of
Kubernetes clusters. The controller connects to a store where cluster resources
are defined. Whenever there is a change in the store, the controller takes
action.

The controller's only mission is to take the defined cluster from its current
state to its desired state. While the cluster transitions between the different
states, the controller will update the cluster definition in the store.
*/
package controller
