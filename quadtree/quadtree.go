package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"strings"
	"sync"
)

const (
	ERROR_THRESHOLD = 10
	NUM_ITERATIONS  = 10000
	DEBUG           = false
)

var iteration int = 0

type Node struct {
	dimensions       image.Rectangle
	children         [4]image.Rectangle
	meanSquaredError color.RGBA
	averageColor     color.RGBA
	hasChildren      bool
}

type QuadTree struct {
	root      Node
	numNodes  int32
	leafNodes []Node
	mutex     sync.Mutex
}

func getMeanSquareError(colour *color.RGBA) float32 {
	return float32(colour.A+colour.B+colour.R+colour.G) / 4
}

func (qTree *QuadTree) appendLeaf(node Node) {
	qTree.mutex.Lock()
	qTree.leafNodes = append(qTree.leafNodes, node)
	qTree.mutex.Unlock()
}

func (qTree *QuadTree) processNode(img *image.Image, rect image.Rectangle, node Node, done chan bool) {
	if DEBUG {
		fmt.Printf("Now processing node : (%d, %d) to (%d, %d)\n",
			rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
	}
	{
		qTree.mutex.Lock()
		iteration += 1
		qTree.numNodes += 1
		qTree.mutex.Unlock()
	}
	bounds := node.dimensions
	dX := bounds.Dx()
	dY := bounds.Dy()
	if 0 == dX || 0 == dY {
		if DEBUG {
			fmt.Printf("Rectangle is 1 pixed wide in at least 1 dimension. Marking as leaf.\n")
		}
		qTree.appendLeaf(node)
		done <- true
		return
	}
	node.averageColor = calcAverage(img, &bounds)
	node.meanSquaredError = calcMeanSquaredError(img, &bounds, &node.averageColor)
	err := getMeanSquareError(&node.meanSquaredError)
	if DEBUG {
		fmt.Println("Mean squared error and average color is ", err, node.averageColor)
	}
	if err < ERROR_THRESHOLD {
		if DEBUG {
			fmt.Printf("error is %f\n", err)
		}
		qTree.appendLeaf(node)
		done <- true
		return
	}

	if iteration >= NUM_ITERATIONS {
		if DEBUG {
			fmt.Printf("Finished %d iterations.\n", NUM_ITERATIONS)
		}
		qTree.appendLeaf(node)
		done <- true
		return
	}

	xMidPoint := bounds.Min.X + dX/2
	yMidPoint := bounds.Min.Y + dY/2

	/** subdivide **/
	node.children[0] = image.Rect(0, 0, xMidPoint, yMidPoint)
	node.children[1] = image.Rect(0, yMidPoint+1, xMidPoint, bounds.Max.Y)
	node.children[2] = image.Rect(xMidPoint+1, 0, bounds.Max.X, yMidPoint)
	node.children[3] = image.Rect(xMidPoint+1, yMidPoint+1, bounds.Max.X, bounds.Max.Y)

	doneChildren := make(chan bool, 4)
	var nodeR1, nodeR2, nodeR3, nodeR4 Node
	childNodes := []*Node{&nodeR1, &nodeR2, &nodeR3, &nodeR4}
	node.hasChildren = true
	for i := 0; i < len(childNodes); i++ {
		childNodes[i].dimensions = node.children[i]
		go qTree.processNode(img, node.children[i], *childNodes[i], doneChildren)
	}
	<-doneChildren
	<-doneChildren
	<-doneChildren
	<-doneChildren
	done <- true

}

func fillColor(outImage *image.RGBA, node *Node) {
	dims := node.dimensions
	if DEBUG {
		fmt.Printf("Node is : ", dims)
	}
	for i := dims.Min.X; i < dims.Max.X; i++ {
		for j := dims.Min.Y; j < dims.Max.Y; j++ {
			outImage.SetRGBA(i, j, node.averageColor)
		}
	}
}

func setBlank(outImage *image.RGBA) {
	bounds := (*outImage).Bounds()

	white := color.RGBA{255, 255, 255, 255}
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			outImage.SetRGBA(i, j, white)
		}
	}
}

