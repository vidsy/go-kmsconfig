FROM alpine:3.7
LABEL maintainer="Vidsy <tech@vidsy.co>"
ARG VERSION
LABEL version=$VERSION
ENV AWS_REGION eu-west-1
RUN apk --update upgrade openssl && \
    apk add ca-certificates      && \
    update-ca-certificates       && \
    rm -rf /var/cache/apk/*
ADD go-kmsconfig /usr/bin/go-kmsconfig
RUN chmod u+x /usr/bin/go-kmsconfig
ENTRYPOINT ["go-kmsconfig"]