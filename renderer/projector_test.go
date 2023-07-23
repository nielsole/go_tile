package renderer

import "testing"

func TestNum2Deg(t *testing.T) {
	// The center of Hamburg
	tile := Tile{
		X: 138346,
		Y: 84715,
		Z: 18,
	}
	bbox := getBoundingBox(tile)
	test_point := bbox.center()
	x, y := deg2num(test_point.Lat, test_point.Lon, 18)
	if x != tile.X {
		t.Errorf("X should be %v, but is %v", tile.X, x)
	}
	if y != tile.Y {
		t.Errorf("Y should be %v, but is %v", tile.Y, y)
	}
}

func TestGetTilesForBoundingBox(t *testing.T) {
	bbox := BoundingBox{
		Min: Point{
			Lat: 53.557078,
			Lon: 9.989095,
		},
		Max: Point{
			Lat: 53.557078,
			Lon: 9.989095,
		},
	}
	tiles := getTilesForBoundingBox(bbox, 18, 18)
	if len(tiles) != 1 {
		t.Errorf("Should be 1 tile, but is %v", len(tiles))
	}
	if tiles[0].X != 138345 {
		t.Errorf("X should be 138346, but is %v", tiles[0].X)
	}
	if tiles[0].Y != 84715 {
		t.Errorf("Y should be 84715, but is %v", tiles[0].Y)
	}
	if tiles[0].Z != 18 {
		t.Errorf("Z should be 18, but is %v", tiles[0].Z)
	}
	tiles = getTilesForBoundingBox(bbox, 1, 17)
	if len(tiles) != 17 {
		t.Errorf("Should be 17 tiles, but is %v", len(tiles))
	}
}
