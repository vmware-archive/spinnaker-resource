LABEL Maintainer="Matt Burdan <burdz@burdz.net>"

FROM golang:alpine as builder
COPY . /spinnaker-resource
WORKDIR /spinnaker-resource
ENV CGO_ENABLED 0
RUN apk add --update git gcc

RUN go build -o /assets/check cmd/check/main.go
RUN go build -o /assets/in cmd/in/main.go
RUN go build -o /assets/out cmd/out/main.go

FROM alpine:edge AS resource
COPY --from=builder /assets /opt/resource

FROM resource
