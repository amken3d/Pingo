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

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Embed the Pico board SVG into the binary at compile time.
// The //go:embed directive makes the file available without reading
// from disk at runtime — the binary is self-contained.
//
//go:embed assets/pico.svg
var assetsFS embed.FS

// loadPicoSVG reads the embedded SVG and rasterises it to a Go image.
//
// Steps:
//  1. Read the SVG bytes from the embedded filesystem.
//  2. Parse the SVG with oksvg.ReadIconStream.
//  3. Create an NRGBA canvas at 6× the SVG's native size for crisp rendering.
//  4. Rasterise the SVG paths onto the canvas with rasterx.
//  5. Return the image.Image for use with ui.Image().
func loadPicoSVG() image.Image {
	data, err := assetsFS.ReadFile("assets/pico.svg")
	if err != nil {
		fmt.Println("SVG load error:", err)
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		fmt.Println("SVG parse error:", err)
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

	// Rotate 90° clockwise so the board is vertical with USB at top.
	return rotateCW90(rgba)
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
