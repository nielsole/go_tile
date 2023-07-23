package utils

import (
	"errors"
	"regexp"
	"strconv"
)

var _matcher = regexp.MustCompile(`^/tile/([0-9]+)/([0-9]+)/([0-9]+).(png|webp)$`)

func ParsePath(path string) (z, x, y uint32, ext string, err error) {
	matches := _matcher.FindStringSubmatch(path)
	if len(matches) != 5 {
		return 0, 0, 0, "", errors.New("could not match path")
	}
	zInt, err := strconv.Atoi(matches[1])
	if err != nil {
		return
	}
	xInt, err := strconv.Atoi(matches[2])
	if err != nil {
		return
	}
	yInt, err := strconv.Atoi(matches[3])
	if err != nil {
		return
	}
	ext = matches[4]
	z = uint32(zInt)
	x = uint32(xInt)
	y = uint32(yInt)
	return
}
