package renderer

import (
	"math"

	"github.com/paulmach/osm"
)

func getBoundingBox(tile Tile) BoundingBox {
	n := math.Pow(2.0, float64(tile.Z))
	lonMin := float64(tile.X)/n*360.0 - 180.0
	latMin := math.Atan(math.Sinh(math.Pi*(1-2*float64(tile.Y)/n))) * 180.0 / math.Pi
	lonMax := float64(tile.X+1)/n*360.0 - 180.0
	latMax := math.Atan(math.Sinh(math.Pi*(1-2*float64(tile.Y+1)/n))) * 180.0 / math.Pi
	pointMin := Point{math.Min(lonMin, lonMax), math.Min(latMin, latMax)}
	pointMax := Point{math.Max(lonMin, lonMax), math.Max(latMin, latMax)}
	return BoundingBox{pointMin, pointMax}
}

func getBoundingBoxFromWay(way *osm.Way) BoundingBox {
	var lonMin float64 = 200.0
	var latMin float64 = 200.0
	var lonMax float64 = -200.0
	var latMax float64 = -200.0
	for _, node := range way.Nodes {
		lonMin = math.Min(lonMin, node.Lon)
		latMin = math.Min(latMin, node.Lat)
		lonMax = math.Max(lonMax, node.Lon)
		latMax = math.Max(latMax, node.Lat)
	}
	pointMin := Point{lonMin, latMin}
	pointMax := Point{lonMax, latMax}
	return BoundingBox{pointMin, pointMax}

}

// Given a latitude, longitude and zoom level, return the tile coordinates
func deg2num(lat_deg, lon_deg float64, zoom uint32) (x, y uint32) {
	lat_rad := math.Pi * lat_deg / 180.0
	n := math.Pow(2.0, float64(zoom))
	x = uint32(math.Floor((lon_deg + 180.0) / 360.0 * n))
	y = uint32(math.Floor((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * n))
	return
}

func getTilesForBoundingBox(bbox BoundingBox, minZ, maxZ uint32) []Tile {
	var tiles []Tile

	for z := minZ; z <= maxZ; z++ {
		minX, minY := deg2num(bbox.Max.Lat, bbox.Min.Lon, z)
		maxX, maxY := deg2num(bbox.Min.Lat, bbox.Max.Lon, z)

		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				tiles = append(tiles, Tile{X: x, Y: y, Z: z})
			}
		}
	}

	return tiles
}
