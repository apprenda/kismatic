/*
Package controller implements a controller that manages the lifecycle of
Kubernetes clusters. The controller connects to a store where cluster resources
are defined. Whenever there is a change in the store, the controller takes
action.

The controller's only mission is to take the defined clusters from their current
states to their desired states. While the clusters transition between the different
states, the controller updates the cluster definitions in the store.
*/
package controller
