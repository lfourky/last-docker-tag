#Build
FROM golang:alpine AS build-env
ADD . /go/src/ldt
WORKDIR /go/src/ldt
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o ldt ldt.go
#Pack
FROM alpine
WORKDIR /bin/ldt/
COPY --from=build-env /go/src/ldt /bin/ldt/
ENTRYPOINT ["./ldt"]    