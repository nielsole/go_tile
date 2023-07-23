package renderer

import (
	"math"
)

type Tile struct {
	X, Y, Z uint32
}

// when all possible tiles including all smaller zoom levels are stored in a quadtree, this function will be used to get the index of a tile
func (tile Tile) index() uint64 {
	// Calculate the total number of tiles for all zoom levels from 0 to tile.Z-1
	total := uint64(0)
	for z := uint32(0); z < tile.Z; z++ {
		total += uint64(math.Pow(4, float64(z)))
	}

	// Calculate the position of the tile within its zoom level
	levelPos := tile.Y*uint32(math.Pow(2, float64(tile.Z))) + tile.X

	return total + uint64(levelPos)
}

// Returns the OSM tile one zoom level above that contains the given point
func (tile Tile) getParent() Tile {
	return Tile{tile.X / 2, tile.Y / 2, tile.Z - 1}
}

type Point struct {
	Lon, Lat float64
}

type BoundingBox struct {
	Min, Max Point
}

type Pixel struct {
	X, Y float64
}

// member function checks if a point is inside a bounding box
func (bbox BoundingBox) contains(point Point) bool {
	return point.Lat >= bbox.Min.Lat && point.Lat <= bbox.Max.Lat && point.Lon >= bbox.Min.Lon && point.Lon <= bbox.Max.Lon
}

func (bbox BoundingBox) overlaps(other BoundingBox) bool {
	return bbox.Min.Lat <= other.Max.Lat && bbox.Max.Lat >= other.Min.Lat && bbox.Min.Lon <= other.Max.Lon && bbox.Max.Lon >= other.Min.Lon
}

func (bbox BoundingBox) center() Point {
	return Point{(bbox.Min.Lon + bbox.Max.Lon) / 2, (bbox.Min.Lat + bbox.Max.Lat) / 2}
}

// Architecture:
// We have tiles. Each tile has a bounding box.
// Each tile has a list of map objects.
// Each map object has a bounding box.
// Each map object has a list of points.
// Each point has a longitude and a latitude.
// Each tile may have 4 children and 1 parent.
// Each tile has x, y, and z coordinates.
// Map objects are stored at the tile that contains their bounding box.
// As Tiles are stored on disk, we cannot reference them through pointers but by their index position.
// There is a quadtree for every zoom level.
// The quadtree stops dividing when there are 64 or less map objects in a tile.
type TileData struct {
	X, Y, Z  uint32
	Children [4]uint64
}

type MapObjectOffset uint64

type Data struct {
	// ChatGPT has the idea to use a map of tiles as a sparse array. Not sure what the performance hit of that is.
	// Then I could skip the whole fancy part of continuous memory layout and just have a map[Tile][] map.
	// TODO: replace Tile with index
	Tiles     map[uint64]*[]MapObjectOffset
	Filepath  string
	MaxPoints int
}

type MapObject struct {
	//The bounding box of the map object
	BoundingBox BoundingBox
	Points      []Point
}
