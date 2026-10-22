# Upgrade UI Node.js to patched 24 LTS

## Overview
- The Node.js security release of 2026-10-21 fixes 10 CVEs. Four are High:
  Brotli decoder OOB write (RCE), two TLS hostname-verification bypasses, and
  `dns.resolve` NUL truncation. Patched versions are v22.23.4, v24.22.1 and
  v26.12.1. Node 20 is EOL (April 2026) and received no fix.
  Source: https://nodejs.org/en/blog/vulnerability/october-2026-security-releases
- The UI builds and, in the Docker Hub image, runs on vulnerable Node. Move
  everything to an exact pin of **Node 24.22.1** (active LTS).
- Only the UI (`ui/`) uses Node. The Go API, Makefile, `infrastructure/`
  compose files and `k8s/` only consume the UI images.

## Context (from discovery)
- `ui/Dockerfile.eks` builder stage uses `FROM node:20-alpine` (EOL,
  unpatched). The runtime stage is `nginx:alpine` with static files only,
  no Node, so it needs no change. `deploy_eks.yaml` builds it in an
  amd64/arm64 matrix. The build requires `VITE_API_URL`,
  `VITE_API_PASSWORD`, `VITE_API_USERNAME`, `VITE_ISSUER_NAME` and
  `VITE_IPFS_GATEWAY_URL`; their `test -n` checks fail otherwise.
- `ui/Dockerfile` uses `FROM nginx:alpine` plus `apk add nodejs npm`. It is
  built by `dockerhub-dev.yml` and `dockerhub-main.yml` for
  `linux/amd64,linux/arm64`, and by `deploy_ui.yml`. Node runs at container
  start: `scripts/deploy.sh` runs `npm run build`.
  - Checked 2026-10-22: `nginx:alpine` is Alpine 3.24.3, whose `nodejs` is
    `24.18.1-r0`. Alpine edge (3.25 alpha) has `24.18.1-r5`. **No Alpine
    release ships ≥ 24.22.1**, so pinning an Alpine release cannot fix this.
  - `node:24.22.1-alpine3.24` exists. It is Alpine 3.24.3, the same as
    `nginx:alpine`, and has node 24.22.1 and npm 11.19.1. `node` links only
    musl, `libstdc++` and `libgcc_s`.
- `.github/workflows/deploy_ui.yml` has an `actions/setup-node@v2` step with
  `node-version: 16`. Nothing in that job uses it; the build runs in
  `docker build`.
- `ui/package.json` has `engines.node: ">=20"` and `@types/node: ^22.7.5`,
  and no `typecheck` script. There is no `.nvmrc`, `packageManager` field or
  `.npmrc`.

## Development Approach
- **testing approach**: no app code changes. Verification means
  multi-arch image builds, a `node -v` check inside the images, and running
  the UI build and lint on the new Node.
- complete each task fully before moving to the next
- **CRITICAL: update this plan file when scope changes during implementation**

## Testing Strategy
- The project has no unit or e2e tests for Dockerfiles or CI. Each task
  verifies with `docker buildx build --platform linux/amd64,linux/arm64`,
  `node -v` inside the image (arm64 under QEMU too), `npm run build`, lint
  and `npx tsc --noEmit`.
- Lint the changed workflow with `actionlint`.

## Progress Tracking
- mark completed items with `[x]` immediately when done
- add newly discovered tasks with ➕ prefix
- document issues/blockers with ⚠️ prefix

## Solution Overview
- Pin `node:24.22.1-alpine3.24` wherever a Node image is used.
- `ui/Dockerfile`: keep the `nginx:alpine` runtime and the build-on-start
  behavior. Drop `apk add nodejs npm` and copy Node and npm from the pinned
  `node` stage instead.
- Keep `nginx:alpine` floating, as it is today, so the runtime keeps getting
  nginx and Alpine security rebuilds. musl and libstdc++ are backward
  compatible, so a binary built on 3.24 runs on later Alpine releases.
  Install `libstdc++` explicitly. The build-time `RUN npm install` runs
  the copied node and npm, so any incompatibility fails the build, not the
  container start.
