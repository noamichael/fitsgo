package fits

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// The following parser is based on the FITS standard
// version 4.0

// Spec: Each FITS structure shall consist of an integral number of
// FITS blocks, which are each 2880 bytes (23040 bits) in length.
var FITS_BLOCK_SIZE = 2880

type Header struct {
	Keyword string
	Value   string
	Comment string
}

type HeaderDataUnit struct {
	Headers     map[string]*Header
	fits        *File
	Data        Data
	headerStart int64
	headerEnd   int64
	dataStart   int64
	dataEnd     int64
	blocks      int64
	read        bool
}

type File struct {
	HeaderDataUnits []*HeaderDataUnit
	headersRaw      string
	readOffset      int64
	fileSize        int64
	filename        string
}

func (f *File) hasMoreData() bool {
	return f.fileSize-f.readOffset > 0
}

func (f *HeaderDataUnit) NaxisHeader(index int) (int, error) {
	naxisHeaderKey := "NAXIS"

	if index > 0 {
		naxisHeaderKey = fmt.Sprintf("NAXIS%d", index)
	}

	return f.HeaderInt(naxisHeaderKey)
}

func (f *HeaderDataUnit) HeaderInt(name string) (int, error) {

	header, headHeader := f.Headers[name]

	if !headHeader {
		return 0, errors.New("could not find " + name)
	}

	headerValue, err := strconv.Atoi(header.Value)

	if err != nil {
		return 0, fmt.Errorf("could not parse NAXIS header, %e", err)
	}

	return int(headerValue), nil
}

func (f *HeaderDataUnit) HeaderFloat(name string) (float32, error) {

	header, headHeader := f.Headers[name]

	if !headHeader {
		return 0, errors.New("could not find " + name)
	}

	headerValue, err := strconv.ParseFloat(header.Value, 32)

	if err != nil {
		return 0, fmt.Errorf("could not parse NAXIS header, %e", err)
	}

	return float32(headerValue), nil
}

func (f *File) HeadersString() string {
	return f.headersRaw
}

func Parse(filename string) *File {
	fitsFile, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer fitsFile.Close()

	info, _ := fitsFile.Stat()
	fileSize := info.Size()

	fmt.Printf("File Size: %d Bytes\n", fileSize)

	// Step 1: parse headers
	fits := &File{
		HeaderDataUnits: make([]*HeaderDataUnit, 0),
		fileSize:        fileSize,
		filename:        filename,
	}

	for {
		headerDataUnit := fits.parseHeaders(fitsFile)
		headerDataUnit.dataStart = fits.readOffset
		headerDataUnit.dataEnd = fits.readOffset + headerDataUnit.calculateEnd()
		hduSize := float64(headerDataUnit.dataEnd - headerDataUnit.dataStart)
		headerDataUnit.blocks = int64(math.Ceil(hduSize / float64(FITS_BLOCK_SIZE)))
		fmt.Printf("HDU Start: %d, HDU End: %d\n", headerDataUnit.dataStart, headerDataUnit.dataEnd)

		nextHeaderStart := fits.readOffset + headerDataUnit.blocks*int64(FITS_BLOCK_SIZE)

		if fits.readOffset+headerDataUnit.blocks*int64(FITS_BLOCK_SIZE) >= fits.fileSize {
			break
		}

		fits.readOffset = nextHeaderStart
	}

	return fits
}

func (f *File) parseHeaders(fs *os.File) *HeaderDataUnit {

	headerDataUnit := &HeaderDataUnit{
		Headers:     make(map[string]*Header),
		headerStart: f.readOffset,
		fits:        f,
	}

	parsingHeaders := true

	for parsingHeaders {
		buffer := make([]byte, FITS_BLOCK_SIZE)
		read, err := fs.ReadAt(buffer, f.readOffset)
		f.readOffset += int64(read)

		if err != nil {
			log.Fatal(err)
		}

		currentHeader := ""
		for _, b := range buffer {
			currentHeader += string(b)
			if len(currentHeader) >= 80 {
				f.headersRaw += currentHeader + "\n"
				header := f.parseAndAddHeader(currentHeader, headerDataUnit.Headers)
				if header.Keyword == "END" {
					parsingHeaders = false
					break
				}
				currentHeader = ""
			}
		}

		// TODO: handle case where we've hit the end of the file
		// but we haven't found the end of the headers yet
		if f.readOffset >= f.fileSize {
			break
		}
	}

	fmt.Printf("Read Offset: %d Bytes\n", f.readOffset)

	f.HeaderDataUnits = append(f.HeaderDataUnits, headerDataUnit)
	headerDataUnit.headerEnd = f.readOffset

	return headerDataUnit
}

