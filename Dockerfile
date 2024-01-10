#FROM --platform=$BUILDPLATFORM golang:1.20.12-bullseye as builder
#FROM --platform=$BUILDPLATFORM golang:1.19.9-alpine3.16 as builder
FROM --platform=$BUILDPLATFORM golang:1.20.12-alpine3.19 as builder


ARG VERSION=v0.0.0

RUN apk add --no-cache make git build-base

# Creates an app directory to hold your app’s source code
WORKDIR /app
 
# Copies everything from your root directory into /app
COPY . .

RUN go mod download


RUN go build -o /nodekitl1

CMD [ “/nodekitl1 ]

#ARG TARGETOS TARGETARCH

#RUN make op-geth-proxy VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

# FROM alpine:3.16

# COPY --from=builder /app/op-geth-proxy/bin/op-geth-proxy /usr/local/bin



# CMD ["op-geth-proxy"]