- When Alpine ships `nodejs` ≥ 24.22.1, this could go back to apk. That is
  not in scope.

### Alternatives considered
- **Pin an Alpine release for apk `nodejs`**: rejected. No release ships
  ≥ 24.22.1 (3.24 and edge have 24.18.1).
- **Base on `node:24.22.1-alpine3.24` and `apk add nginx`**: rejected after
  testing:
  - Alpine's `nginx.conf` includes `conf.d/*.conf` outside `http {}`, so
    the `server {}` block `deploy.sh` copies there breaks startup. Server
    blocks belong in `http.d/`.
  - `/usr/share/nginx/html` does not exist.
  - Logs go to files, not stdout/stderr.
  - The `/docker-entrypoint.d` behaviour is lost.
  - nginx drops from 1.31.8 (nginx.org mainline) to 1.30.6 (Alpine).

  This means rebuilding the nginx image setup by hand. Copying Node is two
  `COPY` lines and an `npm` symlink.

## Technical Details
`ui/Dockerfile` target shape:
```dockerfile
FROM node:24.22.1-alpine3.24 AS node

FROM nginx:alpine

WORKDIR /app

RUN apk add --no-cache libstdc++ apache2-utils
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/npm
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm
COPY package.json package-lock.json ./
...rest unchanged
```
- corepack, yarn and Node headers are not copied. Nothing uses them, and
  apk `nodejs` did not ship them either.

Dummy build args for `Dockerfile.eks` verification:
```sh
--build-arg VITE_API_URL=http://localhost:3001 \
--build-arg VITE_API_PASSWORD=x --build-arg VITE_API_USERNAME=x \
--build-arg VITE_ISSUER_NAME=x --build-arg VITE_IPFS_GATEWAY_URL=https://ipfs.io
```

## What Goes Where
- **Implementation Steps**: Dockerfiles, workflow, package.json in this repo.
- **Post-Completion**: deploy and verify the running containers.

## Implementation Steps

### Task 1: Pin Node 24.22.1 in EKS UI image

**Files:**
- Modify: `ui/Dockerfile.eks`

- [x] change `FROM node:20-alpine AS builder` to
      `FROM node:24.22.1-alpine3.24 AS builder`; leave the nginx runtime
      stage unchanged
- [x] `docker buildx build --platform linux/amd64,linux/arm64 -f
      ui/Dockerfile.eks <dummy build args> ui` completes; `npm ci
      --legacy-peer-deps` and `npm run build` succeed on both architectures
      (719 packages added, vite build OK on both; needs a
      `docker-container` buildx builder, the default `docker` driver
      rejects multi-platform)
- [x] build `--target builder --load` and run `node -v`; it prints
      `v24.22.1` (amd64 and `--platform linux/arm64`) (both `v24.22.1`)

### Task 2: Replace apk Node with pinned Node in Docker Hub UI image

**Files:**
- Modify: `ui/Dockerfile`

- [x] add the `node:24.22.1-alpine3.24` stage; copy `node` and npm, and
      create the `npm` symlink as in Technical Details (the `npx` symlink
      was dropped after review: nothing in the image uses npx)
- [x] change the apk line to `apk add --no-cache libstdc++ apache2-utils`
      (drop `nodejs npm`) (a separate `node -v && npm -v` guard was dropped
      after review: the next `RUN npm install` already runs both)
- [x] remove the stale `# FROM nginx:1.29.3-alpine` comment
- [x] `docker buildx build --platform linux/amd64,linux/arm64 ui` completes
      (`npm install` runs on the copied npm) (done)
- [x] `docker run --rm --entrypoint sh <img> -c 'node -v && npm -v'` prints
      `v24.22.1` and `11.19.1` on amd64 and arm64 (both)
- [x] run the container with the env vars `deploy.sh` reads. The startup
      `npm run build` succeeds and nginx serves the UI (amd64 and arm64,
      insecure mode: vite build OK, `/` and assets return 200)

### Task 3: Update UI package metadata

**Files:**
- Modify: `ui/package.json`
- Modify: `ui/package-lock.json`