func (qTree *QuadTree) drawImage() {
	if DEBUG {
		fmt.Printf("Drawing image with %d leaf nodes\n", len(qTree.leafNodes))
	}
	outImage := image.NewRGBA(qTree.root.dimensions)
	setBlank(outImage)

	for _, leaf := range qTree.leafNodes {
		fillColor(outImage, &leaf)
	}

	f, err := os.Create("owl_2.jpg")
	if nil != err {
		fmt.Printf("Encountered error %s\n", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	jpeg.Encode(writer, outImage.SubImage(outImage.Bounds()), &jpeg.Options{100})

}

func (qTree *QuadTree) BuildTree(img *image.Image, rect *image.Rectangle) {
	qTree.leafNodes = make([]Node, 1)
	qTree.root.dimensions = *rect
	//qTree.root.parent = nil
	done := make(chan bool, 1)
	go qTree.processNode(img, *rect, qTree.root, done)
	<-done
	qTree.drawImage()

	fmt.Printf("Built a tree with %d nodes.\n", qTree.numNodes)
}

func getFile(fileName_p *string) {
	flag.StringVar(fileName_p, "file", "", "input file name")
	flag.Parse()
}

func calcMeanSquaredError(img *image.Image, rectangle *image.Rectangle,
	averageColor *color.RGBA) color.RGBA {
	var rErr, gErr, bErr, aErr float64
	rErr = 0
	gErr = 0
	bErr = 0
	aErr = 0
	for i := rectangle.Min.X; i < rectangle.Max.X; i++ {
		for j := rectangle.Min.Y; j < rectangle.Max.Y; j++ {
			aR, aG, aB, Asq := (*img).At(i, j).RGBA() /** these values are alpha premultiplied **/
			A := uint32(math.Sqrt(float64(Asq)))
			r := float64(aR / A)
			g := float64(aG / A)
			b := float64(aB / A)
			rErr += math.Pow(float64(averageColor.R)-r, 2.0)
			gErr += math.Pow(float64(averageColor.G)-g, 2.0)
			bErr += math.Pow(float64(averageColor.B)-b, 2.0)
		}
	}
	area := 2 * float64(rectangle.Dx()) * float64(rectangle.Dy())
	err := color.RGBA{uint8(rErr / area), uint8(gErr / area), uint8(bErr / area), uint8(aErr / area)}
	return err
}

func calcAverage(img *image.Image, rectangle *image.Rectangle) color.RGBA {
	var r, g, b, a float32
	r = 0
	g = 0
	b = 0
	a = 0
	for i := rectangle.Min.X; i < rectangle.Max.X; i++ {
		for j := rectangle.Min.Y; j < rectangle.Max.Y; j++ {
			R, G, B, A := (*img).At(i, j).RGBA()
			A = uint32(math.Sqrt(float64(A)))
			if DEBUG {
				//fmt.Printf("r = %d, g = %d, b = %d, a = %d\n", R, G, B, A)
			}
			r += float32(R / A)
			g += float32(G / A)
			b += float32(B / A)
			a += float32(A)
		}
	}
	area := float32(rectangle.Dx()) * float32(rectangle.Dy())
	if DEBUG {
		fmt.Printf("area = %d, r = %d, g = %d, b = %d, a = %d\n", area, r, g, b, a)
	}
	r /= area
	g /= area
	b /= area
	a /= area
	rgba := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	return rgba
}

func simpleWrite(img *image.Image) {
	bounds := (*img).Bounds()
	imgCopy := image.NewRGBA(bounds)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			r, g, b, a := (*img).At(i, j).RGBA()
			colour := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			imgCopy.SetRGBA(i, j, colour)
		}
	}
	f, err := os.Create("owl_2.jpg")
	if nil != err {
		fmt.Printf("Encountered error %s\n", err)
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	jpeg.Encode(writer, imgCopy.SubImage(imgCopy.Bounds()), &jpeg.Options{jpeg.DefaultQuality})
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

	fileImage, err := os.Open(file)
	if nil != err {
		panic("Could not open file")
	}
	image, err := jpeg.Decode(fileImage)
	if nil != err {
		panic("Could not decode jpeg image!")
	}

	//simpleWrite(&image)

	bounds := image.Bounds()
	var qTree QuadTree
	qTree.BuildTree(&image, &bounds)
	fmt.Printf("(%d, %d) to (%d, %d)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
}
