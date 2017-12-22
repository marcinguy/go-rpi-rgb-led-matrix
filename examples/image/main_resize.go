package main

// This should resize the animated git to the size of the display
//
//
// Frames in an animated gif aren't necessarily the same size, subsequent
// frames are overlayed on previous frames. Therefore, resizing the frames
// individually may cause problems due to aliasing of transparent pixels. This
// example tries to avoid this by building frames from all previous frames and
// resizing the frames as RGB.

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"os"
	"github.com/nfnt/resize"
        "flag"
	"time"
        "github.com/disintegration/imaging"
	"github.com/mcuadros/go-rpi-rgb-led-matrix"

)

var (
	rows       = flag.Int("led-rows", 32, "number of rows supported")
	parallel   = flag.Int("led-parallel", 1, "number of daisy-chained panels")
	chain      = flag.Int("led-chain", 2, "number of displays daisy-chained")
	brightness = flag.Int("brightness", 100, "brightness (0-100)")
	img        = flag.String("image", "", "image path")

	rotate = flag.Int("rotate", 0, "rotate angle, 90, 180, 270")
)


func main() {

	process(*img)
}

func process(filename string) {

	// Open image file.
	f, err := os.Open(filename )
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	// Decode the original gif.
	im, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create a new RGBA image to hold the incremental frames.
	firstFrame := im.Image[0].Bounds()
	b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
	img := image.NewRGBA(b)

	// Resize each frame.
	for index, frame := range im.Image {
		bounds := frame.Bounds()
		draw.Draw(img, bounds, frame, bounds.Min, draw.Over)
		im.Image[index] = ImageToPaletted(ProcessImage(img))
	}

        out, err := os.Create(filename + ".out.fixed.gif")
        if err != nil {
                log.Fatal(err.Error())
        }
        defer out.Close()

        gif.EncodeAll(out, im)


        config := &rgbmatrix.DefaultConfig
	config.Rows = *rows
	config.Parallel = *parallel
	config.ChainLength = *chain
	config.Brightness = *brightness

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	fatal(err)

	tk := rgbmatrix.NewToolKit(m)
	defer tk.Close()

	switch *rotate {
	case 90:
		tk.Transform = imaging.Rotate90
	case 180:
		tk.Transform = imaging.Rotate180
	case 270:
		tk.Transform = imaging.Rotate270
	}


        f1, err := os.Open(filename + ".out.fixed.gif")
        if err != nil {
            panic(err)
        }

	close, err := tk.PlayGIF(f1)
	fatal(err)

	time.Sleep(time.Second * 30)
	close <- true

        defer f1.Close()

}

func ProcessImage(img image.Image) image.Image {
	return resize.Resize(uint((*rows)*(*parallel)), uint((*rows)*(*chain)), img, resize.NearestNeighbor)
}

func ImageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {

	flag.Parse()
}
