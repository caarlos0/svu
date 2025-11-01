FROM alpine
ARG TARGETPLATFORM
RUN apk add -U git
COPY $TARGETPLATFORM/*.apk /tmp/
RUN git config --global safe.directory '*'
RUN apk add --allow-untrusted /tmp/*.apk
ENTRYPOINT ["svu"]
