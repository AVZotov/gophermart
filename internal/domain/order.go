package domain

import "time"

// OrderStatus represents the lifecycle state of an order as it moves
// through accrual processing.
type OrderStatus string

const (
	// OrderStatusNew is the initial state of an uploaded order, not yet
	// picked up by the accrual polling worker.
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessing marks an order currently being calculated by
	// the external accrual system.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusInvalid marks an order the accrual system rejected; no
	// points are credited.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusProcessed marks an order for which accrual calculation
	// finished and points (if any) were credited to the user's balance.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// Order represents a loyalty order number uploaded by a user for accrual
// processing.
type Order struct {
	Number string
	Status OrderStatus
	// AccrualCents holds the accrual amount in minor currency units (cents)
	// to avoid float64 rounding error in loyalty-point arithmetic. Nil when
	// no accrual has been calculated yet.
	AccrualCents *int64
	UploadedAt   time.Time
}

// NewOrder constructs an Order for number, validating it via Validate. It
// returns domain.ErrInvalidOrderID if number fails Luhn validation.
func NewOrder(number string) (*Order, error) {
	o := &Order{Number: number}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return o, nil
}

// Validate checks that o.Number consists solely of digits and satisfies the
// Luhn checksum. It returns domain.ErrInvalidOrderID otherwise.
func (o *Order) Validate() error {
	if len(o.Number) == 0 {
		return ErrInvalidOrderID
	}

	sum := 0
	alt := false

	for i := len(o.Number) - 1; i >= 0; i-- {
		c := o.Number[i]
		if c < '0' || c > '9' {
			return ErrInvalidOrderID
		}
		d := int(c - '0')

		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}

	if sum%10 != 0 {
		return ErrInvalidOrderID
	}

	return nil
}
