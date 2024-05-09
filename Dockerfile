#FROM --platform=$BUILDPLATFORM golang:1.20.12-bullseye as builder
#FROM --platform=$BUILDPLATFORM golang:1.19.9-alpine3.16 as builder
FROM --platform=$BUILDPLATFORM golang:1.20.12-alpine3.19 as builder


ARG VERSION=v0.0.0

RUN apk add --no-cache make git build-base

COPY . /app/nodekit-commit

# Creates an app directory to hold your app’s source code
WORKDIR /app/nodekit-commit
 
# Copies everything from your root directory into /app

RUN go mod download

RUN CGO_CFLAGS="-O -D__BLST_PORTABLE__" CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__" go build -o ./bin/nodekit-commit ./main.go

FROM alpine:3.19

COPY --from=builder /app/nodekit-commit/bin/nodekit-commit /usr/local/bin


CMD ["nodekit-commit"]

