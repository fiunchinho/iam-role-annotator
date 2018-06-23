FROM alpine:3.7

LABEL maintainer="Jose Armesto <jose@armesto.net>"

WORKDIR "/go/src/github.com/fiunchinho/iam-role-annotator"

RUN apk --no-cache add tini=0.16.1-r0

ENTRYPOINT ["/sbin/tini", "--", "./app"]
CMD ["--help"]

COPY build/iam-role-annotator-linux-amd64 app
