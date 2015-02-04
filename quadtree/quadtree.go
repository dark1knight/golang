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
	"sort"
)

const (
	ITERATIONS = 5000
	ERROR      = 1
	DEBUG      = false
	VERBOSE    = false
)

type Node struct {
	rect     image.Rectangle
	avgColor color.RGBA
	err      color.RGBA
}

/** comparison functions **/
type ByArea []Node

func (s ByArea) Len() int {
	return len(s)
}

func (s ByArea) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByArea) Less(i, j int) bool {
	areaI := s[i].rect.Dx() * s[i].rect.Dy()
	areaJ := s[j].rect.Dx() * s[j].rect.Dy()
	return areaJ < areaI
}

type QuadTree struct {
	workingList []Node
	finalList   []Node
	original    *image.Image
}

func getError(colour color.RGBA) float32 {
	return float32(colour.R+colour.G+colour.B) / 3
}

func getRgb(inR uint32, inG uint32, inB uint32, inA uint32) (r float64, g float64, b float64) {
	r = float64(inR / inA)
	g = float64(inG / inA)
	b = float64(inB / inA)
	return r, g, b
}

func averageColor(img *image.Image, bounds image.Rectangle) color.RGBA {
	var r, g, b, a float32
	r = 0
	g = 0
	b = 0
	a = 0
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			thisR, thisG, thisB, thisA := (*img).At(i, j).RGBA()
			if DEBUG && VERBOSE {
				fmt.Printf("color = (%d, %d, %d, %d)\n", thisR, thisG, thisB, thisA)
			}
			aSqrt := float32(math.Sqrt(float64(thisA)))
			r += float32(thisR) / aSqrt
			g += float32(thisG) / aSqrt
			b += float32(thisB) / aSqrt
			a += aSqrt
		}
	}
	area := float32(bounds.Dx()) * float32(bounds.Dy())
	if DEBUG {
		fmt.Printf("(before) area = %f, r = %f, g = %f, b = %f, a = %f\n", area, r, g, b, a)
	}
	r /= area
	g /= area
	b /= area
	a /= area
	if DEBUG {
		fmt.Printf("(after) area = %f, r = %f, g = %f, b = %f, a = %f\n", area, r, g, b, a)
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func meanSquaredError(img *image.Image, bounds image.Rectangle, avg color.RGBA) color.RGBA {
	var rErr, gErr, bErr float64
	rErr = 0
	gErr = 0
	bErr = 0
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			thisR, thisG, thisB, thisA := (*img).At(i, j).RGBA()
			thisA = uint32(math.Sqrt(float64(thisA)))
			r, g, b := getRgb(thisR, thisG, thisB, thisA)

			rErr += math.Pow(float64(avg.R)-r, 2.0)
			gErr += math.Pow(float64(avg.G)-g, 2.0)
			bErr += math.Pow(float64(avg.B)-b, 2.0)
		}
	}

	area := 2 * float64(bounds.Dx()) * float64(bounds.Dy())
	err := color.RGBA{uint8(rErr / area), uint8(gErr / area), uint8(bErr / area), 0}
	return err
}

