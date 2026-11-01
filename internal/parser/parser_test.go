package parser

import (
	"testing"
)

func TestQueriesCompile(t *testing.T) {
	sp, err := getSpecs()
	if err != nil {
		t.Fatalf("specs: %v", err)
	}
	if _, err := getCompiled(); err != nil {
		t.Fatalf("getCompiled: %v", err)
	}
	if len(sp) != 7 { // 6 languages + tsx variant
		t.Errorf("expected 7 specs, got %d", len(sp))
	}
}

func TestParseGo(t *testing.T) {
	src := []byte(`package payments

import (
	"fmt"
STDATABASE "database/sql"
)

// Process handles payment processing.
func Process(id string) error {
	fmt.Println("processing", id)
	return save(id)
}

type Payment struct{ ID string }

func (p Payment) Save() error { return save(p.ID) }

func save(id string) error { return nil }
`)
	spec := specForLang("go", ".go", t)
	res, err := ParseSource("payments/process.go", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var names []string
	for _, s := range res.Symbols {
		names = append(names, string(s.Kind)+":"+s.QualifiedName)
	}
	want := []string{"function:Process", "struct:Payment", "method:Payment.Save", "function:save"}
	if len(names) != len(want) {
		t.Fatalf("symbols: got %v want %v", names, want)
	}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("symbol %d: got %q want %q", i, names[i], w)
		}
	}
	if len(res.Imports) != 2 {
		t.Fatalf("imports: %v", res.Imports)
	}
	if len(res.Calls) < 3 {
		t.Fatalf("calls: %+v", res.Calls)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestParsePython(t *testing.T) {
	src := []byte(`import json
from orders.models import Order

def create_order(user):
    order = Order(user=user)
    json.dumps(order)
    return order

class Cart:
    def add(self, item):
        return item
`)
	spec := specForLang("python", ".py", t)
	res, err := ParseSource("orders/cart.py", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := len(res.Symbols); got != 3 {
		t.Fatalf("symbols: got %d: %+v", got, res.Symbols)
	}
	if len(res.Imports) < 2 {
		t.Fatalf("imports: %+v", res.Imports)
	}
	calls := 0
	for _, c := range res.Calls {
		if c.Callee == "" {
			t.Fatalf("empty callee in calls: %+v", c)
		}
		calls++
	}
	if calls < 2 {
		t.Fatalf("calls: %+v", res.Calls)
	}
}

func TestParseTypescript(t *testing.T) {
	src := []byte(`import { Payment } from './payment'

export class Invoice {
    constructor(p: Payment) {
        this.p = p
    }
    render(): string {
        return this.p.toString()
    }
    private p: Payment
}

export function total(items: number[]): number {
    return items.reduce((a, b) => a + b, 0)
}`)
	spec := specForLang("typescript", ".ts", t)
	res, err := ParseSource("invoice.ts", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	kinds := map[string]int{}
	for _, s := range res.Symbols {
		kinds[string(s.Kind)]++
	}
	if kinds["class"] != 1 || kinds["method"] != 2 || kinds["function"] != 1 {
		t.Fatalf("kinds: %+v symbols %+v", kinds, res.Symbols)
	}
	exported := []bool{}
	for _, s := range res.Symbols {
		exported = append(exported, s.Exported)
	}
	// Invoice, Invoice.render, Invoice.p(method name), total
	if len(exported) != 4 {
		t.Fatalf("want 4 symbols got %d", len(exported))
	}
}

func TestParseJava(t *testing.T) {
	src := []byte(`package com.acme.orders;

import com.acme.events.OrderCreated;

public class Order {
    public void create(OrderCreated event) {
        Invoice invoice = new Invoice();
        invoice.render();
    }
}`)
	spec := specForLang("java", ".java", t)
	res, err := ParseSource("com/acme/Order.java", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Symbols) != 2 {
		t.Fatalf("symbols: %+v", res.Symbols)
	}
	if len(res.Imports) != 1 {
		t.Fatalf("imports: %+v", res.Imports)
	}
	if len(res.Calls) < 2 {
		t.Fatalf("calls: %+v", res.Calls)
	}
}

func TestParsePHP(t *testing.T) {
	src := []byte(`<?php
use App\\Services\\Payment as PaymentService;

class Checkout {
    public function pay($amount) {
        return PaymentService::charge($amount);
    }
}
`)
	spec := specForLang("php", ".php", t)
	res, err := ParseSource("app/Checkout.php", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Symbols) != 2 {
		t.Fatalf("symbols: %+v", res.Symbols)
	}
	if len(res.Imports) < 1 {
		t.Fatalf("imports: %+v", res.Imports)
	}
	if len(res.Calls) == 0 {
		t.Fatal("no calls")
	}
}

func TestParseJS(t *testing.T) {
	src := []byte(`import express from 'express'
const app = express()

export function handler(req, res) {
  app.get('/ping', heathHandler)
}

const heathHandler = (req, res) => { res.send('ok') }`)
	spec := specForLang("javascript", ".js", t)
	res, err := ParseSource("server.js", spec, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Symbols) != 2 {
		t.Fatalf("symbols: %+v", res.Symbols)
	}
	if len(res.Imports) != 1 {
		t.Fatalf("imports: %+v", res.Imports)
	}
}

func specForLang(lang, ext string, t *testing.T) *LanguageSpec {
	t.Helper()
	sp, err := getSpecs()
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := sp[grammarKey(lang, ext)]
	if !ok {
		t.Fatalf("no spec for %s/%s", lang, ext)
	}
	return spec
}
