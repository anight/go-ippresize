package ippresize

import (
	"github.com/anight/go-libjpeg/rgb"
	"image"
	"image/png"
	"os"
	"testing"
)

func TestRGB(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Errorf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, err := JpegToSquareRGB(reader, 224, InterpolationLanczos)
	if err != nil {
		t.Errorf("JpegToSquareRGB() failed: %v", err)
	}
	if len(im_data) != 224*224*3 {
		t.Errorf("Expected image 224 * 224 * 3 = %v bytes, got %v bytes", 224*224*3, len(im_data))
	}

	im := rgb.Image{
		Pix:    im_data,
		Stride: 224 * 3,
		Rect: image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{224, 224},
		},
	}

	writer, err := os.Create("./test-rgb.png")
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Errorf("png.Encode() failed: %v", err)
	}
}

func TestGray(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Errorf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, err := JpegToSquareGray(reader, 224, InterpolationLanczos)
	if err != nil {
		t.Errorf("JpegToSquareGray() failed: %v", err)
	}
	if len(im_data) != 224*224*1 {
		t.Errorf("Expected image 224 * 224 * 1 = %v bytes, got %v bytes", 224*224*1, len(im_data))
	}

	im := image.Gray{
		Pix:    im_data,
		Stride: 224 * 1,
		Rect: image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{224, 224},
		},
	}

	writer, err := os.Create("./test-gray.png")
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Errorf("png.Encode() failed: %v", err)
	}
}
