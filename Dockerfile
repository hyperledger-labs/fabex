FROM golang:1.14 as build-stage
LABEL maintainer="Vadim Inshakov <vadiminshakov@gmail.com>"
ADD . /app
WORKDIR /app
RUN go build

# production stage
FROM alpine:3.9 as production-stage
WORKDIR /app
COPY ./entrypoint.sh .
COPY --from=build-stage /app/fabex .
COPY --from=build-stage /app/configs ./configs
RUN apk add --no-cache libc6-compat
ENTRYPOINT ["./entrypoint.sh"]