//go:build ignore

package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"

	"go.viam.com/rdk/app"
	"go.viam.com/rdk/logging"
)

func main() {
	ctx := context.Background()
	logger := logging.NewLogger("test-upload")

	apiKey := os.Getenv("VIAM_API_KEY")
	apiKeyID := os.Getenv("VIAM_API_KEY_ID")
	if apiKey == "" || apiKeyID == "" {
		fmt.Println("Set VIAM_API_KEY and VIAM_API_KEY_ID environment variables")
		os.Exit(1)
	}

	datasetID := "697019fc1c1ac68cecf7dc44"
	partID := "b1cb7431-c856-4c3b-8da1-9f7a10dd6df3"

	// Create a simple test image (10x10 red square)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	fmt.Println("Creating Viam client...")
	viamClient, err := app.CreateViamClientWithAPIKey(ctx, app.Options{}, apiKey, apiKeyID, logger)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer viamClient.Close()

	dataClient := viamClient.DataClient()

	fmt.Printf("Uploading test image to dataset %s...\n", datasetID)
	fileID, err := dataClient.UploadImageToDatasets(
		ctx,
		partID,
		img,
		[]string{datasetID},
		[]string{"test-upload"},
		app.MimeTypeJPEG,
		&app.FileUploadOptions{},
	)
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! File ID: %s\n", fileID)
}
