# syntax=docker/dockerfile:1

FROM golang:1.24

WORKDIR /app

COPY buttonBox /

ENTRYPOINT ["/buttonBox"]
