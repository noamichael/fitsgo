package fits

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
)

func NewData(width, height, bitpix int, bzero, bscale float32) Data {
	if bitpix == 8 {
		data := make([][]int8, height)

		for i := 0; i < height; i++ {
			data[i] = make([]int8, width)
		}
		return &Int8Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}

	if bitpix == 16 {
		data := make([][]int16, height)

		for i := 0; i < height; i++ {
			data[i] = make([]int16, width)
		}
		return &Int16Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}

	if bitpix == 32 {
		data := make([][]int32, height)

		for i := 0; i < height; i++ {
			data[i] = make([]int32, width)
		}
		return &Int32Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}

	if bitpix == 64 {
		data := make([][]int64, height)

		for i := 0; i < height; i++ {
			data[i] = make([]int64, width)
		}
		return &Int64Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}

	if bitpix == -32 {
		data := make([][]float32, height)

		for i := 0; i < height; i++ {
			data[i] = make([]float32, width)
		}
		return &Float32Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}
	if bitpix == -64 {
		data := make([][]float64, height)

		for i := 0; i < height; i++ {
			data[i] = make([]float64, width)
		}
		return &Float64Data{
			data:   data,
			bzero:  bzero,
			bscale: bscale,
		}
	}

	return nil

}

type Data interface {
	Write(row, col int, b []byte)
	ReadAsInt(row, col int) int
	GetMaxValue() float64
}

type Int8Data struct {
	data   [][]int8
	bzero  float32
	bscale float32
}

func (d *Int8Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Int8Data) Write(row, col int, b []byte) {
	var p int8
	readAs(b, &p)
	d.data[row][col] = (p + int8(d.bzero)) * int8(d.bscale)
}

func (d *Int8Data) GetMaxValue() float64 {
	return float64(math.MaxInt8)
}

type Int16Data struct {
	data   [][]int16
	bzero  float32
	bscale float32
}

func (d *Int16Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Int16Data) Write(row, col int, b []byte) {
	var p int16
	readAs(b, &p)
	d.data[row][col] = (p + int16(d.bzero)) * int16(d.bscale)
}

func (d *Int16Data) GetMaxValue() float64 {
	return float64(math.MaxInt16)
}

type Int32Data struct {
	data   [][]int32
	bzero  float32
	bscale float32
}

func (d *Int32Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Int32Data) Write(row, col int, b []byte) {
	var p int32
	readAs(b, &p)
	d.data[row][col] = (p + int32(d.bzero)) * int32(d.bscale)
}

func (d *Int32Data) GetMaxValue() float64 {
	return float64(math.MaxInt32)
}

type Int64Data struct {
	data   [][]int64
	bzero  float32
	bscale float32
}

func (d *Int64Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Int64Data) Write(row, col int, b []byte) {
	var p int64
	readAs(b, &p)
	d.data[row][col] = (p + int64(d.bzero)) * int64(d.bscale)
}

func (d *Int64Data) GetMaxValue() float64 {
	return float64(math.MaxInt64)
}

type Float32Data struct {
	data   [][]float32
	bzero  float32
	bscale float32
}

func (d *Float32Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Float32Data) Write(row, col int, b []byte) {
	bits := binary.BigEndian.Uint32(b)
	float := math.Float32frombits(bits)
	d.data[row][col] = (float + d.bzero) * d.bscale
}

func (d *Float32Data) GetMaxValue() float64 {
	return float64(math.MaxFloat32)
}

type Float64Data struct {
	data   [][]float64
	bzero  float32
	bscale float32
}

func (d *Float64Data) ReadAsInt(row, col int) int {
	return int(d.data[row][col])
}

func (d *Float64Data) Write(row, col int, b []byte) {
	var p float64
	readAs(b, &p)
	d.data[row][col] = (p + float64(d.bzero)) * float64(d.bscale)
}

func (d *Float64Data) GetMaxValue() float64 {
	return float64(math.MaxFloat64)
}

func readAs(data []byte, pnter interface{}) {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, pnter)

	if err != nil {
		log.Fatal(err)
	}
}
