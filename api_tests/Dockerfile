# Start with a base Docker image that includes Playwright and a specific browser version.
FROM mcr.microsoft.com/playwright:v1.39.0-jammy

WORKDIR /app

COPY package*.json ./

RUN npm ci

RUN npx playwright install --with-deps

COPY . .
