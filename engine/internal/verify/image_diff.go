package verify

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// CompareImages compares two images pixel by pixel and returns the number of differing pixels.
func CompareImages(img1Path, img2Path string) (int, error) {
	img1, err := loadImage(img1Path)
	if err != nil {
		return 0, fmt.Errorf("verify: load image 1: %w", err)
	}
	img2, err := loadImage(img2Path)
	if err != nil {
		return 0, fmt.Errorf("verify: load image 2: %w", err)
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// If bounds differ, count all extra pixels as differences
	diffPixels := 0
	maxX := bounds1.Max.X
	if bounds2.Max.X > maxX {
		maxX = bounds2.Max.X
	}
	maxY := bounds1.Max.Y
	if bounds2.Max.Y > maxY {
		maxY = bounds2.Max.Y
	}

	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			if !inBounds(x, y, bounds1) || !inBounds(x, y, bounds2) {
				diffPixels++
				continue
			}

			c1 := img1.At(x, y)
			c2 := img2.At(x, y)

			if !colorMatch(c1, c2) {
				diffPixels++
			}
		}
	}

	return diffPixels, nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func inBounds(x, y int, b image.Rectangle) bool {
	return x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y
}

func colorMatch(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// For MVP, exact match only. Can add tolerance later.
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
