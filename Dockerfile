FROM golang:1.12 as build-stage
LABEL maintainer="Vadim Inshakov <vadiminshakov@gmail.com>"
WORKDIR /app
COPY . .
RUN GOSUMDB=off GOPROXY=direct go build

# production stage
FROM alpine:3.9 as production-stage
WORKDIR /app
COPY --from=build-stage /app/fabex .
COPY --from=build-stage /app/configs .
RUN apk add --no-cache libc6-compat
CMD ["/app/fabex --task grpc --configpath /app/ --configname config --enrolluser true --interval 5"]
