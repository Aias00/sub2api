# Cloudbase

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26.4-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-8+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

<a href="https://trendshift.io/repositories/21823" target="_blank"><img src="https://trendshift.io/api/badge/repositories/21823" alt="Wei-Shaw%2Fcloudbase | Trendshift" width="250" height="55"/></a>

**Self-hosted AI Gateway and Business Operations Platform**

English | [中文](README_CN.md) | [日本語](README_JA.md)

</div>

## ⚠️ Important Notice

Please read the following carefully before using this project:

- **🚨 Terms of Service Risk**: Using this project may violate the terms of service of Anthropic and other upstream providers. Please review the relevant providers' user agreements before use; all risks arising from such use are borne solely by the user.
- **⚖️ Compliant Use**: Use this project only in compliance with the laws and regulations of your country or region. Any unlawful use is strictly prohibited.
- **📖 Disclaimer**: This project is provided for technical learning and research purposes only. The authors assume no liability for account bans, service interruptions, data loss, or any other direct or indirect damages resulting from the use of this project.

## Overview

Cloudbase is a self-hosted AI gateway and business operations platform for teams that need to turn multiple upstream AI accounts, API keys, and subscription resources into a managed service. It provides one control plane for request routing, account pools, user access, billing, payments, quota policy, observability, and worker-driven business workflows.

## Business Capabilities

- **Unified AI Gateway** - Route OpenAI-compatible, Claude, Gemini, Grok, Codex, Claude Code, and other AI traffic through managed endpoints
- **Account Pool Operations** - Manage OAuth accounts, API keys, proxies, account groups, priorities, sticky sessions, and failover policy from one console
- **Access and Quota Control** - Issue user API keys, bind groups, configure concurrency, RPM/TPM limits, model access, subscription plans, and per-platform quotas
- **Billing and Payments** - Track token-level usage, maintain balance and gift-credit buckets, support subscriptions, promo/redeem codes, refunds, affiliates, and self-service top-up
- **Payment Provider Management** - Built-in EasyPay, Alipay, WeChat Pay, Stripe, Creem, and Waffo integrations with provider instances, routing, limits, webhook handling, and reconciliation ([Configuration Guide](docs/PAYMENT.md))
- **Usage Analytics and Auditing** - Provide request logs, cost breakdowns, token statistics, user/account attribution, audit trails, and operational troubleshooting data
- **Operational Monitoring** - Monitor channel health, request latency, availability, error passthrough, risk controls, scheduler behavior, and worker status
- **Worker-Driven Business Features** - Run optional workers for WeChat article export, image workspace generation, hot-topic/content jobs, and other business workflows
- **Admin and Public Portals** - Deliver setup, user management, payment, profile, custom menu, public checkout, and administrator workflows through the web console
- **Deployment Automation** - Ship as an embedded Go backend with built frontend assets, Docker Compose profiles, GHCR image publishing, and install scripts

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.26.4, Gin, Ent, go-redis, Zap |
| Frontend | Vue 3.4+, Vite 5, TypeScript 5.6, Pinia, TailwindCSS |
| Database | PostgreSQL 18+ |
| Cache/Queue | Redis 8+ |
| Workers | Node.js workers for WeChat export, image workspace, and hot content jobs |
| Deployment | Docker Compose, embedded frontend assets, GHCR images |

---

## Nginx Reverse Proxy Note

When using Nginx as a reverse proxy for Cloudbase (or CRS) with Codex CLI, add the following to the `http` block in your Nginx configuration:

```nginx
underscores_in_headers on;
```

Nginx drops headers containing underscores by default (e.g. `session_id`), which breaks sticky session routing in multi-account setups.

---

## Deployment

### Method 1: Script Installation (Recommended)

One-click installation script that downloads pre-built binaries from GitHub Releases.

#### Prerequisites

- Linux server (amd64 or arm64)
- PostgreSQL 18+ (installed and running)
- Redis 8+ (installed and running)
- Root privileges

#### Installation Steps

```bash
curl -sSL https://raw.githubusercontent.com/Wei-Shaw/cloudbase/main/deploy/install.sh | sudo bash
```

The script will:
1. Detect your system architecture
2. Download the latest release
3. Install binary to `/opt/cloudbase`
4. Create systemd service
5. Configure system user and permissions

