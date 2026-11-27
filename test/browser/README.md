# Browser checks and frontend baseline

From the repository root on Linux, install Go as specified by `go.mod`, Node.js
22, PostgreSQL 16 server binaries and Cypress's Electron system libraries. Use
an unprivileged account with `initdb`, `pg_ctl`, and `createdb` on PATH or
discoverable through `pg_config`.

```sh
npm ci
npm run test:e2e
npm run test:e2e -- --spec cypress/e2e/training-datasets.cy.js
```

The launcher creates and removes a fresh PostgreSQL cluster under
`/tmp/alicetraint-browser.*`. PostgreSQL listens on a private Unix socket,
and the fixture app listens on a free loopback port. The app verifies the
cluster location before running migrations and seeds. It serves fixture
files from the same temporary root. The server uses the real Go handlers,
repositories, templates and HTMX, with local fakes for CERN SSO, JAliEn,
MonALISA and CCDB. No app `.env`, CERN credentials or GPU is needed.

For manual browser checks, `make run-test` starts the fixture app and
prints its URL. Stop it with Ctrl+C.

Cypress writes screenshots and videos under `artifacts/browser/`. Server
and PostgreSQL logs are under `artifacts/browser/e2e/`. GitHub Actions
uploads these on browser job failure.

## Performance baseline

Install Chrome and run the baseline separately from Cypress to avoid
competition for CPU. Set `CHROME_BIN` if Chrome has a different executable
name. The launcher runs three repetitions by default.

```sh
npm run test:baseline
BASELINE_REPETITIONS=1 CHROME_BIN=chromium npm run test:baseline
```

The JSON report is saved to `artifacts/browser/baseline/results.json`. It
covers the first two pages of a 1,000-task filtered list, two AOD directories, a 6,000-file subset that refreshes the larger
directory, and two tasks with 180 and 738 plot URLs. The task-list check requires 25 rows and one list request under 30 KB per page.
The first-task-frame proxy is two animation frames after task rows appear.
The AOD metadata is
generated in Go. Each plot URL serves a 640×480 synthetic PNG through the
real authenticated file route. The payload is the same for all plots.

Chrome runs headless at 1920×1080 over loopback with its HTTP cache disabled.
Directory navigation finishes before each timed action. The first repetition
fills the server's AOD response cache. Later requests to the same route can
use that cache. AOD payloads come from Resource Timing and exclude HTTP
headers. The report includes the number of elements in the document and
records each action's XHR, each gallery's image responses, and image
elements. The AOD first-frame proxy is two animation frames after HTMX
settles. For the gallery it is two frames after the first carousel image
decodes. A file selection and a click on the next gallery image use the
same frame proxy. Gallery page load completion is recorded separately.

These timings are informational. Unmeasured cases include live CERN
latency, WAN or ingress behavior, warm gallery reloads, frame rate while
scrolling, or the cost of selecting files across directories.

### Recorded sample

2026-11-27. NixOS Linux on an AMD Ryzen 5 PRO 3500U with 21 GiB RAM, Chrome
156.0.8078.79, Node 22.14.0, Go 1.26.10, and PostgreSQL 16.4. Counts and
payloads were identical across three repetitions. Times are medians.
MB and KB use decimal units.

| Dataset action | Available / selected rows after action | Document elements | XHR count | XHR bodies | First available frame | First selected frame | Single select |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Find 10,581 files | 10,581 / 0 | 105,924 | 1 | 10.35 MB | 3.54 s | — | 401 ms |
| Find 13,355 files | 13,355 / 0 | 133,664 | 1 | 13.10 MB | 4.87 s | — | 443 ms |
| Select a 6,000-file subset | 7,355 / 6,000 | 133,664 | 2 | 5.87 + 13.10 MB | 7.52 s | 2.48 s | 309 ms |

| Plots | Document elements | Image elements | Image requests | Image bodies | HTML body | First plot frame | Page load | Next plot |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 180 | 1,009 | 360 | 180 | 4.25 MB | 137 KB | 200 ms | 530 ms | 33 ms |
| 738 | 3,799 | 1,476 | 738 | 17.43 MB | 532 KB | 458 ms | 2.17 s | 33 ms |
