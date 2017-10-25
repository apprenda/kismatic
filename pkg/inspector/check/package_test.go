package check

import "testing"

type stubPkgManager struct {
	installed bool
	available bool
}

func (m stubPkgManager) IsInstalled(q PackageQuery) (bool, error) {
	return m.installed, nil
}

func (m stubPkgManager) IsAvailable(q PackageQuery) (bool, error) {
	return m.available, nil
}

func TestPackageCheck(t *testing.T) {
	tests := []struct {
		installationDisabled bool
		isInstalled          bool
		isAvailable          bool
		shouldNotBeInstalled bool
		expected             bool
		errExpected          bool
	}{
		{
			installationDisabled: true,
			isInstalled:          true,
			isAvailable:          true,
			expected:             true,
		},
		{
			installationDisabled: true,
			isInstalled:          false,
			isAvailable:          true,
			expected:             false,
			errExpected:          true,
		},
		{
			installationDisabled: true,
			isInstalled:          false,
			isAvailable:          false,
			expected:             false,
			errExpected:          true,
		},
		{
			installationDisabled: false,
			isInstalled:          true,
			isAvailable:          true,
			expected:             true,
		},
		{
			installationDisabled: false,
			isInstalled:          false,
			isAvailable:          true,
			expected:             true,
		},
		{
			installationDisabled: false,
			isInstalled:          true,
			isAvailable:          false,
			expected:             true,
		},
		{
			installationDisabled: false,
			isInstalled:          false,
			isAvailable:          false,
			expected:             true,
		},
		{
			installationDisabled: false,
			isInstalled:          true,
			isAvailable:          true,
			shouldNotBeInstalled: true,
			expected:             false,
			errExpected:          true,
		},
		{
			installationDisabled: true,
			isInstalled:          true,
			isAvailable:          true,
			shouldNotBeInstalled: true,
			expected:             true,
			errExpected:          false,
		},
	}

	for i, test := range tests {
		c := PackageCheck{
			PackageQuery:         PackageQuery{"somePkg", "someVersion"},
			PackageManager:       stubPkgManager{installed: test.isInstalled, available: test.isAvailable},
			InstallationDisabled: test.installationDisabled,
			ShouldNotBeInstalled: test.shouldNotBeInstalled,
		}
		ok, err := c.Check()
		if err != nil && !test.errExpected {
			t.Errorf("test #%d - unexpected error occurred: %v", i, err)
		}
		if ok != test.expected {
			t.Errorf("Test #%d - Expected %v, but got %v", i, test.expected, ok)
		}
	}
}
