FROM vidsyhq/go-base:latest
ENV AWS_REGION eu-west-1
ADD go-kmsconfig /usr/bin/go-kmsconfig
RUN chmod u+x /usr/bin/go-kmsconfig
ENTRYPOINT ["go-kmsconfig"]