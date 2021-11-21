package main

import "fmt"

func main() {
	fits := Parse("samples/Single__2021-11-12_23-03-07_Bin1x1_130s__-15C.fit")

	fmt.Print(fits.headersRaw)
}
