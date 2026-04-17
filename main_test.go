package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Unit tests — Luhn logic
// ---------------------------------------------------------------------------

func TestLuhnChecksum(t *testing.T) {
	// Classic example from Wikipedia: payload "7992739871" → check digit 3
	digits := []int{7, 9, 9, 2, 7, 3, 9, 8, 7, 1}
	got := luhnChecksum(digits)
	if got != 3 {
		t.Errorf("expected check digit 3, got %d", got)
	}
}

func TestMakeLuhnValid(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"7992739871", "79927398713"},
		{"1", "18"},
		{"0", "00"},
	}
	for _, c := range cases {
		got, err := MakeLuhnValid(c.input)
		if err != nil {
			t.Errorf("MakeLuhnValid(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got != c.expected {
			t.Errorf("MakeLuhnValid(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestIsLuhnValid(t *testing.T) {
	valid := []string{"79927398713", "4532015112830366", "18"}
	for _, n := range valid {
		ok, err := IsLuhnValid(n)
		if err != nil || !ok {
			t.Errorf("IsLuhnValid(%q) should be valid, got ok=%v err=%v", n, ok, err)
		}
	}

	invalid := []string{"79927398710", "1234567890"}
	for _, n := range invalid {
		ok, err := IsLuhnValid(n)
		if err != nil || ok {
			t.Errorf("IsLuhnValid(%q) should be invalid, got ok=%v err=%v", n, ok, err)
		}
	}
}

func TestParseDigitsIgnoresSeparators(t *testing.T) {
	digits, err := parseDigits("4532 0151 1283 0366")
	if err != nil {
		t.Fatal(err)
	}
	if len(digits) != 16 {
		t.Errorf("expected 16 digits, got %d", len(digits))
	}
}

// ---------------------------------------------------------------------------
// HTTP handler tests
// ---------------------------------------------------------------------------

func TestHandleTransformJSON(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"number": "7992739871"})
	req := httptest.NewRequest(http.MethodPost, "/luhn/transform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handleTransform(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp transformResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result != "79927398713" {
		t.Errorf("expected 79927398713, got %s", resp.Result)
	}
}

func TestHandleTransformForm(t *testing.T) {
	body := strings.NewReader("number=7992739871")
	req := httptest.NewRequest(http.MethodPost, "/luhn/transform", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handleTransform(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleTransformMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/luhn/transform", nil)
	rec := httptest.NewRecorder()
	handleTransform(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleTransformMissingNumber(t *testing.T) {
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/luhn/transform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleTransform(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleValidate(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"number": "79927398713"})
	req := httptest.NewRequest(http.MethodPost, "/luhn/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handleValidate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp transformResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Valid {
		t.Errorf("expected valid=true for a Luhn-valid number")
	}
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
