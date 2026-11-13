# syntax=docker/dockerfile:1.7
# Official images via AWS Public ECR (Docker Hub timeouts are common).
FROM public.ecr.aws/docker/library/node:22-alpine AS deps
WORKDIR /frontend
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \n    npm ci --no-audit --no-fund

FROM public.ecr.aws/docker/library/node:22-alpine AS build
WORKDIR /frontend
COPY --from=deps /frontend/node_modules ./node_modules
COPY . .
ARG NUXT_PUBLIC_SITE_URL
ARG NUXT_PUBLIC_API_URL
ARG NUXT_PUBLIC_TURNSTILE_SITE_KEY
ENV NUXT_PUBLIC_SITE_URL=$NUXT_PUBLIC_SITE_URL
ENV NUXT_PUBLIC_API_URL=$NUXT_PUBLIC_API_URL
ENV NUXT_PUBLIC_TURNSTILE_SITE_KEY=$NUXT_PUBLIC_TURNSTILE_SITE_KEY
ENV NITRO_PRESET=node-server
ENV NUXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM public.ecr.aws/docker/library/node:22-alpine AS runner
WORKDIR /frontend
RUN apk add --no-cache wget
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV NITRO_HOST=0.0.0.0
ENV NITRO_PORT=3000
COPY --from=build /frontend/.output ./.output
USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
