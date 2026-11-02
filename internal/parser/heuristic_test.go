package parser

import (
	"testing"

	"github.com/codedna/codedna/internal/model"
)

func TestParseHeuristic(t *testing.T) {
	// 1. Rust code
	rustSrc := `
use std::collections::HashMap;
use crate::engine::Parser;

pub struct ArchitectureEngine {
    version: String,
}

pub async fn execute_pipeline(name: &str) -> bool {
    true
}
`
	res, err := ParseHeuristic("src/engine.rs", "rust", []byte(rustSrc))
	if err != nil {
		t.Fatalf("ParseHeuristic rust failed: %v", err)
	}
	if len(res.Imports) != 2 {
		t.Errorf("Expected 2 rust imports, got %d", len(res.Imports))
	}
	if len(res.Symbols) != 2 {
		t.Errorf("Expected 2 rust symbols, got %d: %+v", len(res.Symbols), res.Symbols)
	}

	// 2. C++ code
	cppSrc := `
#include <iostream>
#include "architecture.hpp"

class SystemProcessor {
    void processEvent() {
    }
};
`
	res, err = ParseHeuristic("include/processor.cpp", "cpp", []byte(cppSrc))
	if err != nil {
		t.Fatalf("ParseHeuristic cpp failed: %v", err)
	}
	if len(res.Imports) != 2 {
		t.Errorf("Expected 2 cpp includes, got %d", len(res.Imports))
	}
	foundClass := false
	for _, sym := range res.Symbols {
		if sym.Name == "SystemProcessor" && sym.Kind == model.SymbolClass {
			foundClass = true
		}
	}
	if !foundClass {
		t.Errorf("Expected SystemProcessor class in cpp, got %+v", res.Symbols)
	}

	// 3. Ruby code
	rubySrc := `
require 'json'
require_relative 'config'

module PaymentGateway
  class StripeClient
    def process_charge(amount)
      true
    end
  end
end
`
	res, err = ParseHeuristic("lib/stripe.rb", "ruby", []byte(rubySrc))
	if err != nil {
		t.Fatalf("ParseHeuristic ruby failed: %v", err)
	}
	if len(res.Imports) != 2 {
		t.Errorf("Expected 2 ruby requires, got %d", len(res.Imports))
	}
	foundDef := false
	for _, sym := range res.Symbols {
		if sym.Name == "process_charge" && sym.Kind == model.SymbolFunction {
			foundDef = true
		}
	}
	if !foundDef {
		t.Errorf("Expected process_charge def in ruby, got %+v", res.Symbols)
	}
}
