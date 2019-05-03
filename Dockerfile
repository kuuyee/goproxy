# FROM golang:alpine AS build

# RUN apk add --no-cache -U make

# COPY . /src/goproxy
# RUN cd /src/goproxy &&\
#     export CGO_ENABLED=0 &&\
#     make

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache -U git mercurial subversion bzr fossil

# COPY --from=build /src/goproxy/bin/goproxy /goproxy

COPY ./bin ./bin
COPY ./run.sh .
RUN chmod +x /app/run.sh
VOLUME /go

EXPOSE 8080

ENTRYPOINT ["/app/run.sh"]