#### Post-Installation

```bash
# 1. Start the service
sudo systemctl start cloudbase

# 2. Enable auto-start on boot
sudo systemctl enable cloudbase

# 3. Open Setup Wizard in browser
# http://YOUR_SERVER_IP:8080
```

The Setup Wizard will guide you through:
- Database configuration
- Redis configuration
- Admin account creation

#### Upgrade

You can upgrade directly from the **Admin Dashboard** by clicking the **Check for Updates** button in the top-left corner.

The web interface will:
- Check for new versions automatically
- Download and apply updates with one click
- Support rollback if needed

#### Useful Commands

```bash
# Check status
sudo systemctl status cloudbase

# View logs
sudo journalctl -u cloudbase -f

# Restart service
sudo systemctl restart cloudbase

# Uninstall
curl -sSL https://raw.githubusercontent.com/Wei-Shaw/cloudbase/main/deploy/install.sh | sudo bash -s -- uninstall -y
```

---

### Method 2: Docker Compose (Recommended)

Deploy with Docker Compose, including PostgreSQL and Redis containers.

#### Prerequisites

- Docker 20.10+
- Docker Compose v2+

#### Quick Start (One-Click Deployment)

Use the automated deployment script for easy setup:

```bash
# Create deployment directory
mkdir -p cloudbase-deploy && cd cloudbase-deploy

# Download and run deployment preparation script
curl -sSL https://raw.githubusercontent.com/Wei-Shaw/cloudbase/main/deploy/docker-deploy.sh | bash

# Start services
docker compose up -d

# View logs
docker compose logs -f cloudbase
```

**What the script does:**
- Downloads `docker-compose.local.yml` (saved as `docker-compose.yml`) and `.env.example`
- Generates secure credentials (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD)
- Creates `.env` file with auto-generated secrets
- Creates data directories (uses local directories for easy backup/migration)
- Displays generated credentials for your reference

#### Manual Deployment

If you prefer manual setup:

```bash
# 1. Clone the repository
git clone https://github.com/Wei-Shaw/cloudbase.git
cd cloudbase/deploy

# 2. Copy environment configuration
cp .env.example .env

# 3. Edit configuration (generate secure passwords)
nano .env
```

**Required configuration in `.env`:**

```bash
# PostgreSQL password (REQUIRED)
POSTGRES_PASSWORD=your_secure_password_here

# JWT Secret (RECOMMENDED - keeps users logged in after restart)
JWT_SECRET=your_jwt_secret_here

# TOTP Encryption Key (RECOMMENDED - preserves 2FA after restart)
TOTP_ENCRYPTION_KEY=your_totp_key_here

# Optional: Admin account
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your_admin_password

# Optional: Custom port
SERVER_PORT=8080
```

**Generate secure secrets:**
```bash
# Generate JWT_SECRET
openssl rand -hex 32

# Generate TOTP_ENCRYPTION_KEY
openssl rand -hex 32

# Generate POSTGRES_PASSWORD
openssl rand -hex 32
```

```bash
# 4. Create data directories (for local version)
mkdir -p data postgres_data redis_data

# 5. Start all services
# Option A: Local directory version (recommended - easy migration)
docker compose -f docker-compose.local.yml up -d

# Option B: Named volumes version (simple setup)
docker compose up -d

# 6. Check status
docker compose -f docker-compose.local.yml ps

# 7. View logs
docker compose -f docker-compose.local.yml logs -f cloudbase
```

#### Deployment Versions

| Version | Data Storage | Migration | Best For |
|---------|-------------|-----------|----------|
| **docker-compose.local.yml** | Local directories | ✅ Easy (tar entire directory) | Production, frequent backups |
| **docker-compose.yml** | Named volumes | ⚠️ Requires docker commands | Simple setup |

**Recommendation:** Use `docker-compose.local.yml` (deployed by script) for easier data management.

#### Access

Open `http://YOUR_SERVER_IP:8080` in your browser.

If admin password was auto-generated, find it in logs:
```bash
docker compose -f docker-compose.local.yml logs cloudbase | grep "admin password"
```

#### Upgrade

