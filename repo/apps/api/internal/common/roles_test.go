package common

import "testing"

func TestHasRole(t *testing.T) {
	tests := []struct {
		name   string
		roles  []string
		target string
		want   bool
	}{
		{"admin in list", []string{RoleAdministrator}, RoleAdministrator, true},
		{"admin not in list", []string{RoleTraveler}, RoleAdministrator, false},
		{"empty roles", []string{}, RoleAdministrator, false},
		{"nil roles", nil, RoleAdministrator, false},
		{"multiple roles match", []string{RoleTraveler, RoleGroupOrganizer}, RoleGroupOrganizer, true},
		{"admin string is administrator not admin", []string{"admin"}, RoleAdministrator, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRole(tt.roles, tt.target)
			if got != tt.want {
				t.Errorf("HasRole(%v, %q) = %v, want %v", tt.roles, tt.target, got, tt.want)
			}
		})
	}
}

func TestIsAdminOrAccountant(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"admin", []string{RoleAdministrator}, true},
		{"accountant", []string{RoleAccountant}, true},
		{"both", []string{RoleAdministrator, RoleAccountant}, true},
		{"traveler only", []string{RoleTraveler}, false},
		{"supplier only", []string{RoleSupplier}, false},
		{"empty", []string{}, false},
		{"admin literal not matching", []string{"admin"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAdminOrAccountant(tt.roles)
			if got != tt.want {
				t.Errorf("IsAdminOrAccountant(%v) = %v, want %v", tt.roles, got, tt.want)
			}
		})
	}
}

func TestRoleConstantsAreCanonical(t *testing.T) {
	// Ensure role constants match what's in the database seeds
	expected := map[string]string{
		"RoleAdministrator":  "administrator",
		"RoleGroupOrganizer": "group_organizer",
		"RoleTraveler":       "traveler",
		"RoleSupplier":       "supplier",
		"RoleCourierRunner":  "courier_runner",
		"RoleAccountant":     "accountant",
	}
	actual := map[string]string{
		"RoleAdministrator":  RoleAdministrator,
		"RoleGroupOrganizer": RoleGroupOrganizer,
		"RoleTraveler":       RoleTraveler,
		"RoleSupplier":       RoleSupplier,
		"RoleCourierRunner":  RoleCourierRunner,
		"RoleAccountant":     RoleAccountant,
	}
	for name, want := range expected {
		if got := actual[name]; got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	// Specifically: "admin" is NOT a valid role
	if HasRole([]string{"admin"}, RoleAdministrator) {
		t.Error("'admin' should not match RoleAdministrator 'administrator'")
	}
}

func TestRiskActionConstants(t *testing.T) {
	// Verify risk action constants are defined and non-empty
	actions := []string{
		RiskActionCreateRFQ,
		RiskActionIssueRFQ,
		RiskActionSelectSupplier,
		RiskActionCancelBooking,
		RiskActionProcessRefund,
		RiskActionRequestWithdrawal,
	}
	for _, a := range actions {
		if a == "" {
			t.Error("risk action constant should not be empty")
		}
	}
}

func TestNoStrayAdminLiteral(t *testing.T) {
	// The canonical role is "administrator", not "admin".
	// This test ensures the constant is correct.
	if RoleAdministrator != "administrator" {
		t.Errorf("RoleAdministrator = %q, want %q", RoleAdministrator, "administrator")
	}
	// "admin" should never match
	if HasRole([]string{"admin"}, RoleAdministrator) {
		t.Error("'admin' should not match RoleAdministrator")
	}
	if HasRole([]string{RoleAdministrator}, "admin") {
		t.Error("RoleAdministrator should not match 'admin' literal")
	}
}
