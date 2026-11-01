package controllers

import "testing"

func TestProcessPaymentHandler(t *testing.T) {
	tx, err := ProcessPaymentHandler(100.0, "USD")
	if err != nil || tx == "" {
		t.Fatalf("Payment handler failed: %v", err)
	}
}
