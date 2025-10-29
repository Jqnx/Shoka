package thumb

import "github.com/h2non/bimg"

func ToWEBP(src []byte) ([]byte, error) {
	// Create new image
	img := bimg.NewImage(src)

	// Convert to WEBP
	newImg, err := img.Convert(bimg.WEBP)
	if err != nil {
		return nil, err
	}
	return newImg, nil
}
