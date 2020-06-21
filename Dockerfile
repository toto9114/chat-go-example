# Dockerfile

FROM golang:1.13.4

LABEL version="1.0.0" maintainer="Allen"

WORKDIR /chatting-example
COPY . /chatting-example

# General Packages
RUN apt-get update \
    && apt-get install -y software-properties-common \
    && apt-get install -y build-essential \
    && apt-get update \
    && apt-get install -y git \
    && go get github.com/cespare/reflex

RUN cd /chatting-example

EXPOSE 1213
CMD ["reflex", "-c", "reflex.conf"]