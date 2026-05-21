package cmd

import (
	"fmt"
	"strconv"
)

// parseID converts a string to a positive uint, returning an error for invalid input.
func parseID(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("invalid id %q: must be a positive integer", s)
	}
	return uint(n), nil
}

// parseStatus converts a string to a product status code (0, 1, or 2).
func parseStatus(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 2 {
		return 0, fmt.Errorf("invalid status %q: must be 0 (available), 1 (sold), or 2 (removed)", s)
	}
	return n, nil
}
