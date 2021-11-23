package main

import (
	"github.com/noamichael/fitsgo/fits"
)

func main() {
	fits := fits.Parse("samples/color.fit")
	fits.SaveAsJpeg()
}
