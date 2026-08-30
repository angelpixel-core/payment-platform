package domain

import (
	"encoding/json"
	"testing"
)

func TestAmountJSON(t *testing.T) {
	value, err := NewAmount(123)
	if err != nil {
		t.Fatalf("new amount failed: %v", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) != "123" {
		t.Fatalf("expected 123, got %s", data)
	}

	var decoded Amount
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded != value {
		t.Fatalf("expected %v, got %v", value, decoded)
	}
}

func TestCurrencyJSON(t *testing.T) {
	value, err := NewCurrency("usd")
	if err != nil {
		t.Fatalf("new currency failed: %v", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) != `"USD"` {
		t.Fatalf("expected \"USD\", got %s", data)
	}

	var decoded Currency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded != value {
		t.Fatalf("expected %v, got %v", value, decoded)
	}
}
