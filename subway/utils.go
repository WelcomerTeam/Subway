package internal

import (
	"image/color"
	"strconv"
)

func parseHexNumber(arg string) (uint64, error) {
	return strconv.ParseUint(arg, 16, 64)
}

func intToColour(val uint64) color.RGBA {
	return color.RGBA{
		R: uint8((val >> 24) & 0xFF),
		G: uint8((val >> 16) & 0xFF),
		B: uint8((val >> 8) & 0xFF),
		A: uint8(val & 0xFF),
	}
}
