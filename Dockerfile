FROM alpine
RUN apk add -U git
COPY svu*.apk /tmp/
RUN apk add --allow-untrusted /tmp/*.apk
ENTRYPOINT ["svu"]
