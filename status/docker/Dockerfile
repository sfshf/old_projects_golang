FROM node:18-alpine AS base

# 1. Install dependencies only when needed
FROM base AS deps
# Check https://github.com/nodejs/docker-node/tree/b4117f9333da4138b03a546ec926ef50a31506c3#nodealpine to understand why libc6-compat might be needed.
RUN apk add --no-cache libc6-compat
WORKDIR /app
# Install dependencies based on the preferred package manager
COPY backend/package.json backend/yarn.lock* backend/package-lock.json* backend/pnpm-lock.yaml* ./backend/
COPY ui/package.json ui/yarn.lock* ui/package-lock.json* ui/pnpm-lock.yaml* ./ui/
RUN cd /app/backend && \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i; \
  else echo "Lockfile not found." && exit 1; \
  fi
RUN cd /app/ui && \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i; \
  else echo "Lockfile not found." && exit 1; \
  fi

# 2. Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app/ui
COPY --from=deps /app/ui/node_modules ./node_modules
COPY ./ui ./
# This will do the trick, use the corresponding env file for each environment.
RUN yarn build

# 3. Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app/backend
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001
COPY --from=builder --chown=nextjs:nodejs /app/ui/public /app/ui/
# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/ui/.next/standalone /app/ui/
COPY --from=builder --chown=nextjs:nodejs /app/ui/.next/static /app/ui/.next/static
COPY --chown=nextjs:nodejs docker/entrypoint.sh /entrypoint.sh
COPY --chown=nextjs:nodejs ./backend /app/backend
COPY --from=deps --chown=nextjs:nodejs /app/backend/node_modules /app/backend/node_modules
USER nextjs
EXPOSE 3000
EXPOSE 4010
ENTRYPOINT ["/entrypoint.sh"]
CMD ["node", "main.js"]