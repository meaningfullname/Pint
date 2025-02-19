package utils

import (
	"context"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryResponse struct {
	PublicID string
	URL      string
}

func UploadToCloudinary(file *multipart.FileHeader) (*CloudinaryResponse, error) {
	// Initialize cloudinary
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUD_NAME"),
		os.Getenv("CLOUD_API_KEY"),
		os.Getenv("CLOUD_API_SECRET"),
	)
	if err != nil {
		return nil, err
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Upload
	uploadResult, err := cld.Upload.Upload(
		context.Background(),
		src,
		uploader.UploadParams{Folder: "pinterest-clone"},
	)
	if err != nil {
		return nil, err
	}

	return &CloudinaryResponse{
		PublicID: uploadResult.PublicID,
		URL:      uploadResult.SecureURL,
	}, nil
}
