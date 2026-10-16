package cli

import "testing"

func TestSpaced(t *testing.T) {
	for _, tt := range []struct {
		parts []string
		want  string
	}{
		{parts: []string{"3.50", "EUR"}, want: "3.50 EUR"},
		{parts: []string{"3.50", " EUR "}, want: "3.50 EUR"},
		{parts: []string{"", " kg "}, want: "kg"},
		{parts: []string{"Ada", "", "Example"}, want: "Ada Example"},
		{parts: []string{" ", ""}, want: ""},
	} {
		if got := spaced(tt.parts...); got != tt.want {
			t.Errorf("spaced(%q) = %q; want %q", tt.parts, got, tt.want)
		}
	}
}

// TestTrimAmount checks that tables drop the trailing zeros of an amount
// beyond its cents only, as Emporix sends some amounts with 40 decimals.
func TestTrimAmount(t *testing.T) {
	for amount, want := range map[string]string{
		"35.7000000000000000000000000000000000000000": "35.70",
		"30.00000000000000000000":                     "30.00",
		"3.50":                                        "3.50",
		"35.7":                                        "35.7",
		"10.0":                                        "10.0",
		"1.2345":                                      "1.2345",
		"1.234500":                                    "1.2345",
		"-0.5000":                                     "-0.50",
		"100":                                         "100",
		"1.000e2":                                     "1.000e2",
	} {
		if got := trimAmount(amount); got != want {
			t.Errorf("trimAmount(%q) = %q; want %q", amount, got, want)
		}
	}
}
