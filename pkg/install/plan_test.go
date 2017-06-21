package install

import "testing"

func TestGenerateAlphaNumericPassword(t *testing.T) {
	_, err := generateAlphaNumericPassword()
	if err != nil {
		t.Error(err)
	}
}

func TestReadWithDeprecated(t *testing.T) {
	pm := &DeprecatedPackageManager{
		Enabled: true,
	}
	p := &Plan{}
	p.Features = &Features{
		PackageManager: pm,
	}
	b := false
	p.Cluster.AllowPackageInstallation = &b
	readDeprecatedFields(p)

	// features.package_manager should be set to add_ons.package_manager
	if p.AddOns.PackageManager.Disabled || p.AddOns.PackageManager.Provider != "helm" {
		t.Errorf("Expected add_ons.package_manager to be read from features.package_manager")
	}
	// cluster.disable_package_installation shoule be set to cluster.allow_package_installation
	if p.Cluster.DisablePackageInstallation != true {
		t.Errorf("Expected cluster.allow_package_installation to be read from cluster.disable_package_installation")
	}
}
