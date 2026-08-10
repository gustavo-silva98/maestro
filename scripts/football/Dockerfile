FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    xvfb \
    chromium \
    chromium-driver \
    && rm -rf /var/lib/apt/lists/*

RUN pip install selenium pillow pandas webdriver-manager pyscreenshot

ENV DISPLAY=:99

COPY football.py .

CMD ["bash", "-c", "Xvfb :99 -screen 0 1280x720x24 & sleep 1 && python football.py"]