package order

import "errors"

// ErrInvalidTransition is returned when a caller attempts an illegal status change
// (for example, Pending -> Delivered, Shipped -> Cancelled, or Seller cancelling).
var ErrInvalidTransition = errors.New("invalid status transition")

// IsValidStatus reports whether s is one of the four known order statuses.
func IsValidStatus(s string) bool {
	switch s {
	case StatusPending, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}

// CanSellerTransition enforces the seller forward-only state machine:
//
//	Pending -> Shipped -> Delivered
//
// Terminal states (Delivered, Cancelled) cannot transition anywhere, and sellers
// cannot skip steps (Pending -> Delivered) or cancel orders here.
func CanSellerTransition(from, to string) bool {
	switch from {
	case StatusPending:
		return to == StatusShipped
	case StatusShipped:
		return to == StatusDelivered
	default:
		return false
	}
}

// CanBuyerCancel reports whether a buyer is allowed to cancel an order in state from.
// Buyers may only cancel while the order is still Pending (before the seller ships).
func CanBuyerCancel(from string) bool {
	return from == StatusPending
}
