# Results and CCDB upload

A training task's page collects everything the training module sent back: plots, ONNX models and logs. From there, a completed task's models can be uploaded to CCDB, where O2Physics analyses read them.

Open **Tasks** and click a task's name. Results appear step by step while the task runs: the logs of each step when it ends, the ONNX files after training and the plots at the end. A failed task shows only its logs.

## Plots

**Image Results Gallery** shows every image the training module uploaded. Click a name in the list to show the plot, or use the arrows to move through them.

![Image Results Gallery showing the purity of protons against momentum](/docs/static/1-user-guide/4-results-and-ccdb-upload/gallery.png)

For each particle, the gallery has:

- the model's performance against momentum: purity, efficiency and F1 score, and a precision-recall curve;
- the importance of each input variable, for each combination of detectors (for example `TPC_TOF.png`);
- the distributions of the input variables, as histograms, box plots, density plots and 2D plots;
- the share of each particle species and of tracks with missing detectors.

The plots of every particle have the same file names, so the title inside a plot, such as "Purity vs p (proton)", says which particle it belongs to.

## ONNX files

**ONNX Files** lists one model per particle, six in total. Click **Download** to save a model.

![ONNX Files with six models and Upload to CCDB, and Log Files with two logs for each step](/docs/static/1-user-guide/4-results-and-ccdb-upload/onnx-and-logs.png)

| File | Particle | Name in CCDB |
|---|---|---|
| `pion.onnx` | π⁺ | `simple_model_211` |
| `antipion.onnx` | π⁻ | `simple_model_0211` |
| `kaon.onnx` | K⁺ | `simple_model_321` |
| `antikaon.onnx` | K⁻ | `simple_model_0321` |
| `proton.onnx` | p | `simple_model_2212` |
| `antiproton.onnx` | p̄ | `simple_model_02212` |

The number in the CCDB name is the particle's PDG code, with a leading `0` for the antiparticle.

## Logs

**Log Files** has two logs, standard output and standard error, for each step the training module runs. Click **Show** to open a log in the browser or **Download** to save it.

| Log | Step |
|---|---|
| `dataset_out.log`, `dataset_err.log` | Preparing the dataset: downloading the AOD files, running the PIDML producer and subsampling |
| `pdi_train_out.log`, `pdi_train_err.log` | Training the models |
| `pdi_plots_out.log`, `pdi_plots_err.log` | Creating the plots |

When a task fails, the last log uploaded belongs to the step that failed. Its error log usually explains why.

## Uploading to CCDB

**Upload to CCDB** appears under the ONNX files once the task is Completed. Click it and confirm, and AliceTraINT uploads all six models to CCDB under the path the operator configured (see [Configuration](/docs/configuration#sec-ccdb)). The task's status then changes to Uploaded.

On an uploaded task the button reads **Reupload to CCDB** and uploads the six models again, with the same validity period.

> **An upload cannot be undone from AliceTraINT.** The models stay in CCDB, and analyses that read this path can pick them up for runs in their validity period. Any logged-in user can upload any completed task, so check the plots before you upload.

## Validity period

Each model is valid in CCDB from the start of the first run to the end of the last run of the dataset's productions. AliceTraINT takes the list of runs from MonALISA and their start and end times from CCDB when the dataset is created, and stores the range with the dataset.

For a dataset created before AliceTraINT stored this range, the range is looked up when you upload. If MonALISA has no runs for the dataset's productions, AliceTraINT uses the lowest and highest run number of the dataset's own files instead.

## Troubleshooting

| Message | Solution |
|---|---|
| `TrainingTask's result file: kaon.onnx not found` (or another model) | The training module did not upload all six models, so nothing was uploaded. Check `pdi_train_err.log`, then run the task again |
| `"CCDB" external service is unreachable` | CCDB did not respond; try again later. Models uploaded before the error stay in CCDB, and the status does not change |
| `"Monalisa" external service is unreachable` | Only for datasets without a stored range: MonALISA did not respond; try again later |
