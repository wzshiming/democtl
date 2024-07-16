FROM ubuntu:24.04

COPY --chmod=755 ./democtl.sh /usr/local/bin/democtl

RUN apt-get install -U -y --no-install-recommends \
    ffmpeg \
    unzip \
    git \
    nodejs \
    npm \
    python3 \
    python3-pip \
    ca-certificates \
    wget \
    curl \
    && update-ca-certificates \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc \
    && chmod a+r /etc/apt/keyrings/docker.asc \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get install -U -y --no-install-recommends \
    docker-ce-cli \
    docker-compose-plugin \
    && /usr/local/bin/democtl \
    --install asciinema \
    --install playpty \
    --install svg_term_cli \
    --install svg_to_video \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && pip cache purge \
    && npm cache clean --force

ENTRYPOINT [ "/usr/local/bin/democtl" ]
