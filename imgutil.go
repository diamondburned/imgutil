package imgutil

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"

	_ "image/jpeg"

	"github.com/pkg/errors"
)

func ensurePaletteTransparent(palette color.Palette) color.Palette {
	// If we already have a transparent color, then we don't need to add any
	// more.
	for _, c := range palette {
		if c == color.Transparent {
			return palette
		}
	}

	// TODO: properly quantize
	if len(palette) > 255 {
		palette = palette[:255]
	}
	return append(palette, color.Transparent)
}

var PNGEncoder = &png.Encoder{
	// Prefer speed over compression, since cache is slightly more optimized
	// now.
	CompressionLevel: png.NoCompression,
}

// ProcessAnimationStream works similarly to ProcessStream, but parses a GIF.
func ProcessAnimationStream(dst io.Writer, src io.Reader, processors []Processor) error {
	GIF, err := gif.DecodeAll(src)
	if err != nil {
		return errors.Wrap(err, "Failed to decode GIF")
	}

	// Add transparency:
	if p, ok := GIF.Config.ColorModel.(color.Palette); ok {
		GIF.Config.ColorModel = ensurePaletteTransparent(p)
	}

	// Encode the GIF frame-by-frame
	for _, frame := range GIF.Image {
		var img = image.Image(frame)
		for _, proc := range processors {
			img = proc(img)
		}

		frame.Rect = img.Bounds()

		if frame.Palette != nil {
			frame.Palette = ensurePaletteTransparent(frame.Palette)
		}

		for x := 0; x < frame.Rect.Dx(); x++ {
			for y := 0; y < frame.Rect.Dy(); y++ {
				frame.Set(x, y, img.At(x, y))
			}
		}
	}

	if len(GIF.Image) > 0 {
		bounds := GIF.Image[0].Bounds()
		GIF.Config.Width = bounds.Dx()
		GIF.Config.Height = bounds.Dy()
	}

	if err := gif.EncodeAll(dst, GIF); err != nil {
		return errors.Wrap(err, "Failed to encode GIF")
	}

	return nil
}

// ProcessStream takes a processor and run them through the image decoded from
// the stream. The returned bytes are PNG-encoded and uncompressed.
func ProcessStream(dst io.Writer, src io.Reader, processors []Processor) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return errors.Wrap(err, "Failed to decode")
	}

	for _, proc := range processors {
		img = proc(img)
	}

	if err := PNGEncoder.Encode(dst, img); err != nil {
		return errors.Wrap(err, "Failed to encode")
	}

	return nil
}

// Prepend prepends p1 before pN.
func Prepend(p1 Processor, pN []Processor) []Processor {
	return append([]Processor{p1}, pN...)
}
