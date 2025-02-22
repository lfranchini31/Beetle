FROM golang:1.15.7

ARG BEETLE_VERSION=0.2.2

ENV GO111MODULE=on

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN mkdir -p /app/var/storage
RUN apt-get update

WORKDIR /app

RUN curl -sL https://github.com/Clivern/Beetle/releases/download/v${BEETLE_VERSION}/Beetle_${BEETLE_VERSION}_Linux_x86_64.tar.gz | tar xz
RUN rm LICENSE
RUN rm README.md
RUN mv Beetle beetle

COPY ./config.dist.yml /app/configs/

EXPOSE 8080

VOLUME /app/configs
VOLUME /app/var

RUN ./beetle version

CMD ["./beetle", "serve", "-c", "/app/configs/config.dist.yml"]
