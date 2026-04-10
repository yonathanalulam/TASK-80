package common

const (
	RoleAdministrator  = "administrator"
	RoleGroupOrganizer = "group_organizer"
	RoleTraveler       = "traveler"
	RoleSupplier       = "supplier"
	RoleCourierRunner  = "courier_runner"
	RoleAccountant     = "accountant"
)

func HasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func IsAdminOrAccountant(roles []string) bool {
	return HasRole(roles, RoleAdministrator) || HasRole(roles, RoleAccountant)
}

const (
	RiskActionCreateRFQ        = "create_rfq"
	RiskActionIssueRFQ         = "issue_rfq"
	RiskActionSelectSupplier   = "select_supplier"
	RiskActionCancelBooking    = "cancel_booking"
	RiskActionProcessRefund    = "process_refund"
	RiskActionRequestWithdrawal = "request_withdrawal"
)
