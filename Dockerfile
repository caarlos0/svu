FROM alpine
RUN apk add -U git
COPY svu*.apk /tmp/
RUN git config --global safe.directory '*'
RUN apk add --allow-untrusted /tmp/*.apk
ENTRYPOINT ["svu"]