// Parses and adds a header record
// Spec 3.3.1: The header of a primary HDU shall consist of one or more
// header blocks, each containing a series of 80-character keyword
// records containing only the restricted set of ASCII-text characters. Each 2880-byte header block contains 36 keyword records.
// The last header block must contain the END keyword (defined in
// Sect. 4.4.1), which marks the logical end of the header. Keyword
// records without information (e.g., following the END keyword)
// shall be filled with ASCII spaces (decimal 32 or hexadecimal20).
func (f *File) parseAndAddHeader(raw string, headers map[string]*Header) *Header {

	equalsIndex := strings.Index(raw, "=")
	keyRaw := ""
	valueAndComment := ""

	// Has equals sign
	if equalsIndex > -1 {
		keyRaw = raw[0:equalsIndex]
		valueAndComment = raw[equalsIndex+1:]
	} else {
		if strings.HasPrefix(raw, "END") {
			return &Header{Keyword: "END"}
		}
		// This means the header span multiple lines
		// Spec: 4.2.1.2 Continued string (long-string) keywords
		// TODO: Support
		return &Header{}
	}

	// parse value
	valueAndCommentParts := strings.Split(valueAndComment, "/")
	valueAndCommentPartsLen := len(valueAndCommentParts)
	value := ""
	comment := ""

	if valueAndCommentPartsLen > 0 {
		value = valueAndCommentParts[0]
	}

	if valueAndCommentPartsLen == 2 {
		comment = valueAndCommentParts[1]
	}

	header := &Header{
		Keyword: strings.TrimSpace(keyRaw),
		Value:   strings.TrimSpace(value),
		Comment: strings.TrimSpace(comment),
	}

	headers[header.Keyword] = header

	return header

}

func (f *File) HeadersRaw() string {
	return f.headersRaw
}

func (hdu *HeaderDataUnit) parseData() {
	fs, err := os.Open(hdu.fits.filename)

	if err != nil {
		log.Fatal(err)
	}

	defer fs.Close()

	hdu.fits.readOffset = hdu.dataStart
	// The number of dimensions for the table data
	// Spec: The primary data array, if present, shall consist of a single data
	// array with from 1 to 999 dimensions
	// (as specified by the NAXI keyword defined in Sect. 4.4.1).
	naxis, err := hdu.NaxisHeader(0)

	if err != nil {
		return
	}

	if naxis != 2 {
		fmt.Printf("[ERROR]: Only 2 axes are supported. This file contains %d \n", naxis)
		os.Exit(1)
	}

	// The dimensions for each array TODO: Flip to Big Endian
	// The is usually of size 2. Example: [400, 200,] for a 400x200 image
	width, _ := hdu.NaxisHeader(1)
	height, _ := hdu.NaxisHeader(2)
	bitpix, _ := hdu.HeaderInt("BITPIX")
	bzero, _ := hdu.HeaderFloat("BZERO")
	bscale, _ := hdu.HeaderFloat("BSCALE")

	if bscale <= 0 {
		bscale = 1
	}

	hdu.Data = NewData(width, height, bitpix, bzero, bscale)

	// BITPIX to pixel
	pixelDataSize := int(math.Abs(float64(bitpix)) / 8)
	// Figure out how many pixels are in a block
	pixelDataPerBlock := FITS_BLOCK_SIZE / pixelDataSize

	fmt.Printf("pixelDataSize = %d, pixelDataPerBlock=%d\n", pixelDataSize, pixelDataPerBlock)
	fmt.Printf("naxis = %d bitpix = %d\n", naxis, bitpix)
	fmt.Printf("Total file size = %d, read offset = %d\n", hdu.fits.fileSize, hdu.fits.readOffset)

	row := 0
	col := 0

	for {
		dataBlock := make([]byte, FITS_BLOCK_SIZE)
		read, _ := fs.ReadAt(dataBlock, hdu.fits.readOffset)
		hdu.fits.readOffset += int64(read)

		// Stop looping if we've hit EOF
		if read < 1 {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		// Translate Big endian to little endian
		for pixel := 0; pixel < FITS_BLOCK_SIZE; pixel += pixelDataSize {
			pixelData := dataBlock[pixel:(pixel + pixelDataSize)]

			hdu.Data.Write(row, col, pixelData)

			col++
			if col >= width {
				col = 0
				row++
				if row >= height {
					break
				}
			}

		}

		if col >= width || row >= height {
			break
		}

	}

	hdu.read = true

	fmt.Println("Done reading!")

}

func (hdu *HeaderDataUnit) SaveAsJpeg() {

	if !hdu.read {
		hdu.parseData()
	}

	out, _ := os.Create("./samples/test.jpg")

	width, _ := hdu.NaxisHeader(1)
	height, _ := hdu.NaxisHeader(2)
	bayerPatternHeader, colorImage := hdu.Headers["BAYERPAT"]

	rectangle := image.Rect(0, 0, width, height)

	var img image.Image

	if !colorImage {
		grayImg := image.NewGray(rectangle)
		hdu.forEachGrayScale(func(x, y int, value uint16) {
			grayImg.Set(x, y, color.Gray16{Y: value * 20})
		})
		img = grayImg
	} else {
		bayerPattern := parseBayer(bayerPatternHeader.Value)
		colorScaleImg := image.NewRGBA(rectangle)
		hdu.debayer(bayerPattern, func(x, y int, r, g, b uint8) {
			colorScaleImg.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		})
		img = colorScaleImg
	}

	jpeg.Encode(out, img, &jpeg.Options{Quality: 100})
}

func (hdu *HeaderDataUnit) calculateEnd() int64 {

	width, _ := hdu.NaxisHeader(1)
	height, _ := hdu.NaxisHeader(2)
	bitpix, _ := hdu.HeaderInt("BITPIX")
	bytesPerPixel := int64(math.Abs(float64(bitpix)) / 8)

	return int64(height) * int64(width) * bytesPerPixel
}
