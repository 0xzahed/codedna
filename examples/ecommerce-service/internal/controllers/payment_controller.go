package controllers

import "fmt"

func ProcessPaymentHandler(amount float64, currency string) (string, error) {
	fmt.Printf("Handling payment request for %f %s\n", amount, currency)
	return "tx_12345", nil
}

func RefundPaymentHandler(txID string) error {
	fmt.Printf("Refunding tx %s\n", txID)
	return nil
}
