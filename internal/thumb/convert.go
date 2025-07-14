package thumb

import "github.com/h2non/bimg"

func ToWEBP(src string) ([]byte, error) {
	// Read image
	buf, err := bimg.Read(src)
	if err != nil {
		return nil, err
	}

	// Create new image
	img := bimg.NewImage(buf)

	// Convert to WEBP
	newImg, err := img.Convert(bimg.WEBP)
	if err != nil {
		return nil, err
	}
	return newImg, nil
}
