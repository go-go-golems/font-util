package spec

import "testing"

func TestParseLengthPt(t *testing.T) {
	tests := map[string]float64{"12pt": 12, "1in": 72, "25.4mm": 72, "2.54cm": 72, "9": 9}
	for in, want := range tests {
		got, err := ParseLengthPt(in)
		if err != nil {
			t.Fatalf("ParseLengthPt(%q): %v", in, err)
		}
		if diff := got - want; diff < -0.0001 || diff > 0.0001 {
			t.Fatalf("ParseLengthPt(%q)=%v want %v", in, got, want)
		}
	}
}

func TestStarterValid(t *testing.T) {
	s := Starter("font.otf", "out.pdf")
	if err := Validate(s); err != nil {
		t.Fatal(err)
	}
	r, err := Resolve(s)
	if err != nil {
		t.Fatal(err)
	}
	if r.PointSize <= 0 || r.Margins.Left <= 0 {
		t.Fatalf("defaults were not resolved: %+v", r)
	}
}
