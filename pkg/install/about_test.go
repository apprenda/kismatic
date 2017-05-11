package install

import (
	"testing"

	"github.com/blang/semver"
)

func TestCompleteNoiseString(t *testing.T) {
	_, err := parseVersion("abbazabba")
	if err == nil {
		t.Errorf("did not catch invalid version")
	}
}

func TestThreeDigits(t *testing.T) {
	ver := mustParseVersion("1.2.3")

	if ver.Major != 1 || ver.Minor != 2 || ver.Patch != 3 {
		t.Errorf("Blank string didn't parse to version 1.2.3, instead was %v", ver)
	}
}

func makeVersions(ver, testVer string) (semver.Version, semver.Version) {
	return mustParseVersion(ver), mustParseVersion(testVer)
}

func TestIsSameMajorMinorPatchNewer(t *testing.T) {
	ver, testVer := makeVersions("1.2.3", "1.2.3")

	if ver.GT(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsOlderMajorMinorPatchOlder(t *testing.T) {
	ver, testVer := makeVersions("1.2.2", "1.2.3")

	if ver.GT(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestSetVersion(t *testing.T) {
	SetVersion("1.2.3")
	v := KismaticVersion
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("expected 1.2.3, but got %v", v)
	}
}

func TestSetInvalidVersion(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("did not panic")
		}
	}()
	SetVersion("1.2.a")
}

func TestIsOlderVersion(t *testing.T) {
	SetVersion("1.2.0")
	if IsOlderVersion(KismaticVersion) {
		t.Error("IsOlder returned true for the same version")
	}
}

func TestGitDescribeVersion(t *testing.T) {
	ver := mustParseVersion("v1.2.2-69-g51dfe53-dirty")
	if ver.Major != 1 || ver.Minor != 2 || ver.Patch != 2 {
		t.Errorf("didn't parse to version 1.2.2, instead got %v", ver)
	}
	if ver.Pre[0].VersionStr != "69-g51dfe53-dirty" {
		t.Errorf("didn't parse pre-release to 69-g51dfe53-dirty, instead got %s", ver.Pre[0])
	}
}

func TestPreReleaseVersion(t *testing.T) {
	ver := mustParseVersion("0.0.1-alpha")
	if ver.Major != 0 || ver.Minor != 0 || ver.Patch != 1 {
		t.Errorf("Didn't parse to version 0.0.1, instead got %v", ver)
	}
	if ver.Pre[0].VersionStr != "alpha" {
		t.Errorf("Didn't parse pre-release version to alpha, instead got %s", ver.Pre)
	}
}

func TestVersionPrecedence(t *testing.T) {
	// from semver spec:
	// 1.0.0 < 2.0.0 < 2.1.0 < 2.1.1
	// 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
	order := []semver.Version{
		mustParseVersion("1.0.0-alpha"),
		mustParseVersion("1.0.0-alpha.1"),
		mustParseVersion("1.0.0-alpha.beta"),
		mustParseVersion("1.0.0-beta"),
		mustParseVersion("1.0.0-beta.2"),
		mustParseVersion("1.0.0-beta.11"),
		mustParseVersion("1.0.0-rc.1"),
		mustParseVersion("1.0.0"),
		mustParseVersion("1.2.3"),
		mustParseVersion("1.3.0-alpha"),
		mustParseVersion("1.3.0"),
		mustParseVersion("1.3.1"),
		mustParseVersion("1.4.999"),
		mustParseVersion("2.0.0"),
		mustParseVersion("2.1.0"),
		mustParseVersion("2.1.1"),
	}

	for i := 0; i < len(order); i++ {
		for j := i; j < len(order); j++ {
			if order[i].GT(order[j]) {
				t.Errorf("expected %s < %s but was not the case", order[i], order[j])
			}
		}
	}
}

func mustParseVersion(v string) semver.Version {
	ver, err := parseVersion(v)
	if err != nil {
		panic("failed to parse version: " + v)
	}
	return ver
}
