# Domicile

A self-hosted home management web application built with Go. Track your groceries, stock, shopping lists, recipes, meal plans, chores, equipment, and more.

![Domicile](/static/images/domicile.png)

## Features

- 📦 Product & stock tracking with expiry dates
- 🛒 Shopping lists with shop mode and print view
- 🧹 Chores management
- 🍳 Recipes with web scraping and meal planning
- 🔧 Equipment inventory with manual uploads
- 🔋 Battery tracking
- 📋 Task lists
- 👥 Multi-user support with admin panel
- 📤 Backup & restore

## Quick Start (Docker)

```yaml
services:
  domicile:
    image: ghcr.io/zodiac1913/domicile:latest
    container_name: domicile
    restart: unless-stopped
    ports:
      - "1323:1323"
    volumes:
      - ./data:/app/data
```

```bash
docker compose up -d
```

Then open [http://localhost:1323](http://localhost:1323) and create your first user account.

## Inspiration

Domicile wouldn't exist without these great projects:

- [Grocy](https://grocy.info) — The original self-hosted groceries & household management app
- [Mealie](https://mealie.io) — A beautiful self-hosted recipe manager

Thanks to both teams for paving the way.

## Pages

- [Installation](installation.md) — Docker setup, build from source, configuration
- [Configuration](configuration.md) — Environment variables and options
- [Updating](updating.md) — How to update to new versions
