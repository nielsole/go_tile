package renderer

import "testing"

func TestTileGetParent(t *testing.T) {
	tile := Tile{
		X: 276643,
		Y: 169357,
		Z: 19,
	}
	parent := tile.getParent()
	if parent.X != 138321 {
		t.Errorf("X should be 138321, but is %v", parent.X)
	}
	if parent.Y != 84678 {
		t.Errorf("Y should be 84678, but is %v", parent.Y)
	}
	if parent.Z != 18 {
		t.Errorf("Z should be 18, but is %v", parent.Z)
	}
}
