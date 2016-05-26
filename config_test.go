package main

import (
	"testing"
)

func TestTestConfig(t *testing.T) {
	config := NewTestCharacterConfig()

	characterSet := config.CharacterSet()

	if !characterSet.Contains("Maaya Saraki", "Indy Drone 4", "Fake Character") {
		t.Fatal("Not all characters in ", characterSet.String())
	}

	expected := "agrakari"
	result := config.AlertUsername("Maaya Saraki")

	if result != expected {
		t.Fatalf("Expected %s got %s", expected, result)
	}

	expected = "agrakari"
	result = config.AlertUsername("Indy Drone 4")

	if result != expected {
		t.Fatalf("Expected %s got %s", expected, result)
	}

	expected = "fake_user"
	result = config.AlertUsername("Fake Character")

	if result != expected {
		t.Fatalf("Expected %s got %s", expected, result)
	}
}
