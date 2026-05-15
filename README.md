# Domicile

A self-hosted home management web application inspired by [Grocy](https://grocy.info), built with Go.

Domicile aims to be a modern, user-friendly alternative for managing your household — groceries, stock, shopping lists, chores, recipes, equipment manuals, and more.

## Tech Stack

- **Backend:** [Go](https://go.dev) + [Echo](https://echo.labstack.com) (web framework)
- **Templates:** [Templ](https://templ.guide) (type-safe HTML components)
- **Frontend:** [htmx](https://htmx.org) + [Bootstrap 5.3](https://getbootstrap.com)
- **Database:** SQLite via [sqlx](https://github.com/jmoiron/sqlx) + [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)

## Features (In Progress)

- 📦 Product & stock tracking with expiry dates
- 🛒 Shopping lists
- 🧹 Chores management
- 🍳 Recipes & meal planning
- 🔧 Inventory & instruction manuals
- 🔋 Battery tracking
- 📋 Tasks
- 👥 Multi-user support (planned)
- 📤 Import/Export (planned — because data portability matters)

## Prerequisites

- Go 1.25+
- [Templ CLI](https://templ.guide/quick-start/installation)

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

## Getting Started

```bash
# Clone the repo
git clone https://github.com/zodiac1913/Domicile.git
cd Domicile

# Install dependencies
go mod download

# Generate templ files
templ generate

# Run
go run server.go
```

The app will find an available port starting at 1323 and open your default browser automatically.

## Project Structure

```
├── server.go                  # Entry point
├── internal/
│   ├── database/
│   │   ├── database.go        # DB init & connection
│   │   └── schema.sql         # Full schema (auto-applied on startup)
│   ├── handlers/              # HTTP route handlers
│   └── models/                # Data models
├── templates/                 # Templ components
│   ├── layout.templ           # Base layout (navbar, head, etc.)
│   ├── home.templ             # Dashboard
│   ├── products.templ         # Products page
│   └── partials/              # htmx partial fragments
├── static/
│   ├── css/app.css            # Custom styles
│   └── images/                # Static images
└── data/                      # SQLite database (gitignored, auto-created)
```

## Development

If using VS Code / Kiro, hit **F5** to build and launch (includes `templ generate` as a pre-launch task).

Otherwise:

```bash
# Watch templ files for changes (in one terminal)
templ generate --watch

# Run the server (in another terminal)
go run server.go
```

## Docker

If you'd like to run Domicile as a Docker container, a Dockerfile is included. For full setup instructions, configuration options, and a ready-to-use `compose.yml`, see the [Docker documentation](https://zodiac1913.github.io/Domicile/installation).

Quick start:

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
    environment:
      - DOMICILE_NO_BROWSER=1
```

```bash
docker compose up -d
```

## Inspiration

Domicile wouldn't exist without these great projects:

- [Grocy](https://grocy.info) — The original self-hosted groceries & household management app
- [Mealie](https://mealie.io) — A beautiful self-hosted recipe manager

Thanks to both teams for paving the way.

## License

MIT
