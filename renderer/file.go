package renderer

import (
	"bytes"
	"encoding/binary"
	"os"
)

func WriteMapObject(file *os.File, mo MapObject) (int64, error) {
	// Buffer to store binary representation
	buf := new(bytes.Buffer)

	// Write BoundingBox
	err := binary.Write(buf, binary.LittleEndian, mo.BoundingBox.Min.Lat)
	if err != nil {
		return 0, err
	}
	err = binary.Write(buf, binary.LittleEndian, mo.BoundingBox.Min.Lon)
	if err != nil {
		return 0, err
	}
	err = binary.Write(buf, binary.LittleEndian, mo.BoundingBox.Max.Lat)
	if err != nil {
		return 0, err
	}
	err = binary.Write(buf, binary.LittleEndian, mo.BoundingBox.Max.Lon)
	if err != nil {
		return 0, err
	}

	// Write length of Points slice
	err = binary.Write(buf, binary.LittleEndian, int64(len(mo.Points)))
	if err != nil {
		return 0, err
	}

	// Write Points
	for _, p := range mo.Points {
		err = binary.Write(buf, binary.LittleEndian, p.Lat)
		if err != nil {
			return 0, err
		}
		err = binary.Write(buf, binary.LittleEndian, p.Lon)
		if err != nil {
			return 0, err
		}
	}

	// Write to file
	n, err := file.Write(buf.Bytes())
	return int64(n), err
}

func ReadMapObject(mmapData *[]byte, offset int64, mo *MapObject) error {
	// Go to the correct offset

	// Create a bytes reader for the buffer
	reader := bytes.NewReader((*mmapData)[offset : offset+40])

	// Read BoundingBox and length from the buffer
	binary.Read(reader, binary.LittleEndian, &mo.BoundingBox.Min.Lat)
	binary.Read(reader, binary.LittleEndian, &mo.BoundingBox.Min.Lon)
	binary.Read(reader, binary.LittleEndian, &mo.BoundingBox.Max.Lat)
	binary.Read(reader, binary.LittleEndian, &mo.BoundingBox.Max.Lon)

	var lenPoints int64
	binary.Read(reader, binary.LittleEndian, &lenPoints)

	// Create a bytes reader for the buffer
	reader = bytes.NewReader((*mmapData)[offset+40 : offset+40+lenPoints*16])

	// Ensure Points slice is big enough
	if cap(mo.Points) < int(lenPoints) {
		mo.Points = make([]Point, lenPoints)
	} else {
		mo.Points = mo.Points[:lenPoints]
	}

	// Read Points from the buffer
	for i := int64(0); i < lenPoints; i++ {
		var p Point
		binary.Read(reader, binary.LittleEndian, &p.Lat)
		binary.Read(reader, binary.LittleEndian, &p.Lon)
		mo.Points[i] = p
	}

	return nil
}
