# Stage 1: Build powertagd from C source
FROM alpine:3.20 AS build-powertagd
RUN apk add --no-cache build-base
COPY powertagd-src/ /build/
WORKDIR /build
RUN make clean && make RELEASE=1

# Stage 2: Build powertag2mqtt from Go source
FROM golang:1.22-alpine AS build-mqtt
WORKDIR /build
COPY powertag2mqtt/ .
RUN go mod download && CGO_ENABLED=0 go build -o powertag2mqtt .

# Stage 3: Runtime
FROM ghcr.io/home-assistant/amd64-base:3.20

COPY --from=build-powertagd /build/powertagd /usr/bin/powertagd
COPY --from=build-mqtt /build/powertag2mqtt /usr/bin/powertag2mqtt
COPY run.sh /

RUN chmod +x /run.sh /usr/bin/powertagd /usr/bin/powertag2mqtt

CMD [ "/run.sh" ]
