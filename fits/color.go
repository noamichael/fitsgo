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

	fmt.Println(rowEven, rowOdd)

	// Do some bilinear interpolation
	for row := 0; row < height; row++ {
		rowIsEven := row%2 == 0
		for col := 0; col < width; col++ {
			color := ""

			if rowIsEven {
				color = string(rowEven[col%2])
			} else {
				color = string(rowOdd[col%2])
			}

			pix := &pixel{row: row, col: col, height: height, width: width, f: f}

			corners := []int{
				pix.getTopLeft(),
				pix.getTopRight(),
				pix.getBottomLeft(),
				pix.getBottomRight(),
			}

			neighbors := []int{
				pix.getLeft(),
				pix.getTop(),
				pix.getRight(),
				pix.getBottom(),
			}

			switch color {
			case "R":
				redValue := pix.getValue()

				blues := corners
				blueAverage := averagePositive(blues)

				greens := neighbors
				greenAverage := averagePositive(greens)

				consumer(col, row, uint8(redValue), uint8(greenAverage), uint8(blueAverage))

				break
			case "G":
				greenValue := pix.getValue()
				blueAverage := 0
				redAverage := 0
				leftRight := []int{pix.getLeft(), pix.getRight()}
				topBottom := []int{pix.getTop(), pix.getBottom()}

				if rowIsEven {
					redAverage = averagePositive(leftRight)
					blueAverage = averagePositive(topBottom)
				} else {
					blueAverage = averagePositive(leftRight)
					redAverage = averagePositive(topBottom)
				}

				consumer(col, row, uint8(redAverage), uint8(greenValue), uint8(blueAverage))
				break
			case "B":
				blueValue := pix.getValue()

				reds := corners
				redAverage := averagePositive(reds)

				greens := neighbors
				greenAverage := averagePositive(greens)

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

type pixel struct {
	row, col, height, width int
	f                       *File
}

// ensures the pixel at row, col are inbetween [0, 255]
func (p *pixel) getAtScaled(row, col int) int {
	// r = current range min/max
	rmin := float64(0)
	rmax := float64(p.f.imageData.GetMaxValue())
	// t = target range min/max
	tmin := float64(0)
	tmax := float64(255)
	// TODO Find out why these are negative
	//value := math.Abs(float64((p.f.imageData.ReadAsInt(row, col))))
	value := float64(p.f.imageData.ReadAsInt(row, col))
	// force negative values to be max brightness
	if value < 0 {
		value = rmax
	}
	// https://stats.stackexchange.com/questions/281162/scale-a-number-between-a-range
	scaled := (((value-rmin)/(rmax-rmin))*(tmax-tmin) + tmin)
	return int(scaled)
}

func (p *pixel) getValue() int {
	return p.getAtScaled(p.row, p.col)
}

func (p *pixel) getTopLeft() int {
	if p.row > 0 && p.col > 0 {
		return p.getAtScaled(p.row-1, p.col-1)
	}
	return -1
}

func (p *pixel) getTopRight() int {
	if p.row > 0 && p.col < p.width-1 {
		return p.getAtScaled(p.row-1, p.col+1)
	}
	return -1
}

func (p *pixel) getTop() int {
	if p.row > 0 {
		return p.getAtScaled(p.row-1, p.col)
	}
	return -1
}

func (p *pixel) getLeft() int {
	if p.col > 0 {
		return p.getAtScaled(p.row, p.col-1)
	}
	return -1
}

func (p *pixel) getRight() int {
	if p.col < p.width-1 {
		return p.getAtScaled(p.row, p.col+1)
	}
	return -1
}

func (p *pixel) getBottomLeft() int {
	if p.col > 0 && p.row < p.height-1 {
		return p.getAtScaled(p.row+1, p.col-1)
	}
	return -1
}

func (p *pixel) getBottomRight() int {
	if p.col < p.width-1 && p.row < p.height-1 {
		return p.getAtScaled(p.row+1, p.col+1)
	}
	return -1
}

func (p *pixel) getBottom() int {
	if p.row < p.height-1 {
		return p.getAtScaled(p.row+1, p.col)
	}
	return -1
}

func averagePositive(n []int) int {
	count := 0
	average := 0
	for i := 0; i < len(n); i++ {
		if n[i] >= 0 {
			average += n[i]
			count++
		}
	}

	return average / count
}
