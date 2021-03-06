package ippresize

//go:generate stringer -type=Interpolation -trimprefix Interpolation

/*
#include "image.h"
#cgo pkg-config: libippi
*/
import "C"

import (
	"fmt"
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

type Error struct {
	error string
	code  int
}

func NewError(code int, format string, args ...interface{}) *Error {
	return &Error{code: code, error: fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	return e.error
}

func (e *Error) Code() IppStatus {
	return IppStatus(e.code)
}

func Decode(reader io.Reader, colorspace jpeg.OutColorSpace, bbox image.Point) (image.Image, error) {
	decoderOptions := jpeg.DecoderOptions{
		OutColorSpace: colorspace,
		ScaleTarget:   image.Rectangle{Max: bbox},
		DCTMethod:     jpeg.DCTIFast,
	}
	return jpeg.Decode(reader, &decoderOptions)
}

func Resize(in []uint8, in_stride int, in_size image.Point, out []uint8, out_stride int, out_size image.Point, channels int, interpolation Interpolation) error {

	if in_size.X <= 0 || in_size.Y <= 0 {
		return NewError(0, "one of the input image dimensions is invalid: {width: %v, height: %v}", in_size.X, in_size.Y)
	}

	if out_size.X <= 0 || out_size.Y <= 0 {
		return NewError(0, "one of the output image dimensions is invalid: {width: %v, height: %v}", out_size.X, out_size.Y)
	}

	if len(in) < channels*in_size.X*in_size.Y {
		return NewError(0, "input image buffer size doesn't match image dimensions: {width: %v, height: %v, channels: %v}, len=%v",
			in_size.X, in_size.Y, channels, len(in))
	}

	if len(out) < channels*out_size.X*out_size.Y {
		return NewError(0, "output image buffer size doesn't match image dimensions: {width: %v, height: %v, channels: %v}, len=%v",
			out_size.X, out_size.Y, channels, len(out))
	}

	var img_in C.struct_image_s
	img_in.w = C.uint(in_size.X)
	img_in.h = C.uint(in_size.Y)
	img_in.channels = C.uint(channels)
	img_in.rowstep = C.ulong(in_stride)
	img_in_data := (*C.uchar)(unsafe.Pointer(&in[0]))

	var img_out C.struct_image_s
	img_out.w = C.uint(out_size.X)
	img_out.h = C.uint(out_size.Y)
	img_out.channels = C.uint(channels)
	img_out.rowstep = C.ulong(out_stride)
	img_out_data := (*C.uchar)(unsafe.Pointer(&out[0]))

	const err_size = 1024
	var err [err_size]C.char

	ret := C.image_ipp_resize(&img_in, img_in_data, &img_out, img_out_data, C.image_interpolation_t(interpolation), &err[0], err_size)

	/* make 100% sure garbage collector wont kill these objects in the middle of execution of c function */
	runtime.KeepAlive(img_in)
	runtime.KeepAlive(img_in_data)
	runtime.KeepAlive(img_out)
	runtime.KeepAlive(img_out_data)

	if ret != 0 {
		return NewError(int(ret), "C.image_ipp_resize() failed: %v", C.GoString(&err[0]))
	}

	return nil
}

func ReplicateBorder(in []uint8, in_stride int, in_size image.Point, channels int, src image.Rectangle) error {

	if in_size.X <= 0 || in_size.Y <= 0 {
		return NewError(0, "one of the input image dimensions is invalid: {width: %v, height: %v}", in_size.X, in_size.Y)
	}

	if len(in) < channels*in_size.X*in_size.Y {
		return NewError(0, "input image buffer size doesn't match image dimensions: {width: %v, height: %v, channels: %v}, len=%v",
			in_size.X, in_size.Y, channels, len(in))
	}

	var img C.struct_image_s
	img.w = C.uint(in_size.X)
	img.h = C.uint(in_size.Y)
	img.channels = C.uint(channels)
	img.rowstep = C.size_t(in_stride)
	img_data := (*C.uchar)(unsafe.Pointer(&in[0]))

	const err_size = 1024
	var err [err_size]C.char

	ret := C.image_ipp_replicate_border_inplace(&img, img_data, C.uint(src.Min.X), C.uint(src.Min.Y), C.uint(src.Dx()), C.uint(src.Dy()), &err[0], err_size)

	/* make 100% sure garbage collector wont kill these objects in the middle of execution of c function */
	runtime.KeepAlive(img)
	runtime.KeepAlive(img_data)

	if ret != 0 {
		return NewError(int(ret), "C.image_ipp_replicate_border_inplace() failed: %v", C.GoString(&err[0]))
	}

	return nil
}

func GetProportionalLargestInnerSize(in_size image.Point, box image.Point) image.Point {
	w, h := box.X, box.Y

	if float64(w)/float64(in_size.X) < float64(h)/float64(in_size.Y) {
		h = int(math.Round(float64(in_size.Y) * float64(w) / float64(in_size.X)))
		if h == 0 {
			h = 1
		}
	} else {
		w = int(math.Round(float64(in_size.X) * float64(h) / float64(in_size.Y)))
		if w == 0 {
			w = 1
		}
	}

	return image.Point{X: w, Y: h}
}

func GetProportionalSmallestOuterSize(in_size image.Point, box image.Point) image.Point {
	w, h := box.X, box.Y

	if float64(w)/float64(in_size.X) > float64(h)/float64(in_size.Y) {
		h = int(math.Round(float64(in_size.Y) * float64(w) / float64(in_size.X)))
		if h == 0 {
			h = 1
		}
	} else {
		w = int(math.Round(float64(in_size.X) * float64(h) / float64(in_size.Y)))
		if w == 0 {
			w = 1
		}
	}

	return image.Point{X: w, Y: h}
}

func ResizeProportional(in []uint8, in_stride int, in_size image.Point, channels int, out_size_box image.Point, interpolation Interpolation) ([]uint8, image.Point, error) {
	out_size := GetProportionalLargestInnerSize(in_size, out_size_box)
	out := make([]uint8, channels*out_size.X*out_size.Y)
	out_rowstep := channels * out_size.X
	err := Resize(in, in_stride, in_size, out, out_rowstep, out_size, channels, interpolation)
	return out, out_size, err
}

func ResizePadGray(in []uint8, in_stride int, in_size image.Point, channels int, out_size_box image.Point, interpolation Interpolation) ([]uint8, image.Point, error) {
	out_size := out_size_box
	target_out_size := GetProportionalLargestInnerSize(in_size, out_size_box)
	out := make([]uint8, channels*out_size_box.X*out_size_box.Y)
	if target_out_size.X != out_size.X || target_out_size.Y != out_size.Y {
		for i := range out {
			out[i] = 128
		}
	}
	out_rowstep := channels * out_size.X
	target_offset := channels*int((out_size.X-target_out_size.X)/2) + out_rowstep*int((out_size.Y-target_out_size.Y)/2)
	err := Resize(in, in_stride, in_size, out[target_offset:], out_rowstep, target_out_size, channels, interpolation)
	return out, out_size, err
}

func JpegToRGBA(reader io.Reader, bbox image.Point, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceRGBA, bbox)
	if err != nil {
		return
	}

	im_rgba := im.(*image.RGBA)
	pixdata, size, err = ResizeProportional(im_rgba.Pix, im_rgba.Stride, im_rgba.Bounds().Max, 4, bbox, interpolation)
	return
}

func JpegToRGB(reader io.Reader, bbox image.Point, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceRGB, bbox)
	if err != nil {
		return
	}

	im_rgb := im.(*rgb.Image)
	pixdata, size, err = ResizeProportional(im_rgb.Pix, im_rgb.Stride, im_rgb.Bounds().Max, 3, bbox, interpolation)
	return
}

