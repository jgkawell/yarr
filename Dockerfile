FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod verify

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o bin/yarr .

FROM alpine:latest

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=build /app/bin/yarr /usr/local/bin/yarr

EXPOSE 7070
ENTRYPOINT ["/usr/local/bin/yarr"]
CMD ["-addr", "0.0.0.0:7070"]
