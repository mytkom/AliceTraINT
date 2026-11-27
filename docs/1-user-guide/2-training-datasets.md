# Training datasets

A training dataset is a named list of simulated AOD files that training tasks use as input. The list is fixed once submitted, so the files behind every trained model stay known.

To create a dataset, open **Datasets** and click **Create Training Dataset**. The directory list on the left starts at `/alice/sim/`. Open a production or one of its runs, for example `/alice/sim/2024/LHC24d1/`; the searches below always use the directory shown at the top of this list.

## Selecting files across runs

Use **Choose random subset** on the right when the dataset should cover several runs. Enter the settings below and click **Select Random Subset**.

![Choose random subset with Run count 3 and Files per run 2 for the production LHC26e9](/docs/static/1-user-guide/2-training-datasets/random-subset.png)

| Setting | Default | Meaning |
|---|---|---|
| Run count | 10 | Number of runs to include |
| Files per run | 5 | Number of files from each run; runs with fewer files are skipped |
| Min file size [MB] | 2048 | Smaller files are ignored; lower this value for productions with small files |

Runs and files are chosen at even intervals, so the same settings and directory always give the same selection. If files are already selected, AliceTraINT asks for confirmation before replacing them. The page then lists the remaining files of the directory, so you can add more by hand.

## Choosing individual files

Click **Find AO2Ds** to list every AOD file below the open directory, with its period, run number, file number and size. The directory must be at least a production, such as `/alice/sim/2024/LHC24d1/`; a whole year is refused.

Click a file to select it. To remove a file, click it under **Selected files**. The selection is kept when you search another directory, so one dataset can combine several productions.

![New dataset page with four files of run 526641 of LHC24d1 selected](/docs/static/1-user-guide/2-training-datasets/files-selected.png)

## Naming and submitting

Enter a name in **Training dataset name** and click **Submit**. Dataset names are unique across all users, so choose a descriptive one such as `LHC24d1-526641`.

Submission takes a moment because AliceTraINT looks up the production's runs in MonALISA and their start and end times in CCDB. The time range later becomes the validity period of models uploaded to CCDB. The new dataset then appears in the list and can be used in a training task.

## Simulated data and O-O / Ne-Ne

Training uses simulated data only. The directory list starts at `/alice/sim/`, and training modules always treat a dataset as simulated.

Oxygen-oxygen and neon-neon productions need a different producer configuration. AliceTraINT detects them when the dataset is submitted, from the production's description in MonALISA, so there is no setting for it.

## Viewing and removing datasets

**Datasets** lists every dataset with its periods, number of files, total size and author. Tick **Show only mine** to see only your own. Click a name to see the dataset's files grouped by period and run.

You can remove only your own datasets. **Remove** also appears on other users' datasets. Clicking it there hides the row, but nothing is removed and the dataset is back when the page reloads. Training tasks that already use a removed dataset keep it and still run, but it can no longer be chosen for new tasks.

The name of a removed dataset stays taken, so a new dataset cannot reuse it.

## Troubleshooting

| Message | Solution |
|---|---|
| `AODFiles missing` | Select at least one file |
| `Name must be unique` | Choose another name; names of removed datasets are taken too |
| `path has no AOD files` | Open a directory that contains AOD files |
| `minFileSizeMB is too high for this path` | Lower **Min file size [MB]** |
| `runCount is too high for this path (available runs: …)` | Request at most the number of available runs, or open a parent directory |
| `filesPerRun is too high for selected run count (eligible runs: …)` | Request fewer files per run or fewer runs |
| `"JAlien" external service is unreachable` | Open a deeper directory, at least a production. If the directory is deep enough, JAliEn is not responding; try again later |
| `"Monalisa" external service is unreachable` or `"CCDB" external service is unreachable` | The lookup at submission failed; submit again later |
