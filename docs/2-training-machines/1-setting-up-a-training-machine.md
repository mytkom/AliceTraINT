# Setting up a training machine

A training machine is a computer that runs the training module: it takes training tasks from AliceTraINT, prepares their datasets, trains the models and uploads the results. Anyone in the AliceTraINT group can register a machine and becomes its machine owner.

## Requirements

- A Linux machine with Docker.
- Plenty of disk space: the image builds O2Physics, and prepared datasets are kept in a cache on the machine.
- For training on a GPU: an NVIDIA GPU with the NVIDIA Container Toolkit. Without one, training runs on the CPU.
- A GRID certificate, exported without a password. The module uses it to download AOD files.
- An SSH key added to your GitHub account, because the training code is a Git submodule with an SSH address.
- Outbound network access to AliceTraINT, the GRID and CCDB. AliceTraINT never connects to the machine, so no incoming ports are needed.

## Registering the machine

Open **Machines**, click **Register Training Machine**, enter a name and click **Submit**. AliceTraINT shows the machine's ID and secret key.

![ID of the machine and its secret key after registration](/docs/static/2-training-machines/1-setting-up-a-training-machine/machine-id-and-key.png)

Copy both now. The secret key is shown only once and cannot be recovered; if you lose it, register the machine again under a new name.

## Starting the training module

Clone the repository together with the submodule that contains the training code:

```bash
git clone --recurse-submodules https://github.com/mytkom/AliceTraINT_pidml_training_module
cd AliceTraINT_pidml_training_module
cp .env.example .env
```

Set the following values in `.env` and keep the remaining defaults:

```ini
ALICETRAINT_BASE_URL=https://alicetraint.app.cern.ch
MACHINE_ID=<machine id>
MACHINE_SECRET_KEY=<secret key>
```

Save your GRID certificate in the repository directory as `gridCertificate.p12`. Then build the image; building O2Physics takes a long time:

```bash
docker build -t alicetraint/training-module .
```

Start a container. The `-v` option keeps prepared datasets outside the container, so containers created later can reuse them:

```bash
docker run -d --name alicetraint-module \
  -v <cache directory>:/wd/alice/AliceTraINT_pidml_training_module/dataset_cache \
  alicetraint/training-module
```

To train on a GPU, add `--gpus all` before the image name.

> **Keep the image and the container's logs private.** The image contains your GRID certificate and the machine's secret key.

## Settings

The image copies `.env` when it is built, so rebuild the image after changing it. Every variable must be set; the module stops at startup when one is missing.

| Variable | Value in `.env.example` | Meaning |
|---|---|---|
| `ALICETRAINT_BASE_URL` | `http://localhost:8088` | Address of AliceTraINT; `https://alicetraint.app.cern.ch` for production |
| `MACHINE_ID` | `2` | ID shown at registration |
| `MACHINE_SECRET_KEY` | placeholder | Secret key shown at registration |
| `ALICETRAINT_POOLING_WAIT_SECONDS` | `20` | Seconds between two requests for work while the machine is idle |
| `ALICETRAINT_DATASET_CACHE_DIR_PATH` | `./dataset_cache` | Directory of the dataset cache; mount it as a volume |
| `ALICETRAINT_DATA_DIR_PATH` | `./data` | Working directory for downloaded AOD files and the training configuration; emptied before every task |
| `ALICETRAINT_RESULTS_DIR_PATH` | `./results` | Working directory for logs, models and plots; emptied before every task |
| `ALICETRAINT_SCRIPTS_DIR_PATH` | `./scripts` | Scripts that download, produce and subsample the data |
| `ALICETRAINT_VENV_DIR_PATH` | `./.venv` | Python environment of the training code |
| `ALICETRAINT_PDI_SRC_DIR_PATH` | `./pdi/src` | Source of the training code |

Relative paths are inside `/wd/alice/AliceTraINT_pidml_training_module` in the container. Only the first four variables normally need changing.

## Checking that it works

On **Machines**, **Last activity at** of your machine changes every time the module asks for work. The container's log, from `docker logs alicetraint-module`, shows a request to `/training-machines/<machine id>/training-task` at the same interval.

## Dataset cache

The module keeps every dataset it prepares as one ROOT file in the cache, with a small JSON file next to it. A later task with the same files and the same event count skips preparation. Its size grows with every new dataset, and the module never deletes anything from it.

You can delete files from the cache while the module is not running a task. A task that needs a deleted dataset prepares it again.

## Updating

Pull the new version, rebuild the image and replace the container, keeping the same cache volume:

```bash
git pull --recurse-submodules
docker build -t alicetraint/training-module .
docker rm -f alicetraint-module
docker run -d --name alicetraint-module \
  -v <cache directory>:/wd/alice/AliceTraINT_pidml_training_module/dataset_cache \
  alicetraint/training-module
```

## Removing

Stop and remove the container with `docker rm -f alicetraint-module`, and delete the image if you do not need it. Then click **Remove** next to the machine on **Machines**. From then on its secret key is no longer accepted.

## Troubleshooting

| Problem | Solution |
|---|---|
| Cloning stops at the `pdi` submodule with `Permission denied (publickey)` | Add an SSH key to your GitHub account and clone again |
| The container stops at once and its log shows `ENV: … not present` | Set the named variable in `.env` and rebuild the image |
| The container stops and its log shows `401 Unauthorized` | `MACHINE_ID` or `MACHINE_SECRET_KEY` is wrong, or the machine was removed. Register again if the key is lost, update `.env` and rebuild |
| Tasks fail and `dataset_err.log` shows download errors | The GRID certificate may have expired. Export a new one as `gridCertificate.p12` and rebuild the image |
