FROM bitnami/kubectl:1.20.9 as kubectl

FROM golang:1.18-alpine
COPY --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/

# Install AWSCLIv2 with GLIBC
# https://github.com/aws/aws-cli/issues/4685#issuecomment-615872019
RUN apk --no-cache update && apk --no-cache add groff
ENV GLIBC_VER=2.31-r0
RUN apk --no-cache add \
        curl \
        binutils \
    && curl -sL https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-${GLIBC_VER}.apk \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-bin-${GLIBC_VER}.apk \
    && apk add --no-cache \
        glibc-${GLIBC_VER}.apk \
        glibc-bin-${GLIBC_VER}.apk \
    && curl -sL https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o awscliv2.zip \
    && unzip awscliv2.zip \
    && aws/install \
    && rm -rf \
        awscliv2.zip \
        aws \
        /usr/local/aws-cli/v2/*/dist/aws_completer \
        /usr/local/aws-cli/v2/*/dist/awscli/data/ac.index \
        /usr/local/aws-cli/v2/*/dist/awscli/examples \
    && apk --no-cache del \
        binutils \
    && rm glibc-${GLIBC_VER}.apk \
    && rm glibc-bin-${GLIBC_VER}.apk \
    && rm -rf /var/cache/apk/*

# Install common packages
RUN curl -sL https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 -o /usr/bin/jq && \
    chmod +x /usr/bin/jq

RUN curl -sL https://github.com/mikefarah/yq/releases/download/v4.2.0/yq_linux_amd64 -o /usr/bin/yq && \
    chmod +x /usr/bin/yq

RUN apk add --update --no-cache \
    bash \
    ca-certificates \
    git \
    openssh \
    python3 \
    py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir \
        requests \
    && rm -rf /var/cache/apk/*

RUN adduser -SD -h /app -s /bin/bash -u 1001 webhooker
WORKDIR /app

COPY go.* ./
COPY *.go ./
COPY pipe.sh ./
RUN chown -R webhooker /app

RUN go get github.com/go-cmd/cmd
RUN go mod download
RUN go build -o /app/webhooker

USER webhooker
EXPOSE 8080
CMD [ "/app/webhooker" ]