# Deploying

CERN PaaS builds production from this repository with its `Dockerfile`, from the branch set in the build configuration, normally `master`. Deploying a new version means building a new image and letting the deployment roll it out.

## Releasing a new version

1. Merge the changes into `master`.
2. In the PaaS console, start a new build of the AliceTraINT build configuration, for example from the **Topology** view.
3. When the build finishes, the deployment replaces the running container with the new image.

The web app migrates the database itself when it starts. Sessions are kept in memory, so every user is logged out by a new deployment; release at a quiet time.

## Rolling back

Roll the deployment back to the previous image in the PaaS console. The database migrations only add tables and columns, so an older version normally runs on a newer database.

## First-time setup

A new installation needs, in a PaaS project:

1. **Secrets** for the GRID certificate, the database connection, the SSO client and the EOS credentials ([Certificates and secrets](/docs/production-overview#sec-certificates-and-secrets)).
2. **Storage**: an EOS volume for the data directory, following the [PaaS EOS documentation](https://paas.docs.cern.ch/3._Storage/eos/).
3. **A build configuration** that builds this repository's `Dockerfile`, with the GRID certificate secret mounted and its path passed as the `CERT_PATH` build argument, and an image stream for the built images.
4. **A deployment** of that image with the secrets and the EOS volume mounted and the environment variables from [Configuration](/docs/configuration) set. An HTTP readiness and liveness probe on a public page, such as `/docs`, on port 8088 lets PaaS notice a container that has stopped answering.
5. **A service and a route** that expose port 8088 over HTTPS.

Finally, register the route's address followed by `/callback` as a redirect URI of the SSO application, and set it as `CERN_REDIRECT_URL`.

## Renewing the GRID certificate

The certificate is converted and copied into the image when it is built. To renew it, replace the certificate in its secret with the new `.p12` file, exported without a password, and start a new build.

## Verifying a deployment

After a deployment, check that:

- the container's log shows `Starting server on :8088` and no errors from the migration;
- the docs open without logging in, and logging in through CERN SSO works;
- the directory list on the new-dataset page loads, which shows that the GRID certificate and JAliEn work;
- the plots of an existing training task load, which shows that the EOS area is mounted.

## Troubleshooting

When the container does not start, its log names the cause. The usual ones are a missing or unreadable GRID certificate, a CERN SSO address that cannot be reached and wrong database settings.
