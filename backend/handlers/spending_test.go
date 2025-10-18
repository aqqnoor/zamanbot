package handlers

import (
	"strings"
	"testing"
)

func TestReadCSVInto(t *testing.T) {
	csv := "-2500,food Market
	-800,taxi Yandex"
	n, err := readCSVInto("test", strings.NewReader(csv), ',')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("want 2, got %d", n)
	}
}
