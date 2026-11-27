# Getting started

AliceTraINT trains machine-learning models for particle identification (PIDML) in ALICE. You choose simulated AOD files, queue a training task, and a training machine prepares the data, trains the models and sends back plots and ONNX files that you can upload to CCDB.

## Access

You need a CERN account that is a member of the `wut-alice-pidml` e-group. Click **Login with CERN SSO** on the landing page; the e-group membership is checked there.

These docs are readable without logging in.

## How it works

A training run passes through four places: the GRID, where the AOD files are; the web app, where you choose them and review results; a training machine, which does the work; and CCDB, from which O2Physics analyses read the trained models. The steps from the PIDML producer to training run in the training module on the training machine.

```mermaid
flowchart TB
  accTitle: PIDML workflow
  accDescr: A training dataset lists AOD files on the GRID. A training module downloads them, runs the PIDML producer, subsamples the output into the dataset cache and trains the models. The plots and ONNX models are reviewed in the web app and uploaded to CCDB, valid for the runs of the anchored production, where O2Physics analyses read them.

  dataset[Training dataset in the web app]
  grid[(GRID)]
  producer[PIDML producer]
  subsample[Subsampling]
  cache@{ shape: docs, label: "Dataset cache" }
  train[Training]
  results[Plots and ONNX models in the web app]
  ccdb["CCDB → O2Physics analyses"]

  dataset -->|file list| producer
  grid -->|AOD files| producer
  producer --> subsample --> cache --> train --> results
  results -->|upload, valid for the anchored runs| ccdb
```

A training dataset is a fixed list of simulated AOD files. A training task trains one configuration on one dataset; it waits in a queue until a registered training machine takes it. The training module on that machine prepares the dataset, trains one model per particle and uploads the results. You then review the plots and upload the models to CCDB.

## Your first run

<!-- First-run example: replace the generic steps with the named production task, dataset and screenshot of its task page. -->

This run trains models with the default settings on a small dataset.

1. Open **Datasets** and click **Create Training Dataset**. In the directory list, open a simulated production, then click **Select Random Subset** with the default settings ([Training datasets](/docs/training-datasets)).
2. Enter a name and click **Submit**.
3. Open **Tasks** and click **Create Training Task**. Enter a name, choose your dataset and keep the default configuration. Click **Submit**; the task is Queued.
4. Wait until a training machine takes the task. Its status changes to Training, then Benchmarking and finally Completed.
5. Open the task and check the plots and logs ([Results and CCDB upload](/docs/results-and-ccdb-upload)).

Uploading the models to CCDB is the last step of a real run, and is not part of this example.

## Where to go next

- Machine owners: [Setting up a training machine](/docs/setting-up-a-training-machine).
- Developers: [Architecture](/docs/architecture) and [Local development](/docs/local-development).
- Operators: [Configuration](/docs/configuration).

The code is open source: the [web app](https://github.com/mytkom/AliceTraINT) and the [training module](https://github.com/mytkom/AliceTraINT_pidml_training_module).
