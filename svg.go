// ┌──────────────────────────────────────────────────────────────────────┐
// │  svg.go — Embedding & Rendering SVG Assets                         │
// │                                                                    │
// │  Demonstrates:                                                     │
// │  • Go 1.16+ embed directive for bundling assets into the binary    │
// │  • Using oksvg + rasterx to rasterise SVG → image.NRGBA           │
// │  • Displaying raster images via ui.Image(img).Size(w, h)          │
// │                                                                    │
// │  ImmyGo's ui.Image takes any image.Image and renders it via Gio's │
// │  GPU pipeline. Since Gio doesn't have native SVG support, we      │
// │  rasterise the SVG once at startup and cache the result.           │
// │                                                                    │
// │  The Pico board SVG comes from:                                    │
// │  https://github.com/tinygo-org/playground/tree/main/parts          │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/amken3d/Pingo/pindata"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Embed all SVG assets into the binary at compile time.
//
//go:embed assets/pico.svg assets/qfn56.svg assets/qfn60.svg assets/qfn80.svg
var assetsFS embed.FS

// chipImages maps Board variants to their rasterised images.
// Loaded once at startup by loadAllSVGs().
var chipImages map[pindata.Board]image.Image

// loadAllSVGs rasterises all embedded SVG assets and stores them
// in chipImages for lookup by Board type.
// Bare chip variants (QFN) are rendered dynamically — no SVG needed.
func loadAllSVGs() {
	boardImg := rasterizeSVG("assets/pico.svg", true)
	chipImages = map[pindata.Board]image.Image{
		pindata.Pico:  boardImg,
		pindata.Pico2: boardImg,
	}
}

// boardImage returns the rasterised image for the current board/chip.
// Returns nil for bare chip variants (rendered dynamically).
func boardImage(b pindata.Board) image.Image {
	if img, ok := chipImages[b]; ok {
		return img
	}
	return nil
}

// rasterizeSVG reads an embedded SVG, rasterises it at 6× scale,
// and optionally rotates 90° CW (for landscape board SVGs).
func rasterizeSVG(path string, rotate bool) image.Image {
	data, err := assetsFS.ReadFile(path)
	if err != nil {
		fmt.Println("SVG load error:", path, err)
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		fmt.Println("SVG parse error:", path, err)
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}

	// Scale up 6× for crisp rendering on HiDPI displays.
	w, h := int(icon.ViewBox.W*6), int(icon.ViewBox.H*6)
	icon.SetTarget(0, 0, float64(w), float64(h))

	// Create the target canvas with a white background.
	rgba := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{
		C: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
	}, image.Point{}, draw.Src)

	// Rasterise SVG vector paths onto the canvas.
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)

	if rotate {
		return rotateCW90(rgba)
	}
	return rgba
}

// rotateCW90 rotates an NRGBA image 90° clockwise.
// Source (x, y) in WxH → destination (H-1-y, x) in HxW.
func rotateCW90(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, srcH, srcW))
	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			dst.SetNRGBA(srcH-1-y, x, src.NRGBAAt(x, y))
		}
	}
	return dst
}
