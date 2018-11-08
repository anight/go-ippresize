package ippresize

/*
#include <string.h>
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
	InterpolationNearestNeighbour   Interpolation = C.IMAGE_INTERPOLATION_NN
	InterpolationLinear             Interpolation = C.IMAGE_INTERPOLATION_LINEAR
	InterpolationCubic              Interpolation = C.IMAGE_INTERPOLATION_CUBIC
	InterpolationLanczos            Interpolation = C.IMAGE_INTERPOLATION_LANCZOS
	InterpolationSuper              Interpolation = C.IMAGE_INTERPOLATION_SUPER
	InterpolationAntialiasingLinear Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_LINEAR
	InterpolationAntialiasingCubic  Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_CUBIC
	InterpolationAntiliasingLanczos Interpolation = C.IMAGE_INTERPOLATION_ANTIALIASING_LANCZOS
)

func decode(reader io.Reader, colorspace jpeg.OutColorSpace, sqsize int) (image.Image, error) {
	decoderOptions := jpeg.DecoderOptions{
		OutColorSpace: colorspace,
		ScaleTarget: image.Rectangle{
			image.Point{0, 0},
			image.Point{sqsize, sqsize},
		},
	}
	return jpeg.Decode(reader, &decoderOptions)
}

func Resize(in []uint8, in_stride int, in_w int, in_h int, sqsize int, channels int, interpolation Interpolation) []uint8 {

	pixdata := make([]uint8, channels*sqsize*sqsize)

	var img_in C.struct_image_s
	var img_out C.struct_image_s

	img_in.w = C.uint(in_w)
	img_in.h = C.uint(in_h)
	img_in.channels = C.uint(channels)
	img_in.rowstep = C.ulong(in_stride)

	img_in_data := (*C.uchar)(unsafe.Pointer(&in[0]))

	img_out.w = C.uint(sqsize)
	img_out.h = C.uint(sqsize)
	img_out.channels = img_in.channels
	img_out.rowstep = C.ulong(img_out.w * img_out.channels)

	img_out_data := (*C.uchar)(unsafe.Pointer(&pixdata[0]))

	w, h := sqsize, sqsize

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

	if w != sqsize || h != sqsize {
		C.memset(unsafe.Pointer(img_out_data), 128, C.ulong(img_out.w*img_out.h*img_out.channels))
	}

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

	return pixdata
}

func JpegToSquareRGB(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceRGB, sqsize)
	if err != nil {
		return
	}

	im_rgb := im.(*rgb.Image)
	pixdata = Resize(im_rgb.Pix, im_rgb.Stride, im_rgb.Bounds().Dx(), im_rgb.Bounds().Dy(), sqsize, 3, interpolation)
	return
}

func JpegToSquareGray(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	var im image.Image
	im, err = decode(reader, jpeg.OutColorSpaceGray, sqsize)
	if err != nil {
		return
	}

	im_gray := im.(*image.Gray)
	pixdata = Resize(im_gray.Pix, im_gray.Stride, im_gray.Bounds().Dx(), im_gray.Bounds().Dy(), sqsize, 1, interpolation)
	return
}

func init() {
	C.image_init()
}
