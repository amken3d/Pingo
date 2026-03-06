package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	// Read SVG from the assets dir
	data, err := os.ReadFile("assets/pico.svg")
	if err != nil {
		fmt.Println("Cannot read SVG:", err)
		os.Exit(1)
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		fmt.Println("SVG parse error:", err)
		os.Exit(1)
	}

	w, h := int(icon.ViewBox.W*6), int(icon.ViewBox.H*6)
	icon.SetTarget(0, 0, float64(w), float64(h))

	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{C: color.NRGBA{R: 255, G: 255, B: 255, A: 255}}, image.Point{}, draw.Src)

	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)
	fmt.Printf("SVG rasterized: %dx%d\n", w, h)

	go func() {
		win := new(app.Window)
		win.Option(app.Title("Image Test"), app.Size(unit.Dp(800), unit.Dp(600)))
		if err := loop(win, rgba); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(win *app.Window, img *image.RGBA) error {
	imgOp := paint.NewImageOp(img)
	gioImg := widget.Image{
		Src:      imgOp,
		Fit:      widget.Contain,
		Position: layout.Center,
	}

	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Red background so we can see rendering works
			cl := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			paint.ColorOp{Color: color.NRGBA{R: 200, G: 50, B: 50, A: 255}}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			cl.Pop()

			// Display the image
			layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(600), gtx.Dp(240)))
				return gioImg.Layout(gtx)
			})

			e.Frame(gtx.Ops)
		}
	}
}
