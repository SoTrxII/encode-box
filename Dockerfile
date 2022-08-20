# Dockerfile for an appservice go container (expected size 18Mo)
####################################################################################################
## Builder
####################################################################################################
FROM golang:alpine as builder
WORKDIR /app
COPY . .
# Build the app, strip it (LDFLAGS) and optimize it with UPX
RUN apk add upx && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/server -ldflags="-w -s" ./cmd/server.go && \
    upx --best --lzma ./build/server

####################################################################################################
## Final image
####################################################################################################
FROM alpine as release
RUN apk add --no-cache ffmpeg &&\
    # The runtime user, having no home dir nor password
    adduser -HD -s /bin/ash appuser

WORKDIR /app
# Copy the built app, only allowing our app user to execute it
COPY --from=builder --chmod=0500 --chown=appuser:appuser  /app/build/server ./

EXPOSE 8080
ENTRYPOINT [ "/server" ]