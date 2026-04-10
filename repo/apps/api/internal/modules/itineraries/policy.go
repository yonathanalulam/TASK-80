package itineraries

func CanManageItinerary(userID string, roles []string, itinerary *Itinerary) bool {
	if userID == itinerary.OrganizerID {
		return true
	}
	for _, r := range roles {
		if r == "administrator" {
			return true
		}
	}
	return false
}

func CanViewItinerary(userID string, roles []string, isMember bool, itinerary *Itinerary) bool {
	if CanManageItinerary(userID, roles, itinerary) {
		return true
	}
	return isMember
}
