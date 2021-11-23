package fits

import (
	"fmt"
	"log"
	"strings"
)

type rgbConsumer func(row, col int, r, g, b uint8)
type grayScaleConsumer func(row, col int, value uint16)

func (f *File) debayer(bayerPattern string, consumer rgbConsumer) bool {
	width, _ := f.NaxisHeader(1)
	height, _ := f.NaxisHeader(2)

	if bayerPattern != "RGGB" {
		log.Fatalf("unsupported Bayer Pattern: %s\n", bayerPattern)
	}

	rowEven := bayerPattern[0:2]
	rowOdd := bayerPattern[2:4]

	getAtScaled := func(x, y int) int {
		value := f.imageData.ReadAsInt(x, y)
		// scaled := value / 256
		// if scaled > 255 {
		// 	scaled = 255
		// }
		return value
	}

	fmt.Println(rowOdd, rowEven)

	for row := 0; row < height; row++ {
		rowIsEven := row%2 == 0
		for col := 0; col < width; col++ {

			if row == 0 || row >= height-1 || col == 0 || col == width-1 {
				continue
			}

			color := ""

			if rowIsEven {
				color = string(rowEven[col%2])
			} else {
				color = string(rowOdd[col%2])
			}

			switch color {
			case "R":
				redValue := getAtScaled(row, col)
				// top left
				blue1 := getAtScaled(row-1, col-1)
				// top right
				blue2 := getAtScaled(row+1, col-1)
				// bottom left
				blue3 := getAtScaled(row-1, col+1)
				// bottom right
				blue4 := getAtScaled(row+1, col+1)

				blueAverage := (blue1 + blue2 + blue3 + blue4) / 4

				// left
				green1 := getAtScaled(row-1, col)
				// top
				green2 := getAtScaled(row, col-1)
				// right
				green3 := getAtScaled(row+1, col)
				// bottom
				green4 := getAtScaled(row, col+1)

				greenAverage := (green1 + green2 + green3 + green4) / 4

				consumer(col, row, uint8(redValue), uint8(greenAverage), uint8(blueAverage))

				break
			case "G":
				greenValue := getAtScaled(row, col)
				blueAverage := 0
				redAverage := 0

				if rowIsEven {
					red1 := getAtScaled(row-1, col)
					red2 := getAtScaled(row+1, col)
					redAverage = (red1 + red2) / 2
					blue1 := getAtScaled(row, col-1)
					blue2 := getAtScaled(row, col+1)
					blueAverage = (blue1 + blue2) / 2
				} else {
					red1 := getAtScaled(row, col-1)
					red2 := getAtScaled(row, col+1)
					redAverage = (red1 + red2) / 2
					blue1 := getAtScaled(row-1, col)
					blue2 := getAtScaled(row+1, col)
					blueAverage = (blue1 + blue2) / 2
				}

				consumer(col, row, uint8(redAverage), uint8(greenValue), uint8(blueAverage))
				break
			case "B":
				blueValue := getAtScaled(row, col)
				// top left
				red1 := getAtScaled(row-1, col-1)
				// top right
				red2 := getAtScaled(row+1, col-1)
				// bottom left
				red3 := getAtScaled(row-1, col+1)
				// bottom right
				red4 := getAtScaled(row+1, col+1)

				redAverage := (red1 + red2 + red3 + red4) / 4

				// left
				green1 := getAtScaled(row-1, col)
				// top
				green2 := getAtScaled(row, col-1)
				// right
				green3 := getAtScaled(row+1, col)
				// bottom
				green4 := getAtScaled(row, col+1)

				greenAverage := (green1 + green2 + green3 + green4) / 4

				consumer(col, row, uint8(redAverage), uint8(greenAverage), uint8(blueValue))

				break
			}

		}
	}

	return true
}

func (f *File) forEachGrayScale(consumer grayScaleConsumer) bool {
	width, _ := f.NaxisHeader(1)
	height, _ := f.NaxisHeader(2)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixelValue := f.imageData.ReadAsInt(y, x)
			consumer(x, y, uint16(pixelValue))
		}
	}

	return true
}

func parseBayer(bayer string) string {
	return strings.TrimSpace(strings.ReplaceAll(bayer, "'", ""))
}
