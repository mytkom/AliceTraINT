# Setting up a training machine

A training machine is a computer that runs the training module: it takes training tasks from AliceTraINT, prepares their datasets, trains the models and uploads the results. Anyone in the AliceTraINT group can register a machine and becomes its machine owner.

This page runs the module in Docker. To run it as a Slurm job on the ALICE EPN farm instead, see [Running on the EPN farm](/docs/running-on-epn).

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

The settings are in `.env`, which the image copies when it is built, so rebuild the image after changing it. Besides the ID, the key and the address of AliceTraINT, you may want to change these:

| Variable | Value in `.env.example` | Meaning |
|---|---|---|
| `ALICETRAINT_POOLING_WAIT_SECONDS` | `20` | Seconds between two requests for work while the machine is idle |
| `ALICETRAINT_DATASET_CACHE_DIR_PATH` | `./dataset_cache` | Directory of the dataset cache; mount it as a volume |

The other variables in `.env.example` are paths inside the container and keep their defaults.

## Checking that it works

On **Machines**, **Last activity at** of your machine changes every time the module asks for work. The container's log, from `docker logs alicetraint-module`, shows a request to `/training-machines/<machine id>/training-task` at the same interval.

## Dataset cache

The module keeps every dataset it prepares in the cache, so a later task on the same files skips the preparation. The cache only grows; you can delete its files while the module is idle, and a task that needs a deleted dataset prepares it again.

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

If the container stops or tasks fail, read the container's log with `docker logs alicetraint-module` and the logs of the failed task. The usual causes are a wrong machine ID or secret key, a missing value in `.env` and an expired GRID certificate; fix `.env` or export a new `gridCertificate.p12`, then rebuild the image.
