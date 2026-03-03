package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

// AvatarSize defines avatar dimensions.
type AvatarSize struct {
	Width  int
	Height int
	Suffix string
}

var (
	AvatarFull      = AvatarSize{Width: 200, Height: 200, Suffix: "full"}
	AvatarThumbnail = AvatarSize{Width: 50, Height: 50, Suffix: "thumb"}
	AvatarSizes     = []AvatarSize{AvatarFull, AvatarThumbnail}
)

// ResizedAvatar holds the resized image data for one size.
type ResizedAvatar struct {
	Size AvatarSize
	Data *bytes.Buffer
}

// ResizeAvatar decodes an image from r and produces resized JPEG variants.
func ResizeAvatar(r io.Reader) ([]ResizedAvatar, error) {
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	results := make([]ResizedAvatar, 0, len(AvatarSizes))
	for _, size := range AvatarSizes {
		resized := imaging.Fill(src, size.Width, size.Height, imaging.Center, imaging.Lanczos)

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("encode %s: %w", size.Suffix, err)
		}

		results = append(results, ResizedAvatar{
			Size: size,
			Data: &buf,
		})
	}

	return results, nil
}
