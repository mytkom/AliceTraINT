# Configuration

The web app reads its configuration from environment variables when it starts. Values in a `.env` file in the working directory are loaded too, but a variable that is already set in the environment wins. The defaults below come from `internal/config/config.go` and `internal/auth/auth_service.go`.

## Database

| Variable | Default | Meaning |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `user` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `AliceTraINT_db` | Database name. `make run-test` adds `_test` to it |
| `DB_SSLMODE` | `disable` | PostgreSQL `sslmode`, for example `verify-full` |
| `DB_SSL_CERT_PATH` | empty | CA certificate used to verify the database server (PostgreSQL `sslrootcert`) |

## Web server and paths

| Variable | Default | Meaning |
|---|---|---|
| `ALICETRAINT_PORT` | `8088` | Port the web app listens on |
| `ALICETRAINT_DATA_DIR_PATH` | `data` | Directory where uploaded training task results are stored and served from at `/data/` |
| `ALICETRAINT_DOCS_DIR_PATH` | `docs` | Directory with these docs, read once at startup |
| `ALICETRAINT_NN_ARCH_DIR` | `web/nn_architectures/proposed.json` | JSON file that defines the training configuration form |

Relative paths are resolved from the directory the web app is started in. Templates are always read from `web/templates/` and static files from `static/`.

## CERN SSO

These four variables have no defaults, and the web app stops at startup when it cannot reach `CERN_REALM_URL`.

| Variable | Default | Meaning |
|---|---|---|
| `CERN_REALM_URL` | none | OpenID Connect realm, `https://auth.cern.ch/auth/realms/cern` |
| `CERN_CLIENT_ID` | none | Client ID of the application in the CERN Application Portal |
| `CERN_CLIENT_SECRET` | none | Client secret of that application |
| `CERN_REDIRECT_URL` | none | URL that CERN SSO returns to after login: the web app's address followed by `/callback`, for example `http://localhost:8088/callback` |

The redirect URL must also be registered in the application's settings in the [CERN Application Portal](https://application-portal.web.cern.ch/).

## JAliEn and MonALISA

| Variable | Default | Meaning |
|---|---|---|
| `JALIEN_HOST` | `alice-jcentral.cern.ch` | JAliEn central services host |
| `JALIEN_WSPORT` | `8097` | JAliEn WebSocket port |
| `JALIEN_CERT_CA_DIR` | empty | Directory with the ALICE grid CA certificates ([alien-cas](https://github.com/alisw/alien-cas)). When empty, the web app uses `~/.globus/certificates`, cloning alien-cas there if it is missing, then the usual grid locations, then the system certificates |
| `JALIEN_TIMEOUT_SECONDS` | `180` | Timeout for connecting to JAliEn |
| `ALICETRAINT_JALIEN_CACHE_MINUTES` | `60` | How long JAliEn directory listings and AOD searches are kept in memory |
| `MONALISA_URL` | `https://alimonitor.cern.ch` | MonALISA base URL |

## CCDB

| Variable | Default | Meaning |
|---|---|---|
| `CCDB_URL` | `http://ccdb-test.cern.ch:8080` | CCDB base URL, used to read run times and to upload ONNX models |
| `CCDB_UPLOAD_SUBDIR` | `/Users/m/mmytkows` | CCDB path that ONNX models are uploaded under, relative to `CCDB_URL` |
| `CCDB_CA_CERT_PATH` | empty | Extra CA certificates for CCDB: one file or a directory of `.crt` and `.pem` files. Production CCDB (`https://alice-ccdb.cern.ch`) needs CERN Root CA 2 and CERN Grid CA (1) |

Set `CCDB_UPLOAD_SUBDIR` for every installation. The default points to a personal test area.

## GRID certificate

| Variable | Default | Meaning |
|---|---|---|
| `GRID_CERT_PATH` | empty | GRID user certificate in PEM format |
| `GRID_KEY_PATH` | empty | Its private key in PEM format, without a password |

The web app uses the certificate for JAliEn, MonALISA and CCDB, and stops at startup when either file cannot be read. The Docker image converts `gridCertificate.p12` into `usercert.pem` and `userkey.pem` at build time and sets both variables.
