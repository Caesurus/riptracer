package riptracer

import "testing"

func TestParseELF64RelaEntry(t *testing.T) {
	// Test data representing an ELF64_Rela struct
	data := []byte{200, 246, 148, 1, 0, 0, 0, 0, 7, 0, 0, 0, 113, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// Expected values of the ELF64_Rela struct
	expected := ELF64_Rela{R_offset: 0x194f6c8,
		R_info:   ELF64_Rela_Info{Type: 7, Sym: 1393},
		R_addend: 0x0}

	// Call the parseELF64RelaEntry function with the test data
	result, err := parseELF64RelaEntry(data)
	if err != nil {
		t.Fatalf("parseELF64RelaEntry failed: %s", err)
	}

	// Compare the result to the expected values
	if result != expected {
		t.Fatalf("parseELF64RelaEntry returned incorrect result.\nExpected: %+v\nGot: %+v\n",
			expected, result)
	}
}
