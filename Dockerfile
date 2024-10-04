# Build the app
FROM golang:1.22 AS build-env

# Define build-time variables
ARG BUILD_VERSION
ARG BUILD_DATE
ARG BUILD_COMMIT
ARG BUILD_TAGS
ARG GOARCH
ARG GOOS

WORKDIR /app

COPY . ./
COPY .git .

RUN apt-get update

# Install go dependencies
RUN go mod download

RUN CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=${GOOS} go build -ldflags \
    "-X 'github.com/vizdos-enterprises/pdfify/internal/ldflags.BUILD_VERSION=${BUILD_VERSION}' \
    -X 'github.com/vizdos-enterprises/pdfify/internal/ldflags.BUILD_DATE=${BUILD_DATE}' \
    -X 'github.com/vizdos-enterprises/pdfify/internal/ldflags.BUILD_COMMIT=${BUILD_COMMIT}'" \
    -tags "${BUILD_TAGS}" \
    -o pdfify

# stub_email_addr
#
# Final Stage (make the image smaller w/ multistages)
FROM alpine

WORKDIR /

COPY --from=build-env /app/pdfify /pdfify

ENV PDFIFY_USE_S3=true
ENV PDFIFY_ALLOWED_BUCKETS=example-bucket
ENV PDFIFY_AWS_REGION=us-east-2

ENTRYPOINT ["./pdfify"]
