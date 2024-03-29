FROM golang:alpine AS build_base
#ENV GOARCH arm64
#ENV GOARCH amd64
RUN apk add --no-cache git gcc ca-certificates libc-dev \
    && mkdir -p /go/src/github.com/librespeed/ \
    && cd /go/src/github.com/librespeed/ \
    && git clone https://github.com/masx200/speedtest-go.git
WORKDIR /go/src/github.com/masx200/speedtest-go
RUN go get ./ && go build -ldflags "-w -s" -trimpath -o speedtest main.go

FROM alpine:3.15
RUN apk add ca-certificates
WORKDIR /app
COPY --from=build_base /go/src/github.com/masx200/speedtest-go/speedtest .
COPY --from=build_base /go/src/github.com/masx200/speedtest-go/web/assets ./assets
COPY --from=build_base /go/src/github.com/masx200/speedtest-go/settings.toml .

EXPOSE 8989

CMD ["./speedtest"]
