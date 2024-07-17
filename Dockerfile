FROM ubuntu:24.04

COPY --chmod=755 ./democtl.sh /usr/local/bin/democtl
COPY --chmod=755 ./chrome-no-sandbox /usr/bin/chrome-no-sandbox

ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chrome-no-sandbox \
    PUPPETEER_SKIP_DOWNLOAD=true

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
    gnupg \
    && update-ca-certificates \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc \
    && chmod a+r /etc/apt/keyrings/docker.asc \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get install -U -y --no-install-recommends \
    docker-ce-cli \
    docker-compose-plugin \
    && wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && echo "deb [arch=$(dpkg --print-architecture)] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/chrome.list \
    && apt-get install -U -y --no-install-recommends \
    google-chrome-stable \
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
