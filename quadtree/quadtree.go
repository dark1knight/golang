package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"strings"
)

type Node struct {
	r1               image.Rectangle
	r2               image.Rectangle
	r3               image.Rectangle
	r4               image.Rectangle
	parent           *Node
	meanSquaredError color.Color
}

func getFile(fileName_p *string) {
	flag.StringVar(fileName_p, "file", "", "input file name")
	flag.Parse()
}

func calcAverage(img *image.Image, rectangle image.Rectangle) color.Color {
	var r, g, b, a uint32
	r = 0
	g = 0
	b = 0
	a = 0
	for i := rectangle.Min.X; i < rectangle.Max.X; i++ {
		for j := rectangle.Min.Y; j < rectangle.Max.Y; j++ {
			R, G, B, A := (*img).At(i, j).RGBA()
			r += R
			g += G
			b += B
			a += A
		}
	}
	area := rectangle.Dx() * rectangle.Dy()
	r /= area
	g /= area
	b /= area
	a /= area
	rgba := color.RGBA
	rgba.R = uint8(r)
	rgba.G = uint8(g)
	rgba.B = uint8(b)
	rgba.A = uint8(a)

	return rgba
}

func buildTree(img *image.Image, rect *image.Rectangle) (parent *Node) {
	var node Node
	if nil == rect {
		// this is the parent node
	}

	return &node
}

func main() {
	var file string
	getFile(&file)
	file = strings.Trim(file, " ")
	if 0 == len(file) {
		fmt.Println("Usage : ./quadtree -file=<filename>")
		return
	}
	fmt.Printf("file has value '%s'\n", file)

	content, err := ioutil.ReadFile(file)
	if nil != err {
		panic("Could not open file")
	}

	reader := bytes.NewReader(content)

	image, err := jpeg.Decode(reader)
	if nil != err {
		panic("Could not decode jpeg image!")
	}

	bounds := image.Bounds()
	fmt.Printf("(%d, %d) to (%d, %d)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
}
