#####################################################################
# Compile stage
FROM golang:1.16.4 AS build-env
ADD . /plugmeter
WORKDIR /plugmeter
RUN go build -o /plugmeter



#####################################################################
# Final stage
FROM debian:buster
# FROM gcr.io/distroless/base-debian10
# FROM alpine:latest 

LABEL maintainer="Pierre Rust"

# RUN useradd -u 5000 plugmeter \
#  && mkdir -p /out \
#  && chown -R plugmeter:plugmeter /out

EXPOSE 3000
EXPOSE 5353/udp

ARG OUT_DIR=/out
RUN mkdir -p ${OUT_DIR}
VOLUME [${OUT_DIR}]
# ENV LOG_FILE_LOCATION=${LOG_DIR}/app.log 

WORKDIR /
COPY --from=build-env /plugmeter /

# USER plugmeter:plugmeter
CMD ["/plugmeter"]