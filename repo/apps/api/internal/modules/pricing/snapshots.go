package pricing

import (
	"encoding/json"
	"fmt"
)

type PricingSnapshot struct {
	Items  []BookingItem  `json:"items"`
	Result *PricingResult `json:"result"`
}

func CreateSnapshot(result *PricingResult, items []BookingItem) ([]byte, error) {
	snapshot := PricingSnapshot{
		Items:  items,
		Result: result,
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal pricing snapshot: %w", err)
	}
	return data, nil
}
