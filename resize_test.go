package ippresize

import (
	"fmt"
	"github.com/anight/go-libjpeg/rgb"
	"image"
	"image/png"
	"io"
	"os"
	"reflect"
	"testing"
)

func test_interpolation(t *testing.T, interpolation Interpolation, channels int, im_type reflect.Type, resize func(io.Reader, image.Point, bool, Interpolation) ([]uint8, image.Point, error)) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Fatalf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, im_size, err := resize(reader, image.Point{224, 224}, false, interpolation)
	if err != nil {
		t.Fatalf("%s() failed: %v", t.Name(), err)
	}
	if len(im_data) != im_size.X*im_size.Y*channels {
		t.Fatalf("Expected image %d * %d * %d = %d bytes, got %d bytes", im_size.X, im_size.Y, channels, im_size.X*im_size.Y*channels, len(im_data))
	}

	im := reflect.New(im_type)
	im.Elem().FieldByName("Pix").Set(reflect.ValueOf(im_data))
	im.Elem().FieldByName("Stride").SetInt(int64(im_size.X * channels))
	im.Elem().FieldByName("Rect").FieldByName("Max").FieldByName("X").SetInt(int64(im_size.X))
	im.Elem().FieldByName("Rect").FieldByName("Max").FieldByName("Y").SetInt(int64(im_size.Y))

	writer, err := os.Create(fmt.Sprintf("./test-%v-%v.png", t.Name(), interpolation))
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, im.Interface().(image.Image)); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}

func test(t *testing.T, channels int, im_type reflect.Type, resize func(io.Reader, image.Point, bool, Interpolation) ([]uint8, image.Point, error)) {
	for _, interpolation := range []Interpolation{
		InterpolationNearestNeighbour, InterpolationLinear, InterpolationCubic, InterpolationLanczos, InterpolationSuper,
		InterpolationAntialiasingLinear, InterpolationAntialiasingCubic, InterpolationAntialiasingLanczos} {
		test_interpolation(t, interpolation, channels, im_type, resize)
	}

}

func TestSquareRGBA(t *testing.T) {
	test(t, 4, reflect.TypeOf(image.RGBA{}), func(reader io.Reader, size image.Point, graypad bool, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareRGBA(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestSquareRGB(t *testing.T) {
	test(t, 3, reflect.TypeOf(rgb.Image{}), func(reader io.Reader, size image.Point, graypad bool, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareRGB(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestSquareGray(t *testing.T) {
	test(t, 1, reflect.TypeOf(image.Gray{}), func(reader io.Reader, size image.Point, graypad bool, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareGray(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestRGBA(t *testing.T) {
	test(t, 4, reflect.TypeOf(image.RGBA{}), JpegToRGBA)
}

func TestRGB(t *testing.T) {
	test(t, 3, reflect.TypeOf(rgb.Image{}), JpegToRGB)
}

func TestGray(t *testing.T) {
	test(t, 1, reflect.TypeOf(image.Gray{}), JpegToGray)
}
