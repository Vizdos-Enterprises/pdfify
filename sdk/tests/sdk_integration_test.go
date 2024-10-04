package sdk_tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/vizdos-enterprises/pdfify/sdk"
)

func TestSDKFile(t *testing.T) {
	sdkInstance := sdk.NewSDK(sdk.Config{
		BaseURL: "http://localhost:9999",
	})

	testHTML := "<h1>Test File PDF</h1>"

	opts := sdk.GenerationOptions{
		HTML:   &testHTML,
		Bucket: "", // No bucket means the PDF will be returned directly
	}

	output, err := sdkInstance.GeneratePDF(opts)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	// Test if PDF bytes are returned
	if output.S3Key != "" {
		t.Errorf("Expected direct PDF bytes but got S3Link: %s", output.S3Key)
	}

	if len(output.Bytes) == 0 {
		t.Errorf("Expected PDF bytes to be non-empty, but got empty response")
	}

	// Optional: Validate if the response is a PDF by checking the header
	if !bytes.HasPrefix(output.Bytes, []byte("%PDF-")) {
		t.Errorf("Expected PDF content, but got invalid content")
	}

	// Save the PDF to a file
	outputDir := "./outputs"
	outputFile := outputDir + "/file.pdf"

	// Create the directory if it doesn't exist
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Write the PDF bytes to the file
	err = ioutil.WriteFile(outputFile, output.Bytes, 0644)
	if err != nil {
		t.Fatalf("Failed to write PDF to file: %v", err)
	}

	fmt.Printf("PDF saved successfully to %s\n", outputFile)
}

func TestSDKS3(t *testing.T) {
	sdkInstance := sdk.NewSDK(sdk.Config{
		BaseURL: "http://localhost:9999",
	})

	testHTML := "<h1>Another Test S3 PDF</h1>"

	opts := sdk.GenerationOptions{
		HTML:   &testHTML,
		Bucket: "htmlpdftestbucket",
		UseKey: "test-pdf-s3.pdf",
		Metadata: map[string]string{
			"test": "Custom Metadata",
		},
		Tags: url.Values{
			"owner": []string{"xyz123"},
		},
	}

	output, err := sdkInstance.GeneratePDF(opts)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	// Test if an S3 link is returned
	if output.S3Key == "" {
		t.Errorf("Expected S3 link, but got empty string")
	}

	t.Logf("Received S3 Key: %s", output.S3Key)

	if output.S3Key != opts.UseKey {
		t.Errorf("Did not expect this key: %s", output.S3Key)
	}
}
