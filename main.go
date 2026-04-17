package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Luhn logic
// ---------------------------------------------------------------------------

// luhnChecksum computes the Luhn checksum digit for a sequence of digits
// (without a check digit appended yet).
//
// Algorithm:
//  1. Start from the rightmost digit of the payload and double every second digit.
//  2. If doubling produces a value > 9, subtract 9.
//  3. Sum all digits.
//  4. The check digit is (10 - (sum % 10)) % 10.
func luhnChecksum(digits []int) int {
	sum := 0
	// We treat the payload as if it will have the check digit appended at the
	// right, so the rightmost payload digit is in an "odd" position (1-indexed
	// from the right of the final number, starting at position 2).
	double := true // rightmost payload digit gets doubled first
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return (10 - (sum % 10)) % 10
}

// parseDigits converts a string of digits into a []int, ignoring spaces and
// hyphens (common separators in card numbers).
func parseDigits(s string) ([]int, error) {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	if s == "" {
		return nil, fmt.Errorf("empty input")
	}
	digits := make([]int, len(s))
	for i, ch := range s {
		d, err := strconv.Atoi(string(ch))
		if err != nil {
			return nil, fmt.Errorf("non-digit character %q at position %d", ch, i)
		}
		digits[i] = d
	}
	return digits, nil
}

// MakeLuhnValid appends the correct Luhn check digit to the input number
// and returns the resulting string.
func MakeLuhnValid(input string) (string, error) {
	digits, err := parseDigits(input)
	if err != nil {
		return "", err
	}
	check := luhnChecksum(digits)
	return input + strconv.Itoa(check), nil
}

// IsLuhnValid returns true when the full number (including its last digit as
// the check digit) passes the Luhn algorithm.
func IsLuhnValid(input string) (bool, error) {
	digits, err := parseDigits(input)
	if err != nil {
		return false, err
	}
	if len(digits) < 2 {
		return false, fmt.Errorf("number too short to validate")
	}
	// Validate: checksum over all digits including the last one must be 0.
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0, nil
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

type transformRequest struct {
	Number string `json:"number"`
}

type transformResponse struct {
	Input   string `json:"input"`
	Result  string `json:"result"`
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// POST /luhn/transform
//
// Accepts JSON body: {"number": "7992739871"}
// Returns the number with a Luhn check digit appended.
func handleTransform(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "only POST is allowed"})
		return
	}

	var req transformRequest

	// Support both JSON body and application/x-www-form-urlencoded
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON: " + err.Error()})
			return
		}
	} else {
		// Fall back to form value or raw query param
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "could not parse form: " + err.Error()})
			return
		}
		req.Number = r.FormValue("number")
	}

	if req.Number == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: `field "number" is required`})
		return
	}

	result, err := MakeLuhnValid(req.Number)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, transformResponse{
		Input:   req.Number,
		Result:  result,
		Valid:   true,
		Message: "check digit appended — number now passes Luhn validation",
	})
}

// POST /luhn/validate
//
// Validates whether a full number (including its existing check digit) passes
// the Luhn algorithm. Useful for testing the transform output.
func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "only POST is allowed"})
		return
	}

	var req transformRequest

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON: " + err.Error()})
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "could not parse form: " + err.Error()})
			return
		}
		req.Number = r.FormValue("number")
	}

	if req.Number == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: `field "number" is required`})
		return
	}

	ok, err := IsLuhnValid(req.Number)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}

	msg := "number passes Luhn check"
	if !ok {
		msg = "number fails Luhn check"
	}
	writeJSON(w, http.StatusOK, transformResponse{
		Input:   req.Number,
		Result:  req.Number,
		Valid:   ok,
		Message: msg,
	})
}

// GET /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/luhn/transform", handleTransform)
	mux.HandleFunc("/luhn/validate", handleValidate)
	mux.HandleFunc("/health", handleHealth)

	addr := ":" + port
	log.Printf("luhn-service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
