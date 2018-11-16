package ippresize

/*
#include "image.h"
#cgo pkg-config: ippi
*/
import "C"

import (
	"github.com/anight/go-libjpeg/jpeg"
	"github.com/anight/go-libjpeg/rgb"
	"image"
	"io"
	"math"
	"runtime"
	"unsafe"
)

type Interpolation C.image_interpolation_t

const (
	InterpolationNearestNeighbour    Interpolation = C.IMAGE_INTERPOLATION_NN
	InterpolationLinear              Interpolation = C.IMAGE_INTERPOLATION_LINEAR
	InterpolationCubic               Interpolation = C.IMAGE_INTERPOLATION_CUBIC
	InterpolationLanczos             Interpolation = C.IMAGE_INTERPOLATION_LANCZOS
	InterpolationSuper               Interpolation = C.IMAGE_INTERPOLATION_SUPER
	InterpolationAntialiasingLinear  Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_LINEAR
	InterpolationAntialiasingCubic   Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_CUBIC
	InterpolationAntialiasingLanczos Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_LANCZOS
)

func (i Interpolation) String() string {
	switch i {
	case InterpolationNearestNeighbour:
		return "NearestNeighbour"
	case InterpolationLinear:
		return "Linear"
	case InterpolationCubic:
		return "Cubic"
	case InterpolationLanczos:
		return "Lanczos"
	case InterpolationSuper:
		return "Super"
	case InterpolationAntialiasingLinear:
		return "AntialiasingLinear"
	case InterpolationAntialiasingCubic:
		return "AntialiasingCubic"
	case InterpolationAntialiasingLanczos:
		return "AntialiasingLanczos"
	}
	return ""
}

func decode(reader io.Reader, colorspace jpeg.OutColorSpace, bbox image.Point) (image.Image, error) {
	decoderOptions := jpeg.DecoderOptions{
		OutColorSpace: colorspace,
		ScaleTarget:   image.Rectangle{Max: bbox},
	}
	return jpeg.Decode(reader, &decoderOptions)
}

func Resize(in []uint8, in_stride int, in_size image.Point, channels int, out_size image.Point, graypad bool, interpolation Interpolation) ([]uint8, image.Point) {

	var img_in C.struct_image_s

	img_in.w = C.uint(in_size.X)
	img_in.h = C.uint(in_size.Y)
	img_in.channels = C.uint(channels)
	img_in.rowstep = C.ulong(in_stride)
	img_in_data := (*C.uchar)(unsafe.Pointer(&in[0]))

	w, h := out_size.X, out_size.Y

	if float64(w)/float64(img_in.w) < float64(h)/float64(img_in.h) {
		h = int(math.Round(float64(img_in.h) * float64(w) / float64(img_in.w)))
		if h == 0 {
			h = 1
		}
	} else {
		w = int(math.Round(float64(img_in.w) * float64(h) / float64(img_in.h)))
		if w == 0 {
			w = 1
		}
	}

	var img_out C.struct_image_s
	var pixdata []uint8

	if graypad {
		img_out.w = C.uint(out_size.X)
		img_out.h = C.uint(out_size.Y)
		pixdata = make([]uint8, channels*out_size.X*out_size.Y)
		if w != out_size.X || h != out_size.Y {
			for i := range pixdata {
				pixdata[i] = 128
			}
		}
	} else {
		img_out.w = C.uint(w)
		img_out.h = C.uint(h)
		pixdata = make([]uint8, channels*w*h)
	}

	img_out.channels = img_in.channels
	img_out.rowstep = C.ulong(img_out.w * img_out.channels)

	img_scaled := img_out

	img_scaled.w = C.uint(w)
	img_scaled.h = C.uint(h)
	offset := int(img_out.channels*((img_out.w-img_scaled.w)/2)) + int(img_out.rowstep)*int((img_out.h-img_scaled.h)/2)

	img_scaled_data := (*C.uchar)(unsafe.Pointer(&pixdata[offset]))

	if 0 > C.image_ipp_resize(&img_in, img_in_data, &img_scaled, img_scaled_data, C.image_interpolation_t(interpolation)) {
		panic("C.image_ipp_resize() failed")
	}

	/* make 100% sure garbage collector wont kill these objects in the middle of execution of image_ipp_resize() */
	runtime.KeepAlive(img_in)
	runtime.KeepAlive(img_in_data)
	runtime.KeepAlive(img_scaled)
	runtime.KeepAlive(img_scaled_data)

	return pixdata, image.Point{int(img_out.w), int(img_out.h)}
}

func JpegToRGBA(reader io.Reader, bbox image.Point, graypad bool, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceRGBA, bbox)
	if err != nil {
		return
	}

	im_rgba := im.(*image.RGBA)
	pixdata, size = Resize(im_rgba.Pix, im_rgba.Stride, im_rgba.Bounds().Max, 4, bbox, graypad, interpolation)
	return
}

func JpegToRGB(reader io.Reader, bbox image.Point, graypad bool, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceRGB, bbox)
	if err != nil {
		return
	}

	im_rgb := im.(*rgb.Image)
	pixdata, size = Resize(im_rgb.Pix, im_rgb.Stride, im_rgb.Bounds().Max, 3, bbox, graypad, interpolation)
	return
}

func JpegToGray(reader io.Reader, bbox image.Point, graypad bool, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceGray, bbox)
	if err != nil {
		return
	}

	im_gray := im.(*image.Gray)
	pixdata, size = Resize(im_gray.Pix, im_gray.Stride, im_gray.Bounds().Max, 1, bbox, graypad, interpolation)
	return
}

func JpegToSquareRGBA(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceRGBA, bbox)
	if err != nil {
		return
	}

	im_rgba := im.(*image.RGBA)
	pixdata, _ = Resize(im_rgba.Pix, im_rgba.Stride, im_rgba.Bounds().Max, 4, bbox, true, interpolation)
	return
}

func JpegToSquareRGB(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceRGB, bbox)
	if err != nil {
		return
	}

	im_rgb := im.(*rgb.Image)
	pixdata, _ = Resize(im_rgb.Pix, im_rgb.Stride, im_rgb.Bounds().Max, 3, bbox, true, interpolation)
	return
}

func JpegToSquareGray(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceGray, bbox)
	if err != nil {
		return
	}

	im_gray := im.(*image.Gray)
	pixdata, _ = Resize(im_gray.Pix, im_gray.Stride, im_gray.Bounds().Max, 1, bbox, true, interpolation)
	return
}

func init() {
	C.image_init()
}