```bash
# Pull latest image and recreate container
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

#### Easy Migration (Local Directory Version)

When using `docker-compose.local.yml`, migrate to a new server easily:

```bash
# On source server
docker compose -f docker-compose.local.yml down
cd ..
tar czf cloudbase-complete.tar.gz cloudbase-deploy/

# Transfer to new server
scp cloudbase-complete.tar.gz user@new-server:/path/

# On new server
tar xzf cloudbase-complete.tar.gz
cd cloudbase-deploy/
docker compose -f docker-compose.local.yml up -d
```

#### Useful Commands

```bash
# Stop all services
docker compose -f docker-compose.local.yml down

# Restart
docker compose -f docker-compose.local.yml restart

# View all logs
docker compose -f docker-compose.local.yml logs -f

# Remove all data (caution!)
docker compose -f docker-compose.local.yml down
rm -rf data/ postgres_data/ redis_data/
```

---

### Method 3: Build from Source

Build and run from source code for development or customization.

#### Prerequisites

- Go 1.26.4+
- Node.js 18+
- PostgreSQL 18+
- Redis 8+

#### Build Steps

```bash
# 1. Clone the repository
git clone https://github.com/Wei-Shaw/cloudbase.git
cd cloudbase

# 2. Install pnpm (if not already installed)
npm install -g pnpm

# 3. Build frontend
pnpm install
pnpm run frontend:build
# Output will be in backend/internal/web/dist/

# 4. Build backend with embedded frontend
cd backend
VERSION="$(./scripts/resolve-version.sh)"
go build -tags embed -ldflags="-X main.Version=${VERSION}" -o cloudbase ./cmd/server

# 5. Configure runtime environment
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=your_password
export DATABASE_DBNAME=cloudbase
export DATABASE_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export JWT_SECRET=change-this-to-a-secure-random-string
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export SERVER_MODE=release

# Optional: copy deploy/config.example.yaml to config.yaml only for
# settings that are not convenient to express as environment variables.
```

> **Note:** The `-tags embed` flag embeds the frontend into the binary. Without this flag, the binary will not serve the frontend UI.

`config.yaml` is optional. Environment variables are the preferred runtime
configuration surface for Docker and local binary deployments.

### Sora Status (Temporarily Unavailable)

> ⚠️ Sora-related features are temporarily unavailable due to technical issues in upstream integration and media delivery.
> Please do not rely on Sora in production at this time.
> Existing `gateway.sora_*` configuration keys are reserved and may not take effect until these issues are resolved.

Additional security-related options can be set in `config.yaml` when needed:

- `cors.allowed_origins` for CORS allowlist
- `security.url_allowlist` for upstream/pricing/CRS host allowlists
- `security.url_allowlist.enabled` to disable URL validation (use with caution)
- `security.url_allowlist.allow_insecure_http` to allow HTTP URLs when validation is disabled
- `security.url_allowlist.allow_private_hosts` to allow private/local IP addresses
- `security.response_headers.enabled` to enable configurable response header filtering (disabled uses default allowlist)
- `security.csp` to control Content-Security-Policy headers
- `billing.circuit_breaker` to fail closed on billing errors
- `server.trusted_proxies` to enable X-Forwarded-For parsing
- `turnstile.required` to require Turnstile in release mode

**⚠️ Security Warning: HTTP URL Configuration**

When `security.url_allowlist.enabled=false`, the system performs minimal URL validation by default, **rejecting HTTP URLs** and only allowing HTTPS. To allow HTTP URLs (e.g., for development or internal testing), you must explicitly set:

```yaml
security:
  url_allowlist:
    enabled: false                # Disable allowlist checks
    allow_insecure_http: true     # Allow HTTP URLs (⚠️ INSECURE)
```

**Or via environment variable:**

```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
```

**Risks of allowing HTTP:**
- API keys and data transmitted in **plaintext** (vulnerable to interception)
- Susceptible to **man-in-the-middle (MITM) attacks**
- **NOT suitable for production** environments

**When to use HTTP:**
- ✅ Development/testing with local servers (http://localhost)
- ✅ Internal networks with trusted endpoints
- ✅ Testing account connectivity before obtaining HTTPS
- ❌ Production environments (use HTTPS only)

**Example error without this setting:**
```
Invalid base URL: invalid url scheme: http
```

If you disable URL validation or response header filtering, harden your network layer:
- Enforce an egress allowlist for upstream domains/IPs
- Block private/loopback/link-local ranges
- Enforce TLS-only outbound traffic
- Strip sensitive upstream response headers at the proxy

#### ⚠️ Important: Creating the Admin Account

The initial admin account is **only created via the setup wizard** (served at `http://<host>:8080` on first run). The `default.admin_email` / `default.admin_password` fields in `config.yaml` are **not used** to create it — they exist in the template for historical reasons.

