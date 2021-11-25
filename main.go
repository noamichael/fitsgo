package main

import (
	"fmt"

	"github.com/noamichael/fitsgo/fits"
)

func main() {
	fits := fits.Parse("samples/color.fit")
	fmt.Println(fits.HeadersRaw())
	fits.SaveAsJpeg()
}
