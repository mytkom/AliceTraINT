# Running on the EPN farm

The training module can also run on the ALICE EPN farm. There it runs as a Slurm job on a GPU node instead of in Docker, and it takes O2Physics and the GRID tools from CVMFS, so only the module itself has to be built. This setup uses the `epn-worker` branch of the training module.

## How it works

The worker is a Slurm batch job that you submit from the EPN login node. The job prepares the environment, gets a JAliEn token with your GRID certificate and starts the training module, which then asks AliceTraINT for tasks like any other training machine.

A job may run for at most one day. Before the module starts, each job submits the next one, so the worker keeps running until you stop it. Everything lives in the repository directory in your home: the code, the settings, the dataset cache and the job logs.

## Setting up

You need an EPN account with access to the GPU nodes, your GRID certificate and key in `~/.globus`, and a registered training machine ([Registering the machine](/docs/setting-up-a-training-machine#sec-registering-the-machine)).

On the login node:

1. Install Go in your home directory, if it is not there yet.
2. Clone the `epn-worker` branch of the [training module](https://github.com/mytkom/AliceTraINT_pidml_training_module) with its submodules, and build the module with `make build`.
3. Create a Python environment called `.venv-rocm` with `module load Python/v3.10.19-7`, install `pdi/requirements.txt` into it, and replace PyTorch with its ROCm build, which the EPN GPUs need.
4. Copy `.env.example` to `.env` and fill it in, as described below.

The repository's `EPN_SETUP.md` lists the commands for these steps.

## Settings

Besides the machine ID, the secret key and the address of AliceTraINT ([Settings](/docs/setting-up-a-training-machine#sec-settings)), the EPN worker needs to know which Python environment and which CVMFS packages to use:

```ini
ALICETRAINT_BASE_URL=https://alicetraint.app.cern.ch
MACHINE_ID=<machine id>
MACHINE_SECRET_KEY=<secret key>
ALICETRAINT_VENV_DIR_PATH=./.venv-rocm
ALICETRAINT_ALIENV_BIN=/cvmfs/alice.cern.ch/bin/alienv
ALICETRAINT_O2PHYSICS_TAG=VO_ALICE@O2Physics::<tag>
ALICETRAINT_XJALIENFS_TAG=VO_ALICE@xjalienfs::<tag>
```

The O2Physics and xjalienfs tags must exist on CVMFS. Old tags are removed after a while, so pick a recent one and update it when the worker can no longer load it.

## Running the worker

Submit the worker from the repository directory:

```bash
sbatch scripts/epn-worker.sbatch
```

Check it with `squeue -u $USER`. Each job writes its output to `alicetraint-worker-<job id>.out` and `.err` in the repository directory.

To stop the worker, create the file `.epn-worker-stop` so that no new job starts, and cancel the running one with `scancel -n alicetraint-worker`. To update, stop the worker, pull the new version, run `make build` and submit the worker again.

> **Keep the repository directory private.** Its `.env` and the job logs contain the machine's secret key.
