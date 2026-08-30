package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Amount int64

func NewAmount(value int64) (Amount, error) {
	if value <= 0 {
		return 0, NewError(400, "invalid_amount", "amount must be greater than zero")
	}
	return Amount(value), nil
}

func (a Amount) Int64() int64 { return int64(a) }

func (a Amount) MarshalJSON() ([]byte, error) { return []byte(strconv.FormatInt(int64(a), 10)), nil }

func (a *Amount) UnmarshalJSON(data []byte) error {
	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("amount must be a number: %w", err)
	}
	*a = Amount(value)
	return nil
}

type Currency string

func NewCurrency(value string) (Currency, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", NewError(400, "invalid_currency", "currency is required")
	}
	return Currency(value), nil
}

func (c Currency) String() string { return string(c) }

func (c Currency) MarshalJSON() ([]byte, error) { return json.Marshal(string(c)) }

func (c *Currency) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("currency must be a string: %w", err)
	}
	normalized, err := NewCurrency(value)
	if err != nil {
		return err
	}
	*c = normalized
	return nil
}