func makeNode(img *image.Image, bounds image.Rectangle, nodeChannel chan<- Node) {
	avgColor := averageColor(img, bounds)
	meanSquare := meanSquaredError(img, bounds, avgColor)
	if DEBUG {
		fmt.Printf("Rectangle = (%d, %d) -> (%d, %d)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
		fmt.Printf("average color = (%d, %d, %d, %d), error = (%d, %d, %d)\n",
			avgColor.R, avgColor.G, avgColor.B, avgColor.A, meanSquare.R, meanSquare.G, meanSquare.B)
	}
	node := Node{bounds, avgColor, meanSquare}
	nodeChannel <- node
}

func upperLeft(rect image.Rectangle) image.Rectangle {
	xMidPoint := rect.Min.X + rect.Dx()/2
	yMidPoint := rect.Min.Y + rect.Dy()/2
	if DEBUG {
		fmt.Printf("midx = %d, midy = %d\n", xMidPoint, yMidPoint)
	}
	return image.Rect(rect.Min.X, rect.Min.Y, xMidPoint, yMidPoint)
}

func upperRight(rect image.Rectangle) image.Rectangle {
	xMidPoint := rect.Min.X + rect.Dx()/2
	yMidPoint := rect.Min.Y + rect.Dy()/2
	if DEBUG {
		fmt.Printf("midx = %d, midy = %d\n", xMidPoint, yMidPoint)
	}
	return image.Rect(xMidPoint+1, rect.Min.Y, rect.Max.X, yMidPoint)
}

func lowerLeft(rect image.Rectangle) image.Rectangle {
	xMidPoint := rect.Min.X + rect.Dx()/2
	yMidPoint := rect.Min.Y + rect.Dy()/2
	if DEBUG {
		fmt.Printf("midx = %d, midy = %d\n", xMidPoint, yMidPoint)
	}
	return image.Rect(rect.Min.X, yMidPoint+1, xMidPoint, rect.Max.Y)
}

func lowerRight(rect image.Rectangle) image.Rectangle {
	xMidPoint := rect.Min.X + rect.Dx()/2
	yMidPoint := rect.Min.Y + rect.Dy()/2
	if DEBUG {
		fmt.Printf("midx = %d, midy = %d\n", xMidPoint, yMidPoint)
	}
	return image.Rect(xMidPoint+1, yMidPoint+1, rect.Max.X, rect.Max.Y)
}

type QuarteringStrategy func(image.Rectangle) image.Rectangle

func (qTree *QuadTree) Build(img *image.Image) {
	qTree.original = img
	qTree.workingList = make([]Node, 0)
	qTree.finalList = make([]Node, 0)
	nodeChan := make(chan Node, 1)
	makeNode(img, (*img).Bounds(), nodeChan)
	firstNode := <-nodeChan
	qTree.workingList = append(qTree.workingList, firstNode)

	strategies := []QuarteringStrategy{upperLeft, upperRight,
		lowerLeft, lowerRight}
	for i := 0; i < ITERATIONS; i++ {
		if len(qTree.workingList) == 0 {
			fmt.Println("Finished...no more to do\n")
			break
		}
		/** get first element from working list **/
		numStrategies := len(strategies)
		currentNode := qTree.workingList[0]
		if DEBUG {
			fmt.Printf("Processing node (%d, %d) -> (%d, %d)\n", currentNode.rect.Min.X, currentNode.rect.Min.Y, currentNode.rect.Max.X, currentNode.rect.Max.Y)
			avgColor := currentNode.avgColor
			meanSquare := currentNode.err
			fmt.Printf("average color = (%d, %d, %d, %d), error = (%d, %d, %d)\n",
				avgColor.R, avgColor.G, avgColor.B, avgColor.A, meanSquare.R, meanSquare.G, meanSquare.B)
		}
		nodeChannels := make([]chan Node, numStrategies)
		/** initialize all nodes **/
		for i := range nodeChannels {
			nodeChannels[i] = make(chan Node)
		}
		if DEBUG && VERBOSE {
			fmt.Printf("number of strategies = %d, num node channels = %d\n", numStrategies, len(nodeChannels))
		}
		for k := range nodeChannels {
			go makeNode(img, strategies[k](currentNode.rect), nodeChannels[k])
		}
		/** wait for all responses **/
		for j := range nodeChannels {
			node := <-nodeChannels[j]
			if getError(node.err) < ERROR {
				qTree.finalList = append(qTree.finalList, node)
			} else {
				qTree.workingList = append(qTree.workingList, node)
			}
		}

		/** update the current working list **/
		qTree.workingList = qTree.workingList[1:]
	}
	fmt.Printf("Finished iterating, computed num final nodes = %d, num working nodes = %d\n", len(qTree.finalList), len(qTree.workingList))

	/** whatever is left in the working list now should be merged with the final list **/
	qTree.finalList = append(qTree.finalList, qTree.workingList...)
	fmt.Printf("Merged lists. List len = %d\n", len(qTree.finalList))
}

func (qTree *QuadTree) Paint() {
	bounds := (*qTree.original).Bounds()
	imgCopy := image.NewRGBA(bounds)
	/** sort to make sure that the smaller squares always override the larger ones **/
	sort.Sort(ByArea(qTree.finalList))
	for index := 0; index < len(qTree.finalList); index++ {
		colour := qTree.finalList[index].avgColor
		subImageBounds := qTree.finalList[index].rect
		for i := subImageBounds.Min.X; i < subImageBounds.Max.X; i++ {
			for j := subImageBounds.Min.Y; j < subImageBounds.Max.Y; j++ {
				imgCopy.SetRGBA(i, j, colour)
			}
		}
	}
	f, err := os.Create("test.jpg")
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

	if 0 == len(file) {
		fmt.Println("Usage: ./quadtree -file=<file>")
		return
	}

	fileImage, err := os.Open(file)
	if nil != err {
		panic("Could not open file.")
	}
	image, err := jpeg.Decode(fileImage)
	if nil != err {
		panic("Could not decode jpeg image!")
	}
	/** we've got our jpeg image, start processing **/
	var qTree QuadTree
	qTree.Build(&image)
	qTree.Paint()
}

func getFile(fileName_p *string) {
	flag.StringVar(fileName_p, "file", "", "input file name")
	flag.Parse()

}
