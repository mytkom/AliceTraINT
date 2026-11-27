## Production deployment

AliceTraINT is deployed on **CERN PaaS** (Red Hat OpenShift–based) with a **PostgreSQL** database provided by **CERN DB-on-Demand (DBoD)**.

Operators with access to the AliceTraINT project use:

- [PaaS web console](https://paas.cern.ch/)
- [DB-on-Demand](https://dbod.web.cern.ch/) (after subscribing in the [CERN Resource portal](https://auth-resources.web.cern.ch/))

The deployment follows the official [PaaS documentation](https://paas.docs.cern.ch/); this document explains how that applies specifically to AliceTraINT.

---

## High-level architecture

- **Application container**
  - Fetched from [GitHub](https://github.com/mytkom/AliceTraINT) and built from the repository Dockerfile (`Dockerfile`) as an OpenShift image.
  - To rebuild one need to click on container in Topology view and build new image.
  - Two-stage build: Go/Tailwind builder, then a hardened UBI9 runtime image.
  - The container runs a single Go binary (`AliceTraINT`) and serves the web UI and docs on port **8088**.

- **Database (DBoD PostgreSQL)**
  - PostgreSQL instance provisioned via DB-on-Demand.
  - Application connects using credentials stored in OpenShift secrets and injected as environment variables.

- **Persistent storage (EOS)**
  - A persistent volume backed by EOS storage of a service account, for training results and application data.
  - An additional EOS directory for larger data, accessed by the same service account.

- **Authentication / GRID certificates**
  - The GRID certificate is stored as an OpenShift secret and used by the image build.
  - EOS access uses a secret with the service account's credentials.

---

## Prerequisites

- **Access / roles**
  - Access granted by the current operators to:
    - PaaS project where AliceTraINT is deployed.
    - DB-on-Demand instance used by AliceTraINT.
    - EOS locations used for persistent volumes.

- **Required secrets in the PaaS project**
  - A binary secret with the `.p12` GRID certificate, mounted in the **BuildConfig** so the Docker build can read it via `CERT_PATH`.
  - A secret with the EOS credentials of the service account, mounted in the **Deployment** pods following the [PaaS EOS guidelines](https://paas.docs.cern.ch/3._Storage/eos/).
  - A secret with the database connection settings.

---

## Application image build (Dockerfile)

The application image is built from the repository `Dockerfile` using a two-stage Docker build (Go/Tailwind builder and UBI9 runtime). For full details see the `Dockerfile` in the repository; in production the only AliceTraINT-specific requirement is that:

- The BuildConfig uses this `Dockerfile` as-is.
- The GRID certificate secret is mounted into the build pod and its path is passed as the `CERT_PATH` build argument.
- The resulting image listens on port **8088** and embeds the docs and static assets.

All lower-level steps (exact base images, build tools, Tailwind invocation, certificate conversion) are encoded in the `Dockerfile` itself and normally do not need to be modified for deployment.

---

## OpenShift / PaaS resources

At a minimum, the deployment uses:

- **BuildConfig**
  - Source: this Git repository (AliceTraINT).
  - Strategy: Docker build using the `Dockerfile` from the repo root.
  - Mounts the GRID certificate secret into the build pod and sets the `CERT_PATH` build argument to its path.

- **ImageStream**
  - Stores the built AliceTraINT image versions.
  - The Deployment/DeploymentConfig uses the ImageStream tag as its image reference.

- **Deployment (or DeploymentConfig)**
  - References the ImageStream tag produced by the BuildConfig.
  - Mounts:
    - EOS-backed persistent volume(s).
    - EOS credentials secret.
    - Database credentials secret.
  - Sets environment variables:
    - Database connection settings (host, port, DB name, user, password, SSL mode if used).
    - Application URLs such as JAliEn URL and CCDB URL.
    - Any app-specific configuration (e.g. log level, base paths, etc.).
    - Path(s) to user certificates, when needed (even though the Dockerfile already sets `GRID_CERT_PATH` and `GRID_KEY_PATH`).

- **Service and Route**
  - **Service**
    - Cluster-internal service pointing to port 8088 of the AliceTraINT pods.
  - **Route**
    - Exposes the service externally within CERN’s network.
    - TLS configuration depends on project/PaaS policies.

---

## Persistent storage and EOS integration

### Primary EOS volume

- A persistent volume is provisioned using EOS storage of a service account, following the [PaaS EOS storage guidelines](https://paas.docs.cern.ch/3._Storage/eos/).
- It holds the results that training modules upload and other application data.

### Additional EOS storage

- A second EOS directory, which the service account can access, provides more space for larger data.
- EOS authentication and mounting use the credentials stored in the EOS credentials secret.

Inside the container, EOS is mounted under `/eos`, and both areas are reached through the standard CERN EOS layout.

---

## End-to-end deployment steps (summary)

1. **Prepare access**
   - Ensure you have access to the relevant PaaS project and DB-on-Demand instance.

2. **Create and configure secrets**
   - Create a secret with the `.p12` GRID certificate.
   - Create a secret with the EOS credentials of the service account.
   - Create a secret with the database connection settings for the DBoD PostgreSQL instance.

3. **Configure storage**
   - Set up EOS-backed persistent volumes and persistent volume claims as described in the [PaaS EOS docs](https://paas.docs.cern.ch/3._Storage/eos/).
   - Mount:
     - The EOS volume and, if separate, the additional EOS directory into the pods.

4. **Configure BuildConfig and ImageStream**
   - Create a BuildConfig pointing to the AliceTraINT Git repo.
   - Use Docker build strategy with the repo `Dockerfile`.
   - Mount the GRID certificate secret into the build pod and set the `CERT_PATH` build argument to the mounted path.
   - Configure an ImageStream to hold built images.

5. **Configure Deployment (or DeploymentConfig)**
   - Reference the ImageStream tag built by the BuildConfig.
   - Mount:
     - EOS volume(s) at the paths expected by the application.
     - EOS credentials secret.
     - Database credentials secret.
   - Set all required environment variables:
     - DB connection.
     - JAliEn and CCDB endpoints.
     - Any additional runtime configuration.
   - Configure resource requests/limits and liveness/readiness probes on port 8088.

6. **Expose the service**
   - Create a Service for port 8088.
   - Create a Route to expose the service externally, with appropriate TLS settings according to CERN PaaS policies.

7. **Verify deployment**
   - Check pod logs for startup/migration errors.
   - Verify:
     - Database connectivity (e.g. by triggering a simple training or metadata operation).
     - EOS mounts (data is written to and read from the expected directories).
     - Access to JAliEn/CCDB via the configured URLs.

---

## Runtime configuration (environment variables)

At runtime, AliceTraINT reads configuration from environment variables (and optionally from a `.env` file) via `internal/config/config.go`. Below is a concise summary of the most important groups of variables used in production; see the Go file for the full list and default values.

- **Database (DBoD PostgreSQL)**
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
  - `DB_SSLMODE`, `DB_SSL_CERT_PATH`

- **Application server and caching**
  - `ALICETRAINT_PORT` (HTTP port, `8088` in production)
  - `ALICETRAINT_JALIEN_CACHE_MINUTES`

- **External services (JAliEn, CCDB)**
  - `JALIEN_HOST`, `JALIEN_WSPORT`, `JALIEN_CERT_CA_DIR`
  - `CCDB_URL`, `CCDB_UPLOAD_SUBDIR`

- **GRID certificates**
  - `GRID_CERT_PATH`, `GRID_KEY_PATH`

- **Data and documentation paths**
  - `ALICETRAINT_DATA_DIR_PATH`
  - `ALICETRAINT_NN_ARCH_DIR`
  - `ALICETRAINT_DOCS_DIR_PATH`

In production, these values are normally provided via OpenShift secrets and config on the Deployment, and they must be consistent with the EOS mounts and paths created by the `Dockerfile` and PaaS storage configuration.