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
		packageName                string
		installationDisabled       bool
		dockerInstallationDisabled bool
		isInstalled                bool
		isAvailable                bool

		expected    bool
		errExpected bool
	}{
		{
			packageName:          "somePkg",
			installationDisabled: true,
			isInstalled:          true,
			isAvailable:          true,
			expected:             true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: true,
			isInstalled:          false,
			isAvailable:          true,
			expected:             false,
			errExpected:          true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: true,
			isInstalled:          false,
			isAvailable:          false,
			expected:             false,
			errExpected:          true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: false,
			isInstalled:          true,
			isAvailable:          true,
			expected:             true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: false,
			isInstalled:          false,
			isAvailable:          true,
			expected:             true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: false,
			isInstalled:          true,
			isAvailable:          false,
			expected:             true,
		},
		{
			packageName:          "somePkg",
			installationDisabled: false,
			isInstalled:          false,
			isAvailable:          false,
			expected:             true,
		},
		{
			packageName:                "somePkg",
			installationDisabled:       true,
			dockerInstallationDisabled: true,
			isInstalled:                false,
			isAvailable:                false,
			expected:                   false,
			errExpected:                true,
		},
		{
			packageName:                "docker-ce",
			installationDisabled:       true,
			dockerInstallationDisabled: true,
			isInstalled:                false,
			isAvailable:                false,
			expected:                   true,
		},
	}

	for i, test := range tests {
		c := PackageCheck{
			PackageQuery:               PackageQuery{test.packageName, "someVersion"},
			PackageManager:             stubPkgManager{installed: test.isInstalled, available: test.isAvailable},
			InstallationDisabled:       test.installationDisabled,
			DockerInstallationDisabled: test.dockerInstallationDisabled,
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
