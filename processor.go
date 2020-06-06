package imgutil

import (
	"image"
	"image/draw"

	"github.com/diamondburned/imgutil/circle"
	"github.com/disintegration/imaging"
)

type Processor func(image.Image) image.Image

// Resize uses MaxSize to calculate and resize the image accordingly.
func Resize(maxW, maxH int) Processor {
	return func(img image.Image) image.Image {
		bounds := img.Bounds()
		imgW, imgH := bounds.Dx(), bounds.Dy()

		w, h := MaxSize(imgW, imgH, maxW, maxH)

		return imaging.Resize(img, w, h, imaging.Lanczos)
	}
}

// Round renders an anti-aliased round image.
func Round() Processor {
	// for documentation purposes
	return round
}

func round(img image.Image) image.Image {
	// Scale up
	oldbounds := img.Bounds()
	const scale = 2

	// only bother anti-aliasing if it's not a paletted image.
	var _, paletted = img.(*image.Paletted)
	if !paletted {
		img = imaging.Resize(img, oldbounds.Dx()*scale, oldbounds.Dy()*scale, imaging.Lanczos)
	}

	r := img.Bounds().Dx() / 2

	var dst draw.Image

	switch img.(type) {
	// alpha-supported:
	case *image.RGBA, *image.RGBA64, *image.NRGBA, *image.NRGBA64:
		dst = img.(draw.Image)
	default:
		dst = image.NewRGBA(image.Rect(
			0, 0,
			r*2, r*2,
		))
	}

	roundTo(img, dst, r)

	if paletted {
		return dst
	}

	return imaging.Resize(dst, oldbounds.Dx(), oldbounds.Dy(), imaging.Lanczos)
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