Because step 5 above pre-creates `config.yaml`, the setup wizard will be **skipped on first run**: the server detects an existing config and boots straight into normal mode with an empty `users` table, so the first login attempt fails with `invalid email or password`.

**Two ways to create the admin account:**

1. **Recommended — let the wizard generate `config.yaml`:** Skip step 5 (do not run the `cp`). Start `./cloudbase` directly; the setup wizard at `http://localhost:8080` walks you through database, Redis, and admin account setup, then writes `config.yaml` for you.

2. **If you already created `config.yaml`:** Temporarily move it aside so the wizard can trigger on first run, then restore it afterwards:
   ```bash
   mv config.yaml config.yaml.bak
   ./cloudbase        # wizard runs at http://localhost:8080 and writes a fresh config.yaml
   # stop the server (Ctrl+C) once the wizard completes, then restore your config:
   mv config.yaml.bak config.yaml
   ./cloudbase        # restart in normal mode and log in with the admin you just created
   ```

```bash
# 6. Run the application
./cloudbase
```

#### Development Mode

```bash
# Backend (with hot reload)
cd backend
go run ./cmd/server

# Frontend (with hot reload)
pnpm run frontend:dev
```

#### Code Generation

When editing `backend/ent/schema`, regenerate Ent + Wire:

```bash
cd backend
go generate ./ent
go generate ./cmd/server
```

---

## Simple Mode

Simple Mode is designed for individual developers or internal teams who want quick access without full SaaS features.

- Enable: Set environment variable `RUN_MODE=simple`
- Difference: Hides SaaS-related features and skips billing process
- Security note: In production, you must also set `SIMPLE_MODE_CONFIRM=true` to allow startup

---

## Grok / xAI OAuth Support

Cloudbase supports Grok subscription accounts through xAI OAuth and forwards OpenAI-compatible Responses traffic to xAI.

### Supported Scope

- Platform name: `grok`
- Account type: OAuth subscription accounts
- Public Responses targets: `/v1/responses`, `/responses`, and `/backend-api/codex/responses`, forwarded to `${XAI_BASE_URL:-https://api.x.ai/v1}/responses`
- Public Claude-compatible target: `/v1/messages`, converted to xAI Responses and returned as Anthropic Messages output for Claude CLI style clients
- Public Chat Completions targets: `/v1/chat/completions` and `/chat/completions`, forwarded to `${XAI_BASE_URL:-https://api.x.ai/v1}/chat/completions`
- Codex CLI style Responses WebSocket ingress is accepted on the Responses targets and bridged to xAI HTTP/SSE Responses upstream
- Initial models: `grok-4.3`, `grok-build-0.1`, `grok-4.20-0309-reasoning`, `grok-4.20-0309-non-reasoning`, and `grok-4.20-multi-agent-0309`
- Out of scope for this provider: image, video, TTS, transcription, browser automation, cookies, and Grok web scraping

### OAuth Configuration

The Grok OAuth flow uses PKCE and does not require committing private secrets. The default client details follow the public xAI OAuth flow used by compatible clients, and every value can be overridden by environment variable:

| Variable | Default |
|----------|---------|
| `XAI_OAUTH_CLIENT_ID` | Public xAI OAuth client ID |
| `XAI_OAUTH_SCOPE` | `openid profile email offline_access grok-cli:access api:access` |
| `XAI_OAUTH_REDIRECT_URI` | `http://127.0.0.1:56121/callback` |
| `XAI_OAUTH_AUTHORIZE_URL` | `https://auth.x.ai/oauth2/authorize` |
| `XAI_OAUTH_TOKEN_URL` | `https://auth.x.ai/oauth2/token` |
| `XAI_BASE_URL` | `https://api.x.ai/v1` |

