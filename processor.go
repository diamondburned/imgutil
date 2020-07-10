package imgutil

import (
	"image"
	"image/draw"

	"github.com/diamondburned/imgutil/circle"
	"github.com/disintegration/imaging"
)

type Processor func(image.Image) *image.NRGBA

// Resize uses MaxSize to calculate and resize the image accordingly.
func Resize(maxW, maxH int) Processor {
	return func(img image.Image) *image.NRGBA {
		bounds := img.Bounds()
		imgW, imgH := bounds.Dx(), bounds.Dy()

		w, h := MaxSize(imgW, imgH, maxW, maxH)

		return imaging.Resize(img, w, h, imaging.Lanczos)
	}
}

// Round renders an anti-aliased round image. This image crops the source and
// makes it a square.
func Round(antialias bool) Processor {
	return func(img image.Image) *image.NRGBA {
		// Scale up
		const scale = 2

		// only bother anti-aliasing if it's not a paletted image.
		var _, paletted = img.(*image.Paletted)
		antialias = !paletted && antialias

		// Get the min dimensions.
		var mind = img.Bounds().Dx()
		if y := img.Bounds().Dy(); y < mind {
			mind = y
		}

		// Crop the image to be a square.
		img = imaging.CropAnchor(img, mind, mind, imaging.Top)

		if antialias {
			mind *= scale
			img = imaging.Resize(img, mind, mind, imaging.Lanczos)
		}

		var dst = image.NewNRGBA(image.Rect(0, 0, mind, mind))

		// Actually do the round-corners stuff.
		roundTo(img, dst, mind/2) // radius

		// Return the original image without downscaling if it's not
		// anti-aliased.
		if !antialias {
			return dst
		}

		// Get the original size.
		mind /= scale
		return imaging.Resize(dst, mind, mind, imaging.Lanczos)
	}
}

// roundTo round-crops an image without anti-aliasing.
func roundTo(src image.Image, dst draw.Image, r int) {
	draw.DrawMask(
		dst,
		src.Bounds(),
		src,
		image.ZP,
		circle.New(r),
		image.ZP,
		draw.Src,
	)
}

// MaxSize returns the maximum size that can fit within the given max width and
// height. Aspect ratio is preserved.
func MaxSize(w, h, maxW, maxH int) (int, int) {
	if w < maxW && h < maxH {
		return w, h
	}

	if w > h {
		h = h * maxW / w
		w = maxW
	} else {
		w = w * maxH / h
		h = maxH
	}

	return w, h
}
