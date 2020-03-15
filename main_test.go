package main

import "testing"

func TestFormatTelegramMessage(t *testing.T) {
	ans := formatTelegramMessage("a", "b")
	right := "\n*a*``` b ```"
	if ans != right {
		t.Errorf("formatTelegramMessage = %s; want %s", ans, right)
	}
}
