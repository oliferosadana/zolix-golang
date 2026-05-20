FROM scratch

WORKDIR /app

COPY build/zolix-linux-arm64 /app/zolix
COPY web/static /app/web/static
COPY assets /app/assets

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/zolix"]
