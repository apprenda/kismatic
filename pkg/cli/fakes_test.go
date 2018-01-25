package cli

import (
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/tls"
)

type fakePlanner struct {
	exists            bool
	plan              *install.Plan
	readWriteErr      error
	readCalled        bool
	clusterAddress    string
	clusterAddressErr error
	dashboardURL      string
	dashboardURLErr   error
}

func (fp *fakePlanner) PlanExists() bool { return fp.exists }
func (fp *fakePlanner) Read() (*install.Plan, error) {
	fp.readCalled = true
	return fp.plan, fp.readWriteErr
}
func (fp *fakePlanner) Write(p *install.Plan) error {
	fp.plan = p
	return fp.readWriteErr
}

type fakeExecutor struct {
	installCalled bool
	err           error
}

func (fe *fakeExecutor) AddNode(p *install.Plan, newWorker install.Node, roles []string) (*install.Plan, error) {
	return nil, nil
}

func (fe *fakeExecutor) GenerateCertificates(*install.Plan, bool) error {
	return nil
}

func (fe *fakeExecutor) Install(p *install.Plan) error {
	fe.installCalled = true
	return fe.err
}

func (fe *fakeExecutor) RunPreFlightCheck(p *install.Plan) error {
	return nil
}

func (fe *fakeExecutor) RunNewNodePreFlightCheck(install.Plan, install.Node) error {
	return nil
}

func (fe *fakeExecutor) RunUpgradePreFlightCheck(*install.Plan, install.ListableNode) error {
	return nil
}

func (fe *fakeExecutor) UpgradeNodes(install.Plan, []install.ListableNode, bool, int) error {
	return nil
}

func (fe *fakeExecutor) ValidateControlPlane(install.Plan) error {
	return nil
}

func (fe *fakeExecutor) UpgradeDockerRegistry(install.Plan) error {
	return nil
}

func (fe *fakeExecutor) UpgradeClusterServices(install.Plan) error {
	return nil
}

func (fe *fakeExecutor) RunSmokeTest(p *install.Plan) error {
	return nil
}

func (fe *fakeExecutor) RunPlay(string, *install.Plan) error {
	return nil
}

func (fe *fakeExecutor) AddVolume(*install.Plan, install.StorageVolume) error {
	return nil
}

func (fe *fakeExecutor) DeleteVolume(*install.Plan, string) error {
	return nil
}

type fakePKI struct {
	called              bool
	generateCACalled    bool
	readClusterCACalled bool
	err                 error
}

func (fp *fakePKI) ReadClusterCA(p *install.Plan) (*tls.CA, error) {
	fp.readClusterCACalled = true
	return nil, fp.err
}
func (fp *fakePKI) GenerateClusterCA(p *install.Plan) (*tls.CA, error) {
	fp.generateCACalled = true
	return nil, fp.err
}

func (fp *fakePKI) GenerateClusterCerts(p *install.Plan, ca *tls.CA, users []string) error {
	fp.called = true
	return fp.err
}

func (fp *fakePKI) GenerateCertificate(name string, validityPeriod string, commonName string, subjectAlternateNames []string, organizations []string, ca *tls.CA, overwrite bool) (bool, error) {
	fp.called = true
	return false, fp.err
}
