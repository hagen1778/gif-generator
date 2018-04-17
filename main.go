package main

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var imgDir string

func main() {
	var ok bool
	imgDir, ok = os.LookupEnv("IMAGE_DIR")
	if !ok {
		log.Fatalf("Set `IMAGE_DIR` variable")
	}

	dataset, ok := os.LookupEnv("PATH_DATASET")
	if !ok {
		log.Fatalf("Set `PATH_DATASET` variable")
	}

	model, ok := os.LookupEnv("PATH_MODEL")
	if !ok {
		log.Fatalf("Set `PATH_MODEL` variable")
	}

	num, ok := os.LookupEnv("IMAGE_NUMBER")
	if !ok {
		num = "50"
	}

	n, err := strconv.Atoi(num)
	if err != nil {
		log.Fatalf("wrong IMAGE_NUMBER: %s", err)
	}

	log.Println("Init gif-generator with params")
	log.Println("IMAGE_DIR", imgDir)
	log.Println("PATH_DATASET", dataset)
	log.Println("IMAGE_NUMBER", n)

	list(dataset)

	if _, err := os.Stat(model); err != nil {
		log.Fatalf(err.Error())
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	shutdown := make(chan struct{})
	go func() {
		for {
			switch <-c {
			case syscall.SIGTERM:
				log.Println("SIGTERM recevied. Going to shutdown gracefully...")
				time.Sleep(time.Second*5)
				close(shutdown)
			}
		}
	}()

	for i := 0; i < n; i++ {
		select {
		case <-shutdown:
			log.Println("shutting down")
			return
		default:
			generateImg()
		}
	}
}

func list(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		log.Println(f.Name())
	}
}

var (
	w, h       float64 = 500, 250
	palette            = color.Palette{}
	zCycle     float64 = 8
	zMin, zMax float64 = 1, 15

	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func generateImg() {
	numberOfStars := 500 + r.Intn(4000)
	circles := []Circle{}
	for len(circles) < numberOfStars {
		x, y := rand.Float64()*8-4, rand.Float64()*8-4
		if math.Abs(x) < 0.5 && math.Abs(y) < 0.5 {
			continue
		}
		z := rand.Float64() * zCycle
		rad := r.Intn(5)
		circles = append(circles, Circle{x * w, y * h, z, float64(rad)})
	}

	// Intiialize palette (#000000, #111111, ..., #ffffff)
	palette = color.Palette{}
	for i := 0; i < 16; i++ {
		palette = append(palette, color.Gray{uint8(i) * 0x11})
	}

	// Generate 30 frames
	var images []*image.Paletted
	var delays []int
	count := 30
	for i := 0; i < count; i++ {
		pm := drawFrame(circles, float64(i)/float64(count))
		images = append(images, pm)
		delays = append(delays, 4)
	}

	// Output gif
	f, err := ioutil.TempFile(imgDir, "space-")
	if err != nil {
		log.Fatalf("unable to create file: %s", err)
	}
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	if err := os.Rename(f.Name(), f.Name()+".gif"); err != nil {
		log.Fatalf("unable to rename file: %s", err)
	}
}

type Point struct {
	X, Y float64
}

type Circle struct {
	X, Y, Z, R float64
}

// Draw stars in order to generate perfect loop GIF
func (c *Circle) Draw(gc *draw2dimg.GraphicContext, ratio float64) {
	z := c.Z - ratio*zCycle
	for z < zMax {
		if z >= zMin {
			x, y, r := c.X/z, c.Y/z, c.R/z
			gc.SetFillColor(color.White)
			gc.Fill()
			draw2dkit.Circle(gc, w/2+x, h/2+y, r)
			gc.Close()
		}
		z += zCycle
	}
}

func drawFrame(circles []Circle, ratio float64) *image.Paletted {
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	gc := draw2dimg.NewGraphicContext(img)

	// Draw background
	gc.SetFillColor(color.Gray{0x11})
	draw2dkit.Rectangle(gc, 0, 0, w, h)
	gc.Fill()
	gc.Close()

	// Draw stars
	for _, circle := range circles {
		circle.Draw(gc, ratio)
	}

	// Dithering
	pm := image.NewPaletted(img.Bounds(), palette)
	draw.FloydSteinberg.Draw(pm, img.Bounds(), img, image.ZP)
	return pm
}
