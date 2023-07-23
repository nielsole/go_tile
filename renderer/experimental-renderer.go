//go:build !windows
// +build !windows

package renderer

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"git.sr.ht/~sbinet/gg"
	"github.com/nielsole/go_tile/utils"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func pointToPixels(point Point, bbox BoundingBox, pixels uint32) Pixel {
	x := (point.Lon - bbox.Min.Lon) / (bbox.Max.Lon - bbox.Min.Lon) * float64(pixels)
	y := float64(pixels) - (point.Lat-bbox.Min.Lat)/(bbox.Max.Lat-bbox.Min.Lat)*float64(pixels)
	return Pixel{x, y}
}

// Returns true for types of Ways that should be displayed at zoom level 0-9.
// Includes major roads, railways, waterways, and important landmarks.
func importantLandmarkZ9(tags *osm.Tags) bool {
	return isImportantWay(tags) //|| isRailWay(tags) || isWaterWay(tags) || isImportantLandmark(tags)
}

func isImportantWay(tags *osm.Tags) bool {
	value := tags.Find("highway")
	return value == "motorway" || value == "trunk" || value == "primary" || value == "secondary" || value == "tertiary" || value == "motorway_link" || value == "trunk_link" || value == "primary_link" || value == "secondary_link" || value == "tertiary_link"
}

func nodeToPoint(node *osm.WayNode) Point {
	return Point{node.Lon, node.Lat}
}

func LoadData(path string, maxZ uint32, tempFile *os.File) (*Data, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// The third parameter is the number of parallel decoders to use.
	scanner := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	defer scanner.Close()
	data := Data{
		Tiles:    make(map[uint64]*[]MapObjectOffset),
		Filepath: tempFile.Name(),
	}

	// // Creating an empty list for every layer of zoom 0-20
	// for i := 0; i < 21; i++ {
	// 	data.Levels = append(data.Levels, make([]uint64, 0))
	// }
	var i uint64 = 0
	var maxPoints int = 0
	for scanner.Scan() {
		switch v := scanner.Object().(type) {
		case *osm.Way:
			mapObject := MapObject{
				BoundingBox: getBoundingBoxFromWay(v),
				Points:      make([]Point, 0),
			}
			for _, node := range v.Nodes {
				mapObject.Points = append(mapObject.Points, nodeToPoint(&node))
			}
			if len(mapObject.Points) > maxPoints {
				maxPoints = len(mapObject.Points)
			}
			position, err := tempFile.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
			_, err = WriteMapObject(tempFile, mapObject)
			if err != nil {
				return nil, err
			}
			containingTiles := getTilesForBoundingBox(mapObject.BoundingBox, 0, maxZ)
			for _, containingTile := range containingTiles {
				if !importantLandmarkZ9(&v.Tags) && (containingTile.Z < 11) {
					continue
				}
				index := containingTile.index()
				if _, ok := data.Tiles[index]; ok {
					*data.Tiles[index] = append(*data.Tiles[index], MapObjectOffset(position))
				} else {
					data.Tiles[index] = &[]MapObjectOffset{MapObjectOffset(position)}
				}
			}
			i += 1
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	data.MaxPoints = maxPoints
	return &data, nil
}

func Mmap(path string) (*[]byte, *os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	size := int64(fi.Size())
	// Get system page size
	pageSize := os.Getpagesize()

	// Calculate the number of pages needed, rounding up
	pages := (size + int64(pageSize) - 1) / int64(pageSize)

	// Calculate the size in bytes
	sizeInBytes := pages * int64(pageSize)

	// Memory-map the file
	mmapData, err := syscall.Mmap(int(file.Fd()), 0, int(sizeInBytes), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, nil, err
	}
	return &mmapData, file, nil
}

func HandleRenderRequest(w http.ResponseWriter, r *http.Request, duration time.Duration, data *Data, maxTreeDepth uint32, mmapData *[]byte) {
	z, x, y, ext, err := utils.ParsePath(r.URL.Path)
	if ext != "png" {
		http.Error(w, "Only png is supported", http.StatusBadRequest)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tile := Tile{
		X: x,
		Y: y,
		Z: z,
	}
	bbox := getBoundingBox(tile)
	// fmt.Printf("Bounding box is from Lat %f, Lon %f to Lat %f, Lon %f\n", bbox.Min.Lat, bbox.Min.Lon, bbox.Max.Lat, bbox.Max.Lon)

	const S = 256
	dc := gg.NewContext(S, S)

	dc.SetRGB(1, 1, 1)
	dc.Clear()
	parentTile := tile
	for tempZ := z; tempZ > maxTreeDepth; tempZ-- {
		parentTile = parentTile.getParent()
	}
	wayIndices, ok := data.Tiles[parentTile.index()]
	if !ok {
		// Return 404
		w.WriteHeader(http.StatusNotFound)
		return
	}
	way := MapObject{Points: make([]Point, 0, data.MaxPoints)}
	for _, wayReference := range *wayIndices {
		err := ReadMapObject(mmapData, int64(wayReference), &way)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// way := data.MapObjects[wayReference]
		visible := false
		if !bbox.overlaps(way.BoundingBox) {
			continue
		}
		//tags := way.Tags.Map()
		for i, point := range way.Points {
			// Print Lat and Lng of node
			//fmt.Printf("Lat: %f, Lon: %f\n", node.Lat, node.Lon)

			if bbox.contains(point) {
				visible = true
			}

			// We know that all previous points were outside of bounds, so we can skip them
			if i != 0 && visible {
				dc.SetRGB(0, 0, 0)
				// if highway, ok := tags["highway"]; ok {
				// 	switch highway {
				// 	case "motorway":
				// 		dc.SetRGB(0.9, 0.6, 0.6)
				// 		dc.SetLineWidth(6)
				// 	case "trunk":
				// 		dc.SetRGB(0.85, 0.55, 0.55)
				// 		dc.SetLineWidth(4)
				// 	case "primary":
				// 		dc.SetRGB(0.8, 0.5, 0.5)
				// 		dc.SetLineWidth(3)
				// 	case "secondary":
				// 		dc.SetRGB(0.75, 0.45, 0.45)
				// 		dc.SetLineWidth(2)
				// 	case "tertiary":
				// 		dc.SetRGB(0.7, 0.4, 0.4)
				// 		dc.SetLineWidth(1.5)
				// 	case "unclassified":
				// 		dc.SetRGB(0.65, 0.35, 0.35)
				// 		dc.SetLineWidth(1.5)
				// 	case "residential":
				// 		dc.SetRGB(0.6, 0.3, 0.3)
				// 		dc.SetLineWidth(1.5)
				// 	}
				// } else {
				dc.SetLineWidth(1)
				//}

				previousPoint := way.Points[i-1]
				currentPixel := pointToPixels(point, bbox, S)
				previousPixel := pointToPixels(previousPoint, bbox, S)
				dc.DrawLine(previousPixel.X, previousPixel.Y, currentPixel.X, currentPixel.Y)
				dc.Stroke()
			}
		}
	}

	w.Header().Set("Content-Type", "image/png")
	if err := dc.EncodePNG(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
