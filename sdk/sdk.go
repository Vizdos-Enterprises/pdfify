package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GenerationOptions struct {
	HTML     *string
	Bucket   string
	UseKey   string
	Metadata map[string]string
	Tags     url.Values
}

type GenerationOutput struct {
	S3Key string
	Bytes []byte
}

type apiResponse struct {
	FileKey string `json:"fileKey"`
	Error   string `json:"error"`
	Details string `json:"details"`
}

var (
	ErrEmptyHTML       = errors.New("HTML content is required")
	ErrFailedAPICall   = errors.New("failed to call the API")
	ErrInvalidResponse = errors.New("invalid response from the API")
)

type SDK struct {
	Config Config
}

func NewSDK(config Config) *SDK {
	return &SDK{Config: config}
}

func (s SDK) GeneratePDF(opts GenerationOptions) (*GenerationOutput, error) {
	if opts.HTML == nil || *opts.HTML == "" {
		return nil, ErrEmptyHTML
	}

	// Build request payload
	reqBody, err := json.Marshal(map[string]interface{}{
		"html":     opts.HTML,
		"key":      opts.UseKey,
		"bucket":   opts.Bucket,
		"metadata": opts.Metadata,
		"tags":     opts.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send POST request to the PDF generation API
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/generate", s.Config.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedAPICall, err)
	}
	defer resp.Body.Close()

	// Check for non-200 response codes
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status code %d, response: %s", ErrFailedAPICall, resp.StatusCode, string(bodyBytes))
	}

	// Handle response based on content type
	contentType := resp.Header.Get("Content-Type")
	output := &GenerationOutput{}

	if contentType == "application/json" {
		// Handle S3 link response
		var apiResp apiResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to decode JSON response: %v", ErrInvalidResponse, err)
		}
		if apiResp.Error != "" {
			return nil, fmt.Errorf("API error: %s, details: %s", apiResp.Error, apiResp.Details)
		}
		output.S3Key = apiResp.FileKey
	} else if contentType == "application/pdf" {
		// Handle direct PDF bytes response
		pdfBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF response: %w", err)
		}
		output.Bytes = pdfBytes
	} else {
		return nil, ErrInvalidResponse
	}

	return output, nil
}
