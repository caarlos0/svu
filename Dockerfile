FROM scratch

COPY svu /usr/local/bin/svu

ENTRYPOINT ["svu"]
