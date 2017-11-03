package store

import "github.com/apprenda/kismatic/pkg/install"

type Cluster struct {
	DesiredState string
	CurrentState string
	CanContinue  bool
	Plan         install.Plan
	AwsID        string
	AwsKey       string
}
