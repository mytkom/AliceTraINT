# Production overview

Production AliceTraINT runs on CERN services that the operators look after: the web app on CERN PaaS, its database on DB-on-Demand, its result files on EOS and its login through CERN SSO. Training machines are not part of it; their owners run them.

## What runs where

| Component | Service | Looked after by |
|---|---|---|
| Web app | One container on [CERN PaaS](https://paas.docs.cern.ch/) (OpenShift), built from this repository's `Dockerfile` | Operators |
| Database | PostgreSQL on [DB-on-Demand](https://dbod.web.cern.ch/) | Operators |
| Result files | An EOS area mounted into the container | Operators |
| Login | An application in the [CERN Application Portal](https://application-portal.web.cern.ch/) | Operators |
| Training machines | Docker or the EPN farm, outside PaaS | Machine owners |

The web app runs as a single replica. It keeps sessions and caches in memory, so it is not meant to run as several copies.

## Accounts and access

Operators need access to the PaaS project, the DB-on-Demand instance, the EOS area and the SSO application. New operators get it from the current ones.

Users need only a CERN account in the `wut-alice-pidml` e-group, which the SSO application checks at login.

## Storage

The web app writes every file that a training task uploads (logs, plots, ONNX models) to its data directory, `ALICETRAINT_DATA_DIR_PATH`, which in production is on EOS. It serves them at `/data/` to logged-in users. The database stores only their paths, so the EOS area and the database must be backed up and restored together.

Prepared training datasets are not stored here; they stay in the dataset cache of each training machine.

## Certificates and secrets

| Item | Used for | When it is wrong or expired |
|---|---|---|
| GRID certificate | JAliEn, MonALISA and CCDB; built into the image | Browsing the GRID, creating datasets and CCDB uploads fail. Renew it and rebuild ([Renewing the GRID certificate](/docs/deploying#sec-renewing-the-grid-certificate)) |
| SSO client secret | Logging users in | Nobody can log in |
| Database credentials | The connection to PostgreSQL | The web app does not start |
| EOS credentials | Mounting the EOS area | Result files cannot be stored or shown |

All of them are kept as OpenShift secrets and passed to the build or the container. The environment variables that hold them are listed on [Configuration](/docs/configuration).
