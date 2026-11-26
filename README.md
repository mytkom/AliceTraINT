# AliceTraINT

AliceTraINT is a web app for training particle identification (PID) machine-learning models of the ALICE experiment at CERN. Users build training datasets from AOD files on the GRID, queue training tasks that run on registered training machines, review the results and upload the trained models to CCDB.

The training itself runs in a separate program, the [training module](https://github.com/mytkom/AliceTraINT_pidml_training_module).

## Documentation

The documentation is served by the app at [alicetraint.app.cern.ch/docs](https://alicetraint.app.cern.ch/docs) and is also readable on GitHub in [`docs/`](docs/). It covers using the app, setting up a training machine, the architecture, local development and operations.

## Quick start

You need:

- Docker with Docker Compose,
- an application in the [CERN Application Portal](https://application-portal.web.cern.ch/) with `http://localhost:8088/callback` as a redirect URI,
- a GRID certificate exported without a password.

Create `.env` and set `CERN_CLIENT_ID` and `CERN_CLIENT_SECRET` from your application:

```bash
cp .env.example .env
```

Copy the certificate into the repository root as `gridCertificate.p12`, then start the app and its database:

```bash
docker compose up --build
```

The app is at `http://localhost:8088`. The image contains your certificate and `.env`, so keep it local. [Local development](docs/3-development/5-local-development.md) covers running without Docker, the test login, tests and linting.

## Contributing

Open a pull request against `master`. Tests and the linter run in CI, and docs and diagrams are updated in the same pull request as the code they describe.

## License

GNU General Public License v3.0. See [LICENSE](LICENSE).
