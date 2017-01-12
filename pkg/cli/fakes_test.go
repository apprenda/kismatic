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
func (fp *fakePlanner) GetClusterAddress(p *install.Plan) (string, error) {
	return fp.clusterAddress, fp.clusterAddressErr
}
func (fp *fakePlanner) GetDashboardURL(p *install.Plan) (string, error) {
	return fp.dashboardURL, fp.dashboardURLErr
}

type fakeExecutor struct {
	installCalled bool
	err           error
}

func (fe *fakeExecutor) AddWorker(p *install.Plan, newWorker install.Node) (*install.Plan, error) {
	return nil, nil
}

func (fe *fakeExecutor) Install(p *install.Plan) error {
	fe.installCalled = true
	return fe.err
}

func (fe *fakeExecutor) RunPreFlightCheck(p *install.Plan) error {
	return nil
}

func (fe *fakeExecutor) RunSmokeTest(p *install.Plan) error {
	return nil
}

func (fe *fakeExecutor) RunTask(string, *install.Plan) error {
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
