package riptracer

import (
	"fmt"
	"testing"
)

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestParseNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		err      error
	}{
		{"1 2 3 4 5", []int{1, 2, 3, 4, 5}, nil},
		{"1 2 3 4 5 6", []int{1, 2, 3, 4, 5, 6}, nil},
		{"1 2 3 a 4 5", nil, fmt.Errorf("Error parsing number: strconv.Atoi: parsing \"a\": invalid syntax")},
		{" 1 2 3   4   5  ", []int{1, 2, 3, 4, 5}, nil},
		{"", []int{}, nil},
	}
	for _, test := range tests {
		actual, err := parseNumbers(test.input)
		if err != nil && test.err != nil && err.Error() != test.err.Error() {
			t.Errorf("For input %q, expected error %v but got %v", test.input, test.err, err)
		}
		if err == nil && !equal(actual, test.expected) {
			t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, actual)
		}
	}
}