- [x] set `engines.node` to `"^24.22.1 || ^26.12.1"` (patched lines only;
      advisory: without `engine-strict` npm only warns)
- [x] check the latest `@types/node` 24.x on the registry
      (`npm view @types/node@24 version`); set `@types/node` to `^24.x`. If
      `tsc` breaks, keep `^22` and note it here (24.19.2; `^24.19.2`, tsc OK)
- [x] run `npm install` with Node 24.22.1 to refresh `package-lock.json`
      (first hand-edited to a minimal diff; after review the full
      `npm install` result was committed, which also prunes ~20 stale
      entries such as `meow`, so `npm install` and `npm ci` images resolve
      the same tree)
- [x] `npx tsc --noEmit`, `npm run lint`, `npm run lint-styles` and
      `npm run build` pass on Node 24.22.1 (all pass, also before the change)
- [x] re-run the Task 1 build (covers `npm ci --legacy-peer-deps` on the
      new lock) and the Task 2 build (both pass on amd64 and arm64)

### Task 4: Remove unused setup-node step from deploy_ui workflow

**Files:**
- Modify: `.github/workflows/deploy_ui.yml`

- [x] delete the `actions/setup-node@v2` step (`node-version: 16`); nothing
      in the job uses it
- [x] run `actionlint .github/workflows/deploy_ui.yml`; fix any new issues
      in the changed lines (no new issues; the setup-node@v2 finding is gone;
      pre-existing `checkout@v2` and `set-output` findings remain, out of scope)

### Task 5: Verify acceptance criteria
- [x] `grep -rn --exclude-dir=node_modules "24.22.1" ui .github` lists
      exactly the expected lines (both Dockerfiles, `package.json`)
      (plus the `package-lock.json` root entry mirroring `engines`)
- [x] `grep -rn --exclude-dir=node_modules -E "node:20|nodejs npm|setup-node" ui .github`
      returns nothing (empty)
- [x] both UI images build for amd64 and arm64 and report `v24.22.1`
      (EKS builder and Docker Hub image, both arches; npm 11.19.1)
- [x] scan both images (`trivy image` or `docker scout cves`); no node or
      nodejs CVE from the 2026-10-21 release is reported (trivy 0.76.1 in
      a container, scout needs login: 0 node/nodejs findings, none of the
      10 advisory CVEs). For `ui/Dockerfile` this is not evidence: trivy
      does not inventory the copied `node` binary, so `node -v` is the
      version proof

### Task 6: [Final] Update documentation
- [x] update README.md if it states a Node version (none found at
      discovery); after review added a "Development (UI)" section to the
      root README with the Node >= 24.22.1 requirement (fixes the
      `#development-ui` links from `ui/README.md` and `ui/.env.sample`)
- [x] move this plan to `docs/plans/completed/` (handed off: the
      orchestrator moves the file after reviews)

## Post-Completion
*Items requiring manual intervention or external systems*

**Deployment**:
- rebuild and push UI images (Docker Hub dev/main, ECR, EKS), then redeploy
- on the running containers, check `node -v` = `v24.22.1`

**Out of scope, worth a follow-up**:
- `aws-actions/configure-aws-credentials@v1-node16`, `actions/checkout@v2`
  and `amazon-ecr-login@v1` in `deploy_ui.yml`/`deploy_backend.yml` run on
  deprecated GitHub Actions runtimes. This is not related to this CVE.
- Once Alpine ships `nodejs` ≥ 24.22.1, `ui/Dockerfile` could go back to
  `apk add nodejs npm`.
- No CI job builds the UI Dockerfiles or runs lint/tsc/build before
  deploy (pre-existing gap). A path-filtered PR workflow, maybe also
  weekly to catch `nginx:alpine` moves, would need its own plan.
- Image scanners cannot see the Node copied into `ui/Dockerfile`. Track
  the `FROM node:` tag with Dependabot/Renovate, or scan the node stage.
- `engine-strict=true` in `ui/.npmrc` would turn `engines` into a hard
  check for developers; left advisory, as it changes local workflows.
- `ui/Dockerfile` uses `npm install`, not `npm ci`; switching is a
  behavior change outside this plan.
