package main

import (
	"github.com/noamichael/fitsgo/fits"
)

func main() {
	fits := fits.Parse("samples/Single__2021-11-12_23-03-07_Bin1x1_130s__-15C.fit")
	fits.SaveAsJpeg()
}