func JpegToGray(reader io.Reader, bbox image.Point, interpolation Interpolation) (pixdata []uint8, size image.Point, err error) {
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceGray, bbox)
	if err != nil {
		return
	}

	im_gray := im.(*image.Gray)
	pixdata, size, err = ResizeProportional(im_gray.Pix, im_gray.Stride, im_gray.Bounds().Max, 1, bbox, interpolation)
	return
}

func JpegToSquareRGBA(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceRGBA, bbox)
	if err != nil {
		return
	}

	im_rgba := im.(*image.RGBA)
	pixdata, _, err = ResizePadGray(im_rgba.Pix, im_rgba.Stride, im_rgba.Bounds().Max, 4, bbox, interpolation)
	return
}

func JpegToSquareRGB(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceRGB, bbox)
	if err != nil {
		return
	}

	im_rgb := im.(*rgb.Image)
	pixdata, _, err = ResizePadGray(im_rgb.Pix, im_rgb.Stride, im_rgb.Bounds().Max, 3, bbox, interpolation)
	return
}

func JpegToSquareGray(reader io.Reader, sqsize int, interpolation Interpolation) (pixdata []uint8, err error) {
	bbox := image.Point{sqsize, sqsize}
	var im image.Image
	im, err = Decode(reader, jpeg.OutColorSpaceGray, bbox)
	if err != nil {
		return
	}

	im_gray := im.(*image.Gray)
	pixdata, _, err = ResizePadGray(im_gray.Pix, im_gray.Stride, im_gray.Bounds().Max, 1, bbox, interpolation)
	return
}

func JpegToRGBAImage(reader io.Reader, bbox image.Point, interpolation Interpolation) (im image.Image, err error) {
	var pixdata []uint8
	var size image.Point
	pixdata, size, err = JpegToRGBA(reader, bbox, interpolation)
	if err == nil {
		im = &image.RGBA{
			Pix:    pixdata,
			Stride: size.X * 4,
			Rect:   image.Rectangle{Max: size},
		}
	}
	return
}

