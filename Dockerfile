FROM nixos/nix:2.33.4 AS builder

RUN mkdir -p /etc/nix && \
    echo "filter-syscalls = false" >> /etc/nix/nix.conf && \
    echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf

WORKDIR /tmp/build
COPY . .

RUN --mount=type=cache,id=nix-eval,target=/root/.cache/nix \
    nix build --print-build-logs

RUN mkdir /tmp/nix-store-closure && \
    cp -a $(nix-store -qR result) /tmp/nix-store-closure/


FROM scratch

LABEL org.opencontainers.image.source="https://github.com/aegis-run/aegis"

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group  /etc/group
COPY --from=builder /tmp/nix-store-closure /nix/store
COPY --from=builder /tmp/build/result /app

USER nobody
WORKDIR /app

ENTRYPOINT ["/app/bin/aegis"]
CMD ["serve"]
