FROM golang:1.20.3-alpine3.17 as builder
LABEL authors="chrisyeo"
RUN mkdir /build
WORKDIR /build
COPY cmd/ ./cmd
COPY internal/ ./internal
COPY ./go.mod .
COPY ./go.sum .
RUN go build -o command-on-demand cmd/main.go

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /build/command-on-demand /usr/local/bin
CMD ["command-on-demand"]