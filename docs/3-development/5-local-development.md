# Local development

You can run the web app on your own computer with Docker Compose, directly with Go, or with a test login that skips CERN SSO. Every way needs a GRID certificate, because the web app does not start without one.

## Requirements

- Go 1.25 and a C compiler (the tests use SQLite through cgo).
- Docker with Docker Compose.
- A GRID certificate exported without a password as `gridCertificate.p12`.
- For a real login: an application in the [CERN Application Portal](https://application-portal.web.cern.ch/) with `http://localhost:8088/callback` as a redirect URI. Membership of `wut-alice-pidml` gives access to the production application.
- To change templates or CSS: the [Tailwind CSS standalone CLI](https://github.com/tailwindlabs/tailwindcss/releases/tag/v3.4.10), version 3.4.10, the version the Docker image uses.

Every environment variable is described on [Configuration](/docs/configuration).

## Running with Docker Compose

Docker Compose starts PostgreSQL, the web app and a Tailwind watcher that rebuilds `static/css/output.css` when templates change. Create `.env` from the example and fill in `CERN_CLIENT_ID` and `CERN_CLIENT_SECRET`:

```bash
cp .env.example .env
```

Copy your certificate into the repository root as `gridCertificate.p12`, then build and start the containers:

```bash
docker compose up --build
```

The web app is at `http://localhost:8088`. The image build copies `.env` and the certificate into the image, so rebuild after changing either. Templates and static files are mounted into the container, but Go code and docs are not, so rebuild after changing them too.

> **Do not push this image to a registry.** It contains your GRID private key and the `.env` file with the SSO client secret.

## Running without Docker

Start only the database from Docker Compose:

```bash
docker compose up -d postgres
```

Create `.env` as above. The web app needs the certificate as two PEM files; this converts it into `~/.globus`:

```bash
mkdir -p ~/.globus
openssl pkcs12 -clcerts -nokeys -in gridCertificate.p12 -out ~/.globus/usercert.pem -password pass:
openssl pkcs12 -nocerts -nodes -in gridCertificate.p12 -out ~/.globus/userkey.pem -password pass:
chmod 400 ~/.globus/userkey.pem
```

Point `.env` at the two files:

```ini
GRID_CERT_PATH=/home/<user>/.globus/usercert.pem
GRID_KEY_PATH=/home/<user>/.globus/userkey.pem
```

Start the web app from the repository root:

```bash
make run
```

The web app migrates the database on startup. On the first start without `JALIEN_CERT_CA_DIR`, it clones the ALICE grid CA certificates into `~/.globus/certificates`.

`static/css/output.css` is committed, so you only rebuild it after changing templates or `static/css/input.css`. Use `make css` once or `make css-watch` while editing.

## Running without CERN SSO

`make run-test` starts the web app with a test login and seeded data, which is the quickest way to work on pages:

```bash
make run-test
```

It creates a separate database named after `DB_NAME` with `_test` added, fills it with two users and example datasets, machines and tasks, and drops it when you stop the web app with Ctrl+C. An existing database with that name is dropped first.

Open `http://localhost:8088/login` to log in as the first user, Albert Einstein, or `http://localhost:8088/login?userId=2` to log in as Niels Bohr. The test mode replaces only the login: it still needs the database, the GRID certificate and network access to JAliEn, MonALISA and CCDB for the pages that use them.

## Tests and linting

Run the Go tests:

```bash
make test
```

The tests use an in-memory SQLite database and mocked external services, so they need no PostgreSQL and no network. `internal/service/docs_service_test.go` also checks every internal link, anchor and image in `docs/`.

The Cypress tests in `cypress/e2e/` drive a browser against `make run-test` at `http://localhost:8088`. Start `make run-test`, then run `cypress run` from the repository root.

Lint with `make lint`. It needs golangci-lint; CI uses version 1.64.8:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
make lint
```

GitHub Actions runs the Go tests and golangci-lint on every pull request to `master`. Cypress does not run in CI.

## Project layout

| Path | Contents |
|---|---|
| `cmd/AliceTraINT/` | The web app's entry point |
| `cmd/AliceTraINT_test/` | The web app with the test login and seeded data (`make run-test`) |
| `cmd/aod_batcher/` | Command-line tool that writes batches of AOD file lists for dataset creation |
| `cmd/chroma-css/` | Generates `static/css/chroma.css` (`make chroma-css`) |
| `internal/` | The web app's packages; see [Architecture](/docs/architecture) |
| `web/templates/` | HTML templates |
| `web/nn_architectures/` | Definition of the training configuration form |
| `static/` | CSS, JavaScript and images |
| `docs/` | These docs |
| `test/` | Integration tests of handlers, services and repositories |
| `cypress/` | Browser tests |

## Contributing

Create a branch from `master` and open a pull request against `master`. The tests and the linter must pass.

Update the docs in the same pull request as the code they describe. This includes diagrams: a change to a table, a route or the middleware updates the diagram that shows it. [Updating docs](/docs/updating-docs) explains how.
