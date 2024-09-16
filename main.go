/*
takes an input image and pixelates it
*/
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
)

type pixel struct {
	r, g, b int
	x, y    int
}

func main() {
	flag.CommandLine.Usage = printUsage
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 || len(args) >= 3 {
		printUsage()
		return
	}

	var res string
	if len(args) == 1 {
		res = "8"
	} else if len(args) == 2 {
		res = args[1]
	}

	path := args[0]

	file := openImage(path)

	width, height := file.Bounds().Max.X, file.Bounds().Max.Y
	scale, scaleErr := strconv.ParseInt(res, 0, 0)
	if scaleErr != nil {
		printUsage()
		return
	}

	fmt.Println(width, height, scale)
	newImg := make([][]pixel, width/int(scale))
	for x := range newImg {
		newImg[x] = make([]pixel, height/int(scale))
	}

	for x := range newImg {
		for y := range newImg[x] {
			newImg[x][y] = getAverageColor(file, x, y, int(scale))
		}
	}

	// Create a new image with the original dimensions
	outputImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// Calculate the corresponding pixel in the scaled-down image
			scaledX := x / int(scale)
			scaledY := y / int(scale)

			// Ensure we don't access out of bounds
			if scaledX >= len(newImg) {
				scaledX = len(newImg) - 1
			}
			if scaledY >= len(newImg[scaledX]) {
				scaledY = len(newImg[scaledX]) - 1
			}

			// Set the color of the output pixel
			outputImg.Set(x, y, color.RGBA{
				R: uint8(newImg[scaledX][scaledY].r),
				G: uint8(newImg[scaledX][scaledY].g),
				B: uint8(newImg[scaledX][scaledY].b),
				A: 255,
			})
		}
	}

	// Create the output file name
	outputPath := fmt.Sprintf("%s_%sdwnscl.png", path[:len(path)-len(filepath.Ext(path))], res)

	// Save the image as PNG
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, outputImg); err != nil {
		fmt.Println("Error encoding image:", err)
		return
	}

	fmt.Printf("Pixelated image saved as: %s\n", outputPath)
}

func getAverageColor(img image.Image, x, y, scale int) pixel {
	var totalR, totalG, totalB, count int
	for i := x * int(scale); i < (x+1)*int(scale) && i < img.Bounds().Max.X; i++ {
		for j := y * int(scale); j < (y+1)*int(scale) && j < img.Bounds().Max.Y; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			totalR += int(r >> 8)
			totalG += int(g >> 8)
			totalB += int(b >> 8)
			count++
		}
	}
	return pixel{
		r: totalR / count,
		g: totalG / count,
		b: totalB / count,
		x: x,
		y: y,
	}
}

func printUsage() {
	fmt.Println("dwnscl: downscales images")
	fmt.Println("usage: dwnscl <path/to/image> [strength int]")
}

func openImage(path string) image.Image {
	f, fErr := os.Open(path)
	if fErr != nil {
		fmt.Println("err: file could not be opened")
		fmt.Println(fErr)
		os.Exit(1)
	}
	defer f.Close()

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)

	img, _, imgErr := image.Decode(f)
	if imgErr != nil {
		fmt.Println("err: file could not be decoded")
		fmt.Println(imgErr)
		os.Exit(1)
	}
	return img
}
