package install

import "testing"

func TestGenerateAlphaNumericPassword(t *testing.T) {
	_, err := generateAlphaNumericPassword()
	if err != nil {
		t.Error(err)
	}
}

func TestReadWtihDeprecated(t *testing.T) {
	pm := &PackageManager{
		Enabled:  true,
		Provider: "helm",
	}
	p := &Plan{}
	p.Features.PackageManager = pm
	readDeprecatedFields(p)

	// add_ons.package_manager should be set to features.package_manager
	if !p.AddOns.PackageManager.Enabled || p.AddOns.PackageManager.Provider != "helm" {
		t.Errorf("Expected add_ons.package_manager to be read from features.package_manager")
	}
}
