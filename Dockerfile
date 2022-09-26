FROM bitnami/kubectl:1.20.9 as kubectl

FROM golang:1.16-alpine
COPY --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/

RUN apk add --update --no-cache \
    tar \
    gcc \
    bash \
    musl-dev \
    ca-certificates \
    git \
    jq \
    yq \
    openssh \
    linux-headers \
    python3 \
    py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir \
        awscli \
    && rm -rf /var/cache/apk/*

RUN adduser -SD -h /app -s /bin/bash -u 1001 webhooker
WORKDIR /app

COPY go.* ./
COPY *.go ./
COPY pipe.sh ./
RUN chown -R webhooker /app

RUN go mod download
RUN go build -o /app/webhooker

USER webhooker
EXPOSE 8080
CMD [ "/app/webhooker" ]