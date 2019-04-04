package ippresize

import (
	"bytes"
	"github.com/anight/go-libjpeg/rgb"
	"image"
	"image/png"
	"io"
	"os"
	"reflect"
	"testing"
)

var allInterpolations = [...]Interpolation{
	InterpolationNearestNeighbour,
	InterpolationLinear,
	InterpolationCubic,
	InterpolationLanczos,
	InterpolationSuper,
	InterpolationAntialiasingLinear,
	InterpolationAntialiasingCubic,
	InterpolationAntialiasingLanczos,
}

func testResizeInterpolation(t *testing.T, interpolation Interpolation, channels int, im_type reflect.Type, resize func(io.Reader, image.Point, Interpolation) ([]uint8, image.Point, error)) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Fatalf("os.Open() failed: %v", err)
	}
	defer reader.Close()

	im_data, im_size, err := resize(reader, image.Point{224, 224}, interpolation)
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

	writer, err := os.OpenFile("/dev/null", os.O_WRONLY, 0)
	//	writer, err := os.Create(fmt.Sprintf("./test-%v-%v.png", t.Name(), interpolation))
	if err != nil {
		t.Fatalf("os.OpenFile() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, im.Interface().(image.Image)); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}

func testResize(t *testing.T, channels int, im_type reflect.Type, resize func(io.Reader, image.Point, Interpolation) ([]uint8, image.Point, error)) {
	for _, interpolation := range allInterpolations {
		testResizeInterpolation(t, interpolation, channels, im_type, resize)
	}
}

func TestSquareRGBA(t *testing.T) {
	testResize(t, 4, reflect.TypeOf(image.RGBA{}), func(reader io.Reader, size image.Point, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareRGBA(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestSquareRGB(t *testing.T) {
	testResize(t, 3, reflect.TypeOf(rgb.Image{}), func(reader io.Reader, size image.Point, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareRGB(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestSquareGray(t *testing.T) {
	testResize(t, 1, reflect.TypeOf(image.Gray{}), func(reader io.Reader, size image.Point, interpolation Interpolation) ([]uint8, image.Point, error) {
		im_data, err := JpegToSquareGray(reader, size.X, interpolation)
		im_size := image.Point{size.X, size.X}
		return im_data, im_size, err
	})
}

func TestRGBA(t *testing.T) {
	testResize(t, 4, reflect.TypeOf(image.RGBA{}), JpegToRGBA)
}

func TestRGB(t *testing.T) {
	testResize(t, 3, reflect.TypeOf(rgb.Image{}), JpegToRGB)
}

func TestGray(t *testing.T) {
	testResize(t, 1, reflect.TypeOf(image.Gray{}), JpegToGray)
}

func TestReplicateBorder(t *testing.T) {
	pix := []uint8{
		0, 0, 0, 0, 0, 0,
		0, 1, 2, 3, 4, 0,
		0, 5, 6, 7, 8, 0,
		0, 0, 0, 0, 0, 0,
	}
	expected_pix := []uint8{
		1, 1, 2, 3, 4, 4,
		1, 1, 2, 3, 4, 4,
		5, 5, 6, 7, 8, 8,
		5, 5, 6, 7, 8, 8,
	}
	im := image.NewGray(image.Rect(0, 0, 6, 4))
	im.Pix = pix
	ReplicateBorder(im.Pix, im.Stride, im.Rect.Max, 1, image.Rect(1, 1, 5, 3))
	if !bytes.Equal(im.Pix, expected_pix) {
		t.Fatalf("expected %v, got %v", expected_pix, im.Pix)
	}
}

type proportionsTest struct {
	f                func(image.Point, image.Point) image.Point
	im, to, expected image.Point
}

func TestProportions(t *testing.T) {
	test := []proportionsTest{
		{GetProportionalLargestInnerSize, image.Point{640, 640}, image.Point{160, 160}, image.Point{160, 160}},
		{GetProportionalLargestInnerSize, image.Point{480, 640}, image.Point{160, 160}, image.Point{120, 160}},
		{GetProportionalLargestInnerSize, image.Point{640, 480}, image.Point{160, 160}, image.Point{160, 120}},
		{GetProportionalSmallestOuterSize, image.Point{640, 640}, image.Point{160, 160}, image.Point{160, 160}},
		{GetProportionalSmallestOuterSize, image.Point{480, 640}, image.Point{160, 160}, image.Point{160, 213}},
		{GetProportionalSmallestOuterSize, image.Point{640, 480}, image.Point{160, 160}, image.Point{213, 160}},
	}
	for _, item := range test {
		result := item.f(item.im, item.to)
		if result.X != item.expected.X || result.Y != item.expected.Y {
			t.Errorf("result: %v, expected: %v, (%v, %v)", result, item.expected, item.im, item.to)
		}
	}
}

func ippCanHandle(interpolation Interpolation, inSize, outSize int) bool {
	switch interpolation {
	case InterpolationLinear, InterpolationAntialiasingLinear:
		if inSize > 1 {
			return true
		} else {
			return false
		}
	case InterpolationCubic, InterpolationAntialiasingCubic, InterpolationAntialiasingLanczos:
		if inSize > 4 {
			return true
		} else {
			return false
		}
	case InterpolationLanczos:
		if inSize > 5 {
			return true
		} else {
			return false
		}
	case InterpolationSuper:
		if inSize >= outSize {
			return true
		} else {
			return false
		}
	}

	// by default assume ipp can handle anything
	return true
}

func TestIppResizeEdgeCases(t *testing.T) {
	const outSize = 100

	for _, interpolation := range allInterpolations {
		for inputSize := 1; inputSize < 200; inputSize++ {
			for _, channels := range [...]int{1, 3, 4} {
				in := make([]uint8, inputSize*inputSize*channels)
				in_stride := inputSize * channels
				in_size := image.Point{inputSize, inputSize}
				out := make([]uint8, outSize*outSize*channels)
				out_stride := outSize * channels
				out_size := image.Point{outSize, outSize}
				err := Resize(in, in_stride, in_size, out, out_stride, out_size, channels, interpolation)
				if err != nil {
					specificErr, ok := err.(*Error)
					if !ok {
						t.Fatalf("oops, %#v", err)
					}
					//					t.Logf("%v %v %v -> %v", interpolation, inputSize, outSize, ippCanHandle(interpolation, inputSize, outSize))
					if specificErr.Code() == IppStsSizeErr && !ippCanHandle(interpolation, inputSize, outSize) {
						continue
					}
					t.Errorf("%v: size=%v, channels=%v, code=%v", interpolation, inputSize, channels, specificErr.Code())
				}
			}
		}
	}
}
