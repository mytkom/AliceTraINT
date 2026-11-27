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

Each model is valid in CCDB from the start of the first run to the end of the last run of the dataset's productions. AliceTraINT works out this range when the dataset is created.

## Troubleshooting

| Message | Solution |
|---|---|
| `TrainingTask's result file: … not found` | The task does not have all six models, so nothing was uploaded. Check the training logs and run the task again |
| `… external service is unreachable` | CCDB or MonALISA did not respond; try again later |
