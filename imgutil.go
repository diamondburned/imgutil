package imgutil

import (
	"image"
	"image/color"
	"image/draw"
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
	GIF, err := ProcessAnimation(src, processors)
	if err != nil {
		return err
	}

	if err := gif.EncodeAll(dst, GIF); err != nil {
		return errors.Wrap(err, "Failed to encode GIF")
	}

	return nil
}

// ProcessAnimation works similarly to Process, but parses a GIF.
func ProcessAnimation(src io.Reader, processors []Processor) (*gif.GIF, error) {
	GIF, err := gif.DecodeAll(src)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode GIF")
	}

	// Error if no frames.
	if len(GIF.Image) == 0 {
		return nil, errors.New("GIF has no frames.")
	}

	// Make a temporary frame to draw over.
	var lst *image.Paletted // latest frame to draw from

	// Encode the GIF frame-by-frame
	for _, frame := range GIF.Image {
		// Copy frame over to do postprocessing.
		var img = image.Image(frame)
		for _, proc := range processors {
			img = proc(img)
		}

		// Update bounds.
		frame.Rect = img.Bounds()

		// Copy the last frame to the gif frame, if available. Usually when it
		// is not, it means that we're in the first frame.
		if lst != nil {
			draw.Draw(frame, frame.Rect, lst, frame.Rect.Min, draw.Src)
			// Draw the processed image over the gif frame.
			draw.Draw(frame, frame.Rect, img, frame.Rect.Min, draw.Over)
		} else {
			// Completely override everything. This is done because just drawing
			// over the first frame will not apply alpha properly.
			draw.Draw(frame, frame.Rect, img, frame.Rect.Min, draw.Src)
		}

		// Assign this frame to the last frame.
		lst = frame
	}

	// Set the new bounds.
	bounds := GIF.Image[0].Bounds()
	GIF.Config.Width = bounds.Dx()
	GIF.Config.Height = bounds.Dy()

	return GIF, nil
}

// ProcessStream takes a processor and run them through the image decoded from
// the stream. The returned bytes are PNG-encoded and uncompressed.
func ProcessStream(dst io.Writer, src io.Reader, processors []Processor) error {
	img, err := Process(src, processors)
	if err != nil {
		return err
	}

	if err := PNGEncoder.Encode(dst, img); err != nil {
		return errors.Wrap(err, "Failed to encode")
	}

	return nil
}

// Process takes a processor and run them through the image decoded from the
// stream. The returned image is the processed image.
func Process(src io.Reader, processors []Processor) (image.Image, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode")
	}

	for _, proc := range processors {
		img = proc(img)
	}

	return img, nil

}

// Prepend prepends p1 before pN.
func Prepend(p1 Processor, pN []Processor) []Processor {
	return append([]Processor{p1}, pN...)
}
