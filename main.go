package main

import (
	"fmt"

	"github.com/noamichael/fitsgo/fits"
)

func main() {
	fits := fits.Parse("samples/gray.fits")
	fmt.Println(fits.HeadersRaw())
	fits.HeaderDataUnits[0].SaveAsJpeg()
}
