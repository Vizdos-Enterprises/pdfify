package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/vizdos-enterprises/pdfify/internal/api"
	internal_s3 "github.com/vizdos-enterprises/pdfify/internal/s3"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

type generationBody struct {
	HTML string `json:"html"`

	// If saving to S3, a key can be set.
	// If none, a random key will be used.
	UseKey string `json:"key,omit_empty"`

	// Specify which allowed bucket to save to
	// REQUIRED IF S3 ENABLED
	Bucket string `json:"bucket,omit_empty"`

	// If saving to bucket (s3), metadata can also be set.
	Metadata map[string]string `json:"metadata,omit_empty"`

	// If saving to bucket (s3), tags can also be set.
	Tags url.Values `json:"tags"`
}

type GenerationEndpointHTTP struct{}

func (c GenerationEndpointHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteResponse(w, api.APIResponseError{
			Reason: "Only POST requests are allowed",
		}, http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request generationBody
	err := decoder.Decode(&request)
	if err != nil {
		api.WriteResponse(w, api.APIResponseError{
			Reason: "Failed to decode request body",
		}, http.StatusBadRequest)
		return
	}

	if internal_s3.S3Enabled {
		fmt.Println(internal_s3.AllowedBuckets, request.Bucket)
		if !slices.Contains(internal_s3.AllowedBuckets, request.Bucket) {
			api.WriteResponse(w, api.APIResponseError{
				Reason: "S3 bucket not allowed",
			}, http.StatusForbidden)
			return
		}
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		buf.Reset() // Ensure buffer is cleared
		bufferPool.Put(buf)
		buf = nil // Help GC by setting buffer to nil
	}()

	err = Generate([]byte(request.HTML), buf)
	if err != nil {
		api.WriteResponse(w, api.APIResponseError{
			Reason:  "Failed to generate PDF",
			Details: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	if internal_s3.S3Enabled {
		fileKey := request.UseKey

		if fileKey == "" {
			fileKey = uuid.NewString()
		}

		_, err := internal_s3.S3client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(request.Bucket),
			Key:         &fileKey,
			Body:        buf,
			ContentType: aws.String("application/pdf"),
			Metadata:    request.Metadata,
			Tagging:     aws.String(request.Tags.Encode()),
		})

		if err != nil {
			api.WriteResponse(w, api.APIResponseError{
				Reason:  "Failed to save to S3",
				Details: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		api.WriteResponse(w, map[string]interface{}{
			"fileKey": fileKey,
		}, http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, buf)
	if err != nil {
		api.WriteResponse(w, api.APIResponseError{
			Reason:  "Failed to copy PDF buffer",
			Details: err.Error(),
		}, http.StatusInternalServerError)
	}
}
