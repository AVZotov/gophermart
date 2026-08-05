package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr error
	}{
		{name: "valid Luhn number", number: "79927398713", wantErr: nil},
		{name: "valid single digit zero", number: "0", wantErr: nil},
		{name: "valid long number", number: "4561261212345467", wantErr: nil},
		{name: "empty string", number: "", wantErr: ErrInvalidOrderID},
		{name: "fails Luhn checksum", number: "12345678900", wantErr: ErrInvalidOrderID},
		{name: "single invalid digit", number: "9", wantErr: ErrInvalidOrderID},
		{name: "non-digit characters", number: "7992739871a", wantErr: ErrInvalidOrderID},
		{name: "whitespace", number: "799 27398713", wantErr: ErrInvalidOrderID},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				o := &Order{Number: tt.number}
				err := o.Validate()
				if tt.wantErr == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, tt.wantErr)
				}
			},
		)
	}
}

func TestNewOrder_Success(t *testing.T) {
	o, err := NewOrder("79927398713")

	require.NoError(t, err)
	assert.Equal(t, "79927398713", o.Number)
}

func TestNewOrder_InvalidNumber(t *testing.T) {
	o, err := NewOrder("12345678900")

	require.ErrorIs(t, err, ErrInvalidOrderID)
	assert.Nil(t, o)
}
