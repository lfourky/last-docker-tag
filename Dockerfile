#Build
FROM golang:alpine AS build-env
ADD . /last-docker-tag
WORKDIR /last-docker-tag
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o ldt ldt.go
#Pack
FROM alpine
WORKDIR /last-docker-tag
COPY --from=build-env /last-docker-tag /last-docker-tag/
ENTRYPOINT ["./ldt"]