func JpegToRGBImage(reader io.Reader, bbox image.Point, interpolation Interpolation) (im image.Image, err error) {
	var pixdata []uint8
	var size image.Point
	pixdata, size, err = JpegToRGB(reader, bbox, interpolation)
	if err == nil {
		im = &rgb.Image{
			Pix:    pixdata,
			Stride: size.X * 3,
			Rect:   image.Rectangle{Max: size},
		}
	}
	return
}

func JpegToGrayImage(reader io.Reader, bbox image.Point, interpolation Interpolation) (im image.Image, err error) {
	var pixdata []uint8
	var size image.Point
	pixdata, size, err = JpegToGray(reader, bbox, interpolation)
	if err == nil {
		im = &image.Gray{
			Pix:    pixdata,
			Stride: size.X * 1,
			Rect:   image.Rectangle{Max: size},
		}
	}
	return
}

func JpegToImage(reader io.Reader, bbox image.Point, interpolation Interpolation) (im image.Image, err error) {
	im, err = Decode(reader, jpeg.OutColorSpaceSame, bbox)
	if err != nil {
		return
	}

	size := GetProportionalLargestInnerSize(im.Bounds().Max, bbox)

	switch i := im.(type) {
	case *image.Gray:
		im, err = ResizeGray(i, size, interpolation)
	case *image.YCbCr:
		im, err = ResizeLimitedYCbCr(i, size, interpolation)
	default:
		err = NewError(0, "unsupported color model")
	}

	return
}

func ResizeGray(gray *image.Gray, size image.Point, interpolation Interpolation) (resized *image.Gray, err error) {
	resized = image.NewGray(image.Rectangle{Max: size})
	err = Resize(gray.Pix, gray.Stride, image.Point{gray.Bounds().Dx(), gray.Bounds().Dy()}, resized.Pix, resized.Stride, resized.Rect.Max, 1, interpolation)
	return
}

func ResizeLimitedYCbCr(ycbcr *image.YCbCr, size image.Point, interpolation Interpolation) (resized *image.YCbCr, err error) {

	// IPP has no support for images with Y, Cb and Cr separate planes which is a standard golang representation
	// of the most common jpeg image format so we have to resize each plane individually

	var downresW, downresH int

	switch ycbcr.SubsampleRatio {
	case image.YCbCrSubsampleRatio444:
		downresW, downresH = 1, 1
	case image.YCbCrSubsampleRatio422:
		downresW, downresH = 2, 1
	case image.YCbCrSubsampleRatio420:
		downresW, downresH = 2, 2
	case image.YCbCrSubsampleRatio440:
		downresW, downresH = 1, 2
	case image.YCbCrSubsampleRatio411:
		downresW, downresH = 4, 1
	case image.YCbCrSubsampleRatio410:
		downresW, downresH = 4, 2
	}

	// log.Printf("%v -> %v, %v (h%vv%v)", ycbcr.Rect.Max, size, ycbcr.SubsampleRatio, downresW, downresH)

	if ycbcr.Rect.Empty() {
		err = NewError(0, "Empty source image: %v", ycbcr.Rect)
		return
	}

	if ycbcr.Rect.Min.X&(downresW-1) > 0 ||
		ycbcr.Rect.Min.Y&(downresH-1) > 0 ||
		ycbcr.Rect.Max.X&(downresW-1) > 0 ||
		ycbcr.Rect.Max.Y&(downresH-1) > 0 {
		err = NewError(0, "Unaligned source image dimensions: %v, SubsampleRatio=%v", ycbcr.Rect, ycbcr.SubsampleRatio)
		return
	}

	if size.X&(downresW-1) > 0 || size.Y&(downresH-1) > 0 {
		err = NewError(0, "Unaligned destination image dimensions: %v, SubsampleRatio=%v", size, ycbcr.SubsampleRatio)
		return
	}

	resized = image.NewYCbCr(image.Rectangle{Max: size}, ycbcr.SubsampleRatio)

	err = Resize(ycbcr.Y, ycbcr.YStride, image.Point{ycbcr.Bounds().Dx(), ycbcr.Bounds().Dy()}, resized.Y, resized.YStride, size, 1, interpolation)

	if err != nil {
		return
	}

	err = Resize(ycbcr.Cb, ycbcr.CStride, image.Point{ycbcr.Bounds().Dx() / downresW, ycbcr.Bounds().Dy() / downresH}, resized.Cb, resized.CStride, image.Point{size.X / downresW, size.Y / downresH}, 1, interpolation)

	if err != nil {
		return
	}

	err = Resize(ycbcr.Cr, ycbcr.CStride, image.Point{ycbcr.Bounds().Dx() / downresW, ycbcr.Bounds().Dy() / downresH}, resized.Cr, resized.CStride, image.Point{size.X / downresW, size.Y / downresH}, 1, interpolation)

	return
}

func init() {
	C.image_init()
}
