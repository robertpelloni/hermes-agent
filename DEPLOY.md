# Deployment Instructions

Hermes Agent can be deployed via multiple backends: local, Docker, SSH, Singularity, Modal, Daytona, and Vercel Sandbox.

## General Setup
1. Define environment variables in `.env` (using `.env.example` as a template). Do not hardcode secrets.
2. Run `hermes setup` to configure the system.
3. Start the gateway using `hermes gateway start`.

## Docker
A standard `docker-compose.yml` is provided for containerized setups.
```bash
docker-compose up -d
```
See `docker/` for advanced configurations.

## Serverless (Modal / Daytona)
Refer to the respective CLI plugins and configurations inside `plugins/environments/`.