Administrators can create or reauthorize Grok accounts from the dashboard, or use the admin API:

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/admin/grok/oauth/auth-url` | Generate an xAI OAuth authorization URL |
| `POST /api/v1/admin/grok/oauth/exchange-code` | Exchange a callback URL, query string, or code for OAuth credentials |
| `POST /api/v1/admin/grok/oauth/refresh-token` | Validate or refresh a Grok refresh token |
| `POST /api/v1/admin/grok/accounts/:id/refresh` | Refresh an existing Grok account |

Credential storage reuses the existing account JSON fields: `access_token`, `refresh_token`, `token_type`, `expires_at`, optional `email`, optional `subscription_tier`, and `entitlement_status`.

### Usage And Quota Display

xAI quota is passive. Cloudbase does not invent subscription quota values; it records whitelisted xAI rate-limit headers from successful or rate-limited upstream responses when xAI sends them. Before the first usable upstream response, the dashboard shows quota as unknown and still displays local Cloudbase usage stats.

`401` responses mark the account as needing reauthorization. `403` responses are treated as entitlement or subscription-tier failures instead of token-refresh loops. `429` responses use `Retry-After` or a short cooldown to temporarily remove the account from scheduling.

---

## Antigravity Support

Cloudbase supports [Antigravity](https://antigravity.so/) accounts. After authorization, dedicated endpoints are available for Claude and Gemini models.

### Dedicated Endpoints

| Endpoint | Model |
|----------|-------|
| `/antigravity/v1/messages` | Claude models |
| `/antigravity/v1beta/` | Gemini models |

### Claude Code Configuration

```bash
export ANTHROPIC_BASE_URL="http://localhost:8080/antigravity"
export ANTHROPIC_AUTH_TOKEN="sk-xxx"
```

### Hybrid Scheduling Mode

Antigravity accounts support optional **hybrid scheduling**. When enabled, the general endpoints `/v1/messages` and `/v1beta/` will also route requests to Antigravity accounts.

> **⚠️ Warning**: Anthropic Claude and Antigravity Claude **cannot be mixed within the same conversation context**. Use groups to isolate them properly.

### Known Issues

In Claude Code, Plan Mode cannot exit automatically. (Normally when using the native Claude API, after planning is complete, Claude Code will pop up options for users to approve or reject the plan.)

**Workaround**: Press `Shift + Tab` to manually exit Plan Mode, then type your response to approve or reject the plan.

---

## Project Structure

```
cloudbase/
├── backend/                  # Go backend service
│   ├── cmd/server/           # Application entry
│   ├── internal/             # Internal modules
│   │   ├── config/           # Configuration
│   │   ├── model/            # Data models
│   │   ├── service/          # Business logic
│   │   ├── handler/          # HTTP handlers
│   │   └── gateway/          # API gateway core
│   └── resources/            # Static resources
│
├── frontend/                 # Vue 3 frontend
│   └── src/
│       ├── api/              # API calls
│       ├── stores/           # State management
│       ├── views/            # Page components
│       └── components/       # Reusable components
│
└── deploy/                   # Deployment files
    ├── docker-compose.yml    # Docker Compose configuration
    ├── .env.example          # Environment variables for Docker Compose
    ├── config.example.yaml   # Optional config override example
    └── install.sh            # One-click installation script
```

## Disclaimer

> **Please read carefully before using this project:**
>
> :rotating_light: **Terms of Service Risk**: Using this project may violate Anthropic's Terms of Service. Please read Anthropic's user agreement carefully before use. All risks arising from the use of this project are borne solely by the user.
>
> :book: **Disclaimer**: This project is for technical learning and research purposes only. The author assumes no responsibility for account suspension, service interruption, or any other losses caused by the use of this project.

---

## Star History

<a href="https://star-history.com/#Wei-Shaw/cloudbase&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=Wei-Shaw/cloudbase&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=Wei-Shaw/cloudbase&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=Wei-Shaw/cloudbase&type=Date" />
 </picture>
</a>

---

## License

This project is licensed under the [GNU Lesser General Public License v3.0](LICENSE) (or later).

Copyright (c) 2026 Wesley Liddick

---

<div align="center">

**If you find this project useful, please give it a star!**

</div>
