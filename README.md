# PDFify

A light-weight, fast, and memory-efficient HTML to PDF microservice.

Due to it being purpose-built as a microservice, you can run it easily with Docker and AWS Fargate (for example).

There is a working SDK for Golang contained here which you can `go get` and use. There's also a general API that can be used by other programming languages, open a PR with an edit to this page if you make your own SDK :)

(please note, as of now, this is meant to be run inside of a private VPC, there is no auth!)

## Get Started:

Install:

```
go get github.com/vizdos-enterprises/pdfify
```

### S3 Mode

Set the following environment variables on the PDFify server:

```
PDFIFY_USE_S3=true
PDFIFY_ALLOWED_BUCKETS=exampleBucket1,exampleBucket2
PDFIFY_AWS_REGION=us-east-2
```

**NOTE THE ALLOWED_BUCKETS, WILL REJECT IF NOT PRESENT.**

Generate on the client (change base URL):

```
sdkInstance := sdk.NewSDK(sdk.Config{
	BaseURL: "http://localhost:9999",
})

testHTML := "<h1>Test S3 PDF</h1>"

opts := sdk.GenerationOptions{
	HTML:   &testHTML,
	Bucket: "exampleBucket1", // make sure this is allowed on PDFify server
	UseKey: "test-pdf-s3.pdf", // leave this blank if you want a random key
	Metadata: map[string]string{
		"test": "Custom Metadata",
	},
	Tags: url.Values{
		"owner": []string{"xyz123"},
	},
}

output, err := sdkInstance.GeneratePDF(opts)

// output.S3Key contains the S3 key used
```

### File Mode

Instead of saving to an S3 bucket, the API will return a `ContentType: application/pdf` response.

Set the following environment variables on the PDFify server:

```
PDFIFY_USE_S3=false
```

On the client side, check out the `sdk/test/sdk_integration_test.go` for an example (its pretty simple)
