package handler

import (
	"fmt"

	"github.com/apprenda/kismatic/pkg/store"
	"github.com/apprenda/kismatic/pkg/util"
)

type validatable interface {
	validate() (bool, []error)
}

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{
		errs: []error{},
	}
}

func (v *validator) addError(err ...error) {
	v.errs = append(v.errs, err...)
}

func (v *validator) validate(obj validatable) {
	if ok, err := obj.validate(); !ok {
		v.addError(err...)
	}
}

func (v *validator) valid() (bool, []error) {
	if len(v.errs) > 0 {
		return false, v.errs
	}
	return true, nil
}

func (r *ClusterRequest) validate() (bool, []error) {
	v := newValidator()
	if r.Name == "" {
		v.addError(fmt.Errorf("name cannot be empty"))
	}
	if r.DesiredState == "" {
		v.addError(fmt.Errorf("desiredState cannot be empty"))
	} else {
		if !util.Contains(r.DesiredState, validStates) {
			v.addError(fmt.Errorf("%s is not a valid desiredState, options are: %v", r.DesiredState, validStates))
		}
	}
	if r.EtcdCount <= 0 {
		v.addError(fmt.Errorf("cluster.etcdCount must be greater than 0"))
	}
	if r.MasterCount <= 0 {
		v.addError(fmt.Errorf("cluster.masterCount must be greater than 0"))
	}
	if r.WorkerCount <= 0 {
		v.addError(fmt.Errorf("cluster.workerCount must be greater than 0"))
	}
	if r.IngressCount < 0 {
		v.addError(fmt.Errorf("cluster.ingressCount must be greater than or equal to 0"))
	}
	v.validate(&r.Provisioner)
	return v.valid()
}

func (p *Provisioner) validate() (bool, []error) {
	v := newValidator()
	if p.Provider == "" {
		v.addError(fmt.Errorf("provisioner.provider cannot be empty"))
	} else {
		if !util.Contains(p.Provider, validProvisionerProviders) {
			v.addError(fmt.Errorf("%s is not a valid provisioner.provider, options are: %v", p.Provider, validProvisionerProviders))
		}
		switch p.Provider {
		case "aws":
			if p.AWSOptions == nil || p.AWSOptions.AccessKeyID == "" {
				v.addError(fmt.Errorf("provisioner.options.accessKeyID cannot be empty"))
			}
			if p.AWSOptions == nil || p.AWSOptions.SecretAccessKey == "" {
				v.addError(fmt.Errorf("provisioner.options.secretAccessKey cannot be empty"))
			}
		}
	}
	return v.valid()
}

// validate that the requested changes can be done against the existing cluster
type clusterUpdate struct {
	id      string
	request ClusterRequest
	inStore store.Cluster
}

func (c *clusterUpdate) validate() (bool, []error) {
	v := newValidator()
	if c.id != c.request.Name {
		v.addError(fmt.Errorf("name must match the cluster requested"))
	}
	if c.request.DesiredState == "" {
		v.addError(fmt.Errorf("desiredState cannot be empty"))
	} else {
		if !util.Contains(c.request.DesiredState, validStates) {
			v.addError(fmt.Errorf("%s is not a valid desiredState, options are: %v", c.request.DesiredState, validStates))
		}
	}
	if c.request.EtcdCount != 0 && (c.request.EtcdCount != c.inStore.Spec.EtcdCount) {
		v.addError(fmt.Errorf("cluster.etcdCount cannot be modified"))
	}
	// allow adding/removing of master, worker or ingress nodes
	if c.request.MasterCount == 0 {
		v.addError(fmt.Errorf("cluster.masterCount must be greater than 0"))
	}
	if c.request.WorkerCount == 0 {
		v.addError(fmt.Errorf("cluster.workerCount must be greater than 0"))
	}
	if c.request.IngressCount < 0 {
		v.addError(fmt.Errorf("cluster.ingressCount must be greater than or equal to 0"))
	}
	// always require provisioner credentials
	v.validate(&c.request.Provisioner)
	return v.valid()
}

func formatErrs(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, err := range errs {
		out = append(out, err.Error())
	}
	return out
}
