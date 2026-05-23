package renderpdf

import "testing"

func TestQuadToCubic(t *testing.T) {
	c1, c2 := quadToCubic(point{0, 0}, point{3, 3}, point{6, 0})
	if c1.X != 2 || c1.Y != 2 || c2.X != 4 || c2.Y != 2 {
		t.Fatalf("unexpected cubic controls: %+v %+v", c1, c2)
	}
}
