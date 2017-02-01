package install

import "testing"

func TestParseEmptyString(t *testing.T) {
	ver := parseVersion("")

	if ver.Major != 0 || ver.Minor != 0 || ver.Patch != 0 {
		t.Errorf("Blank string didn't parse to version 0, instead was %v", ver)
	}
}

func TestCompleteNoiseString(t *testing.T) {
	ver := parseVersion("abbazabba")

	if ver.Major != 0 || ver.Minor != 0 || ver.Patch != 0 {
		t.Errorf("Nonsense string didn't parse to version 0, instead was %v", ver)
	}
}

func TestSingleDigit(t *testing.T) {
	ver := parseVersion("1")

	if ver.Major != 1 || ver.Minor != 0 || ver.Patch != 0 {
		t.Errorf("Blank string didn't parse to version 1.0.0, instead was %v", ver)
	}
}

func TestTwoDigits(t *testing.T) {
	ver := parseVersion("1.2")

	if ver.Major != 1 || ver.Minor != 2 || ver.Patch != 0 {
		t.Errorf("Blank string didn't parse to version 1.2.0, instead was %v", ver)
	}
}

func TestThreeDigits(t *testing.T) {
	ver := parseVersion("1.2.3")

	if ver.Major != 1 || ver.Minor != 2 || ver.Patch != 3 {
		t.Errorf("Blank string didn't parse to version 1.2.3, instead was %v", ver)
	}
}

func makeVersions(ver, testVer string) (Version, Version) {
	return parseVersion(ver), parseVersion(testVer)
}

func TestIsEmptyStringOlder(t *testing.T) {
	ver, testVer := makeVersions("", "1.2.3")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsSameMajorNewer(t *testing.T) {
	ver, testVer := makeVersions("1", "1")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsSameMajorMinorNewer(t *testing.T) {
	ver, testVer := makeVersions("1.2", "1.2")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsSameMajorMinorPatchNewer(t *testing.T) {
	ver, testVer := makeVersions("1.2.3", "1.2.3")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsOlderMajorOlder(t *testing.T) {
	ver, testVer := makeVersions("1", "2")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsOlderMajorMinorOlder(t *testing.T) {
	ver, testVer := makeVersions("1.1", "1.2")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsOlderMajorMinorPatchOlder(t *testing.T) {
	ver, testVer := makeVersions("1.2.2", "1.2.3")

	if ver.isNewerThan(testVer) {
		t.Errorf("%v is newer than %v", ver, testVer)
	}
}

func TestIsOlderVersion(t *testing.T) {
	SetVersion("1.2.0")
	if IsOlderVersion(AboutKismatic.ShortVersion) {
		t.Error("IsOlder returned true for the same version")
	}
}
