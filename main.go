package main

import (
	"fmt"

	"github.com/noamichael/fitsgo/fits"
)

func main() {
	fits := fits.Parse("samples/color.fits")
	fmt.Println(fits.HeadersRaw())
	fits.HeaderDataUnits[0].SaveAsJpeg()
}
