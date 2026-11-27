# Running on the EPN farm

On the ALICE EPN farm, the training module runs as a Slurm job on a GPU node instead of in Docker. It uses O2Physics and xjalienfs from CVMFS, so nothing has to be built except the module itself. This setup needs the `epn-worker` branch of the training module.

Register a training machine for the worker first, as described in [Setting up a training machine](/docs/setting-up-a-training-machine#sec-registering-the-machine). One registered machine corresponds to one worker job.

With this branch, tasks may also leave **Use subsample** unticked; the training module then uses the full producer output.

## Requirements

- An account on the EPN farm, with access to the login node and to the partition with GPU nodes.
- Your GRID certificate and key in `~/.globus/usercert.pem` and `~/.globus/userkey.pem`. The worker uses them to get a JAliEn token and download AOD files.
- An SSH key added to your GitHub account on the EPN, for the training code submodule.

## Installing

Do these steps once, on the login node, in your home directory.

Install Go, if `~/go` does not exist yet:

```bash
cd ~
wget https://go.dev/dl/go1.22.10.linux-amd64.tar.gz
tar -xzf go1.22.10.linux-amd64.tar.gz
```

Clone the `epn-worker` branch with its submodule and build the module:

```bash
git clone --branch epn-worker --recurse-submodules https://github.com/mytkom/AliceTraINT_pidml_training_module ~/AliceTraINT_pidml_training_module
cd ~/AliceTraINT_pidml_training_module
export PATH=$HOME/go/bin:$PATH
make build
```

Create the Python environment for the training code. The EPN GPUs need the ROCm build of PyTorch, so the environment is called `.venv-rocm`:

```bash
module load Python/v3.10.19-7
python3 -m venv .venv-rocm
source .venv-rocm/bin/activate
pip install -r pdi/requirements.txt
<ROCm PyTorch install command>
```

## Settings

Create `.env` from the example and set the values below. The other variables are described in [Settings](/docs/setting-up-a-training-machine#sec-settings) and can keep their defaults.

```ini
ALICETRAINT_BASE_URL=https://alicetraint.app.cern.ch
MACHINE_ID=<machine id>
MACHINE_SECRET_KEY=<secret key>
ALICETRAINT_VENV_DIR_PATH=./.venv-rocm
ALICETRAINT_ALIENV_BIN=/cvmfs/alice.cern.ch/bin/alienv
ALICETRAINT_O2PHYSICS_TAG=VO_ALICE@O2Physics::<tag>
ALICETRAINT_XJALIENFS_TAG=VO_ALICE@xjalienfs::<tag>
```

| Variable | Default | Meaning |
|---|---|---|
| `ALICETRAINT_VENV_DIR_PATH` | none | Python environment that runs the training; `./.venv-rocm` on the EPN |
| `ALICETRAINT_ALIENV_BIN` | `alienv` | The `alienv` command that loads packages from CVMFS |
| `ALICETRAINT_O2PHYSICS_TAG` | `O2Physics/latest` | O2Physics package used by the PIDML producer and the subsampling helper |
| `ALICETRAINT_XJALIENFS_TAG` | `xjalienfs/latest` | xjalienfs package used to download AOD files |
| `ALICETRAINT_TRAINING_DEVICE` | unset | Set to `cpu` to train without a GPU |

Pick an O2Physics tag that exists on CVMFS; the newest ones are listed by:

```bash
ls /cvmfs/alice.cern.ch/el9-x86_64/Modules/modulefiles/O2Physics | sort | tail
```

List xjalienfs tags the same way, in the `xjalienfs` directory next to `O2Physics`. Old tags are removed from CVMFS after a while; when the worker fails because a tag is gone, choose a new one.

## Starting the worker

Submit the worker job from the repository directory on the login node:

```bash
sbatch scripts/epn-worker.sbatch
```

The job asks for one node with one GPU and 32 GB of memory on the `dev` partition, for at most one day. It gets a JAliEn token, builds the subsampling helper and starts the module, which then asks AliceTraINT for work like any training machine.

Before it starts the module, the job submits its own successor, which starts when the job ends. So the worker keeps running day after day until you stop it. A lock file, `.epn-worker.lock`, makes sure that only one worker runs from the directory at a time.

Check that the job is running:

```bash
squeue -u $USER
```

> **A task that is still running when the one-day limit is reached is interrupted.** It stays in Training, and the next worker job takes a new task. Run the task again with **Reuse in new task**.

## Logs

Each job writes `alicetraint-worker-<job id>.out` and `.err` in the repository directory, and deletes the logs of earlier jobs when it starts. Follow the logs of the running job with:

```bash
jid=$(squeue -u $USER -n alicetraint-worker -t R -h -o %i | head -1)
tail -f alicetraint-worker-${jid}.out alicetraint-worker-${jid}.err
```

> **Keep the repository directory private.** The worker logs and `.env` contain the machine's secret key.

The results of the last task stay in the data and results folders until the next task starts, so you can inspect them after a task ends.

## Stopping the worker

Create the stop file, so that the successor job exits, and cancel the running job:

```bash
touch .epn-worker-stop
scancel -n alicetraint-worker
```

Submitting the worker again with `sbatch` removes the stop file.

## Updating

Pull the new version and rebuild on the login node, then stop and start the worker:

```bash
git pull --recurse-submodules
make build
```

## Debugging on a GPU node

To try the module by hand, open an interactive shell on a GPU node and run it there:

```bash
srun -N1 -p dev --gres=gpu:1 --mem=32G --pty bash -i
set -a; source .env; set +a
export PATH=$HOME/go/bin:$PATH
module load Python/v3.10.19-7
source .venv-rocm/bin/activate
"$ALICETRAINT_ALIENV_BIN" setenv "$ALICETRAINT_XJALIENFS_TAG" -c alien-token-init
"$ALICETRAINT_ALIENV_BIN" setenv "$ALICETRAINT_O2PHYSICS_TAG" -c make subsample
make run
```

Stop the worker job first, because both would take tasks for the same machine.

## Troubleshooting

| Problem | Solution |
|---|---|
| The job ends at once with `error: missing ./bin/AliceTraINT_pidml_training_module (run make build)` | Build the module on the login node with `make build` |
| The job ends at once with `exiting: another worker holds .epn-worker.lock` | A worker is already running from this directory; check with `squeue -u $USER` |
| The job ends at once with `exiting: stop file present` | Remove `.epn-worker-stop`, or submit the worker again with `sbatch` |
| `alien-token-init` fails | Check that the GRID certificate and key are in `~/.globus` and that the certificate has not expired |
| The producer or subsampling helper fails to load O2Physics | The tag is no longer on CVMFS; set a current one in `ALICETRAINT_O2PHYSICS_TAG` |
