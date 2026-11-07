package order

import "testing"

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"Pending is valid", StatusPending, true},
		{"Shipped is valid", StatusShipped, true},
		{"Delivered is valid", StatusDelivered, true},
		{"Cancelled is valid", StatusCancelled, true},
		{"empty string is invalid", "", false},
		{"lowercase pending is invalid", "pending", false},
		{"arbitrary string is invalid", "Refunded", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsValidStatus(%q) = %v; want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestCanSellerTransition(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Valid transitions
		{"Pending to Shipped is allowed", StatusPending, StatusShipped, true},
		{"Shipped to Delivered is allowed", StatusShipped, StatusDelivered, true},

		// Illegal transitions
		{"Pending to Delivered (skipping Shipped) is rejected", StatusPending, StatusDelivered, false},
		{"Pending to Cancelled (seller cannot cancel) is rejected", StatusPending, StatusCancelled, false},
		{"Shipped to Pending (cannot backtrack) is rejected", StatusShipped, StatusPending, false},
		{"Shipped to Cancelled (cannot cancel shipped order) is rejected", StatusShipped, StatusCancelled, false},
		{"Delivered to anything (terminal state) is rejected", StatusDelivered, StatusPending, false},
		{"Cancelled to anything (terminal state) is rejected", StatusCancelled, StatusShipped, false},
		{"Same status transition is rejected", StatusPending, StatusPending, false},
		{"Unknown source status is rejected", "Unknown", StatusShipped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanSellerTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanSellerTransition(%q, %q) = %v; want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestCanBuyerCancel(t *testing.T) {
	tests := []struct {
		name string
		from string
		want bool
	}{
		{"Pending order can be cancelled", StatusPending, true},
		{"Shipped order cannot be cancelled", StatusShipped, false},
		{"Delivered order cannot be cancelled", StatusDelivered, false},
		{"Already cancelled order cannot be cancelled again", StatusCancelled, false},
		{"Unknown status cannot be cancelled", "Processing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanBuyerCancel(tt.from)
			if got != tt.want {
				t.Errorf("CanBuyerCancel(%q) = %v; want %v", tt.from, got, tt.want)
			}
		})
	}
}

