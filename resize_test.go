package ippresize

import (
	"github.com/anight/go-libjpeg/rgb"
	"image"
	"image/png"
	"os"
	"testing"
)

func TestSquareRGB(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Errorf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, err := JpegToSquareRGB(reader, 224, InterpolationLanczos)
	if err != nil {
		t.Fatalf("JpegToSquareRGB() failed: %v", err)
	}
	if len(im_data) != 224*224*3 {
		t.Fatalf("Expected image 224 * 224 * 3 = %v bytes, got %v bytes", 224*224*3, len(im_data))
	}

	im := rgb.Image{
		Pix:    im_data,
		Stride: 224 * 3,
		Rect: image.Rectangle{
			Max: image.Point{224, 224},
		},
	}

	writer, err := os.Create("./test-square-rgb.png")
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}

func TestSquareGray(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Fatalf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, err := JpegToSquareGray(reader, 224, InterpolationLanczos)
	if err != nil {
		t.Fatalf("JpegToSquareGray() failed: %v", err)
	}
	if len(im_data) != 224*224*1 {
		t.Fatalf("Expected image 224 * 224 * 1 = %v bytes, got %v bytes", 224*224*1, len(im_data))
	}

	im := image.Gray{
		Pix:    im_data,
		Stride: 224 * 1,
		Rect: image.Rectangle{
			Max: image.Point{224, 224},
		},
	}

	writer, err := os.Create("./test-square-gray.png")
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}

func TestRGB(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Errorf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, im_size, err := JpegToRGB(reader, image.Point{224, 224}, false, InterpolationLanczos)
	if err != nil {
		t.Fatalf("JpegToRGB() failed: %v", err)
	}
	if len(im_data) != im_size.X*im_size.Y*3 {
		t.Fatalf("Expected image %d * %d * 3 = %v bytes, got %v bytes", im_size.X, im_size.Y, im_size.X*im_size.Y*3, len(im_data))
	}

	im := rgb.Image{
		Pix:    im_data,
		Stride: im_size.X * 3,
		Rect: image.Rectangle{
			Max: im_size,
		},
	}

	writer, err := os.Create("./test-rgb.png")
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}

func TestGray(t *testing.T) {
	reader, err := os.Open("./test.jpg")
	if err != nil {
		t.Errorf("os.Open() failed: %v", err)
	}
	defer reader.Close()
	im_data, im_size, err := JpegToGray(reader, image.Point{224, 224}, false, InterpolationLanczos)
	if err != nil {
		t.Fatalf("JpegToGray() failed: %v", err)
	}
	if len(im_data) != im_size.X*im_size.Y*1 {
		t.Fatalf("Expected image %d * %d * 1 = %v bytes, got %v bytes", im_size.X, im_size.Y, im_size.X*im_size.Y*1, len(im_data))
	}

	im := image.Gray{
		Pix:    im_data,
		Stride: im_size.X * 1,
		Rect: image.Rectangle{
			Max: im_size,
		},
	}

	writer, err := os.Create("./test-gray.png")
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	defer writer.Close()

	if err := png.Encode(writer, &im); err != nil {
		t.Fatalf("png.Encode() failed: %v", err)
	}
}
