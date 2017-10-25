package check

import (
	"errors"
	"fmt"
	"testing"
)

type runMock struct {
	aptGetOut string
	aptGetErr error
	yumOut    string
	yumErr    error
	dpkgOut   string
	dpkgErr   error
}

func (m runMock) run(cmd string, args ...string) ([]byte, error) {
	switch cmd {
	default:
		panic(fmt.Sprintf("mock does not implement command %s", cmd))
	case "apt-get":
		return []byte(m.aptGetOut), m.aptGetErr
	case "yum":
		return []byte(m.yumOut), m.yumErr
	case "dpkg":
		return []byte(m.dpkgOut), m.dpkgErr
	}
}

func TestRPMPackageManagerPackageAvailable(t *testing.T) {
	out := `
potentiallySomeGarbageData
389-ds-base.x86_64                        1.3.4.0-33.el7_2               updates
389-ds-base-devel.x86_64                  1.3.4.0-33.el7_2               updates
389-ds-base-libs.x86_64                   1.3.4.0-33.el7_2               updates
Cython.x86_64                             0.19-3.el7                     base
ElectricFence.i686                        2.2.2-39.el7                   base
NetworkManager.x86_64           1:1.0.6-30.el7_2               @koji-override-1`

	mock := runMock{
		yumOut: out,
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"NetworkManager", "1:1.0.6-30.el7_2"}
	ok, _ := m.IsAvailable(p)
	if !ok {
		t.Error("expected true, but got false")
	}
}

func TestRPMPackageManagerPackageNotFound(t *testing.T) {
	out := `Error: No matching Packages to list`
	mock := runMock{
		yumOut: out,
		yumErr: errors.New("yum exits with non-zero if no packages match"),
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"NonExistent", "1.0"}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err != nil {
		t.Errorf("got an unexpected error: %v", err)
	}
}

func TestRPMPackageManagerNameMatchVersionNoMatch(t *testing.T) {
	out := `
NetworkManager.x86_64           1:1.0.6-30.el7_2               @koji-override-1`
	mock := runMock{
		yumOut: out,
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"NetworkManager", "1:1.0.7-30.el7_2"}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err != nil {
		t.Errorf("got an unexpected error: %v", err)
	}
}

func TestRPMPackageManagerAnyVersion(t *testing.T) {
	out := `
NetworkManager.x86_64           1:1.0.6-30.el7_2               @koji-override-1`
	mock := runMock{
		yumOut: out,
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"NetworkManager", ""}
	ok, err := m.IsAvailable(p)
	if !ok {
		t.Error("expected true, but got false")
	}
	if err != nil {
		t.Errorf("got an unexpected error: %v", err)
	}
}

func TestRPMPackageManagerNameNoMatchVersionMatch(t *testing.T) {
	out := `
NetworkManager.x86_64           1:1.0.6-30.el7_2               @koji-override-1`
	mock := runMock{
		yumOut: out,
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"NetworkManagr", "1:1.0.6-30.el7_2"}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err != nil {
		t.Errorf("got an unexpected error: %v", err)
	}
}

func TestRPMPackageManagerExecError(t *testing.T) {
	mock := runMock{
		yumErr: fmt.Errorf("some error"),
	}
	m := rpmManager{
		run: mock.run,
	}
	p := PackageQuery{"SomePkg", "1.0"}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err == nil {
		t.Error("expected an error, but didn't get one")
	}
}

func TestDebPackageManagerIsInstalled(t *testing.T) {
	out := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name                                                  Version                         Architecture                    Description
+++-=====================================================-===============================-===============================-================================================================================================================
ii  libc6:amd64                                           2.23-0ubuntu3                   amd64                           GNU C Library: Shared libraries`
	mock := runMock{
		dpkgOut: out,
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"libc6", "2.23"}
	ok, _ := m.IsAvailable(p)
	if !ok {
		t.Errorf("expected true, but got false")
	}
}

func TestDebPackageManagerAnyVersion(t *testing.T) {
	out := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name                                                  Version                         Architecture                    Description
+++-=====================================================-===============================-===============================-================================================================================================================
ii  libc6:amd64                                           2.23-0ubuntu3                   amd64                           GNU C Library: Shared libraries`
	mock := runMock{
		dpkgOut: out,
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"libc6", ""}
	ok, _ := m.IsAvailable(p)
	if !ok {
		t.Errorf("expected true, but got false")
	}
}

func TestDebPackageManagerPackageNotInstalledButAvailable(t *testing.T) {
	mock := runMock{
		dpkgOut:   "dpkg-query: no packages found matching libc6a",
		dpkgErr:   errors.New("dpkg returns error msg and exits non-zero in this case"),
		aptGetOut: "we don't really care about this output, just the non-zero exit status",
		aptGetErr: nil,
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"libc6a", "1.0"}
	ok, err := m.IsAvailable(p)
	if !ok {
		t.Errorf("expected true, got false")
	}
	if err != nil {
		t.Errorf("unexpected error occurred")
	}
}

func TestDebPackageManagerPackageNotInstalledNotAvailable(t *testing.T) {
	mock := runMock{
		dpkgOut:   "dpkg-query: no packages found matching libc6a",
		dpkgErr:   errors.New("dpkg returns error msg and exits non-zero in this case"),
		aptGetErr: errors.New("apt-get returns error msg and exits non-zero when package not found"),
		aptGetOut: "E: Unable to locate package libc6a",
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"libc6a", "1.0"}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Errorf("expected false, but got true")
	}
	if err != nil {
		t.Errorf("got an unexpected error: %v", err)
	}
}

func TestDebPackageManagerExecError(t *testing.T) {
	mock := runMock{
		dpkgErr: errors.New("some error happened"),
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"", ""}
	ok, err := m.IsInstalled(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err == nil {
		t.Error("expected an error, but didn't get one")
	}
}

func TestDebPackageManagerExecError2(t *testing.T) {
	mock := runMock{
		aptGetErr: errors.New("some error happened"),
	}
	m := debManager{
		run: mock.run,
	}
	p := PackageQuery{"", ""}
	ok, err := m.IsAvailable(p)
	if ok {
		t.Error("expected false, but got true")
	}
	if err == nil {
		t.Error("expected an error, but didn't get one")
	}
}
