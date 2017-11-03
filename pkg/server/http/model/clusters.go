package model

import "github.com/apprenda/kismatic/pkg/install"

type ClusterRequest struct {
	Name         string
	DesiredState string
	AwsID        string
	AwsKey       string
	Etcd         int
	Master       int
	Worker       int
}

type ClusterResponse struct {
	Name         string
	DesiredState string
	CurrentState string
	install.Plan
}
