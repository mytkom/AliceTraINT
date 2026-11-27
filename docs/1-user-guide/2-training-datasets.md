# Training datasets

A training dataset is a named list of simulated AOD files that training tasks use as input. The list is fixed once submitted, so the files behind every trained model stay known.

To create a dataset, open **Datasets** and click **Create Training Dataset**. In the directory list on the left, open a production or one of its runs, for example `/alice/sim/2024/LHC24d1/`. Both ways of choosing files below search that directory.

## Selecting files across runs

Use **Choose random subset** on the right when the dataset should cover several runs. Enter the settings below and click **Select Random Subset**.

![Choose random subset with Run count 3 and Files per run 2 for the production LHC26e9](/docs/static/1-user-guide/2-training-datasets/random-subset.png)

| Setting | Default | Meaning |
|---|---|---|
| Run count | 10 | Number of runs to include |
| Files per run | 5 | Number of files from each run; runs with fewer files are skipped |
| Min file size [MB] | 2048 | Smaller files are ignored; lower this value for productions with small files |

Runs and files are chosen at even intervals, so the same settings always give the same selection. You can add or remove files by hand afterwards.

## Choosing individual files

Click **Find AO2Ds** to list every AOD file below the open directory, with its period, run number, file number and size.

Click a file to select it. To remove a file, click it under **Selected files**. The selection is kept when you search another directory, so one dataset can combine several productions.

![New dataset page with four files of run 526641 of LHC24d1 selected](/docs/static/1-user-guide/2-training-datasets/files-selected.png)

## Naming and submitting

Enter a name in **Training dataset name** and click **Submit**. Dataset names are unique across all users, so choose a descriptive one such as `LHC24d1-526641`.

Submission takes a moment, because AliceTraINT looks up the production's runs to work out the validity period of models uploaded to CCDB later. The new dataset then appears in the list and can be used in a [training task](/docs/training-tasks).

## Simulated data and O-O / Ne-Ne

Training uses simulated data only. Oxygen-oxygen and neon-neon productions need a different producer configuration; AliceTraINT recognises them from the production's description, so there is no setting for it.

## Viewing and removing datasets

**Datasets** lists every dataset with its periods, number of files, total size and author. Tick **Show only mine** to see only your own. Click a name to see the dataset's files grouped by period and run.

Click **Remove** to remove one of your own datasets. Training tasks that already use it keep it, but it can no longer be chosen for new tasks, and its name stays taken.

## Troubleshooting

| Message | Solution |
|---|---|
| `Name must be unique` | Choose another name |
| `… is too high for this path` | Lower the run count, files per run or minimum file size, or open a larger directory |
| `… external service is unreachable` | A CERN service did not respond in time; try again later, or open a smaller directory |
