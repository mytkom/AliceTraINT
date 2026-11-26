# Getting Started

## Who can use AliceTraINT

To use this application you need a CERN account and you must log in through CERN SSO.  
Your account also has to be a member of the `wut-alice-pidml` group.

## What AliceTraINT is for

AliceTraINT is a simple web application that helps the ALICE WUT team manage training runs
for PID (Particle Identification) machine‑learning models.

In one place you can:

- create training datasets from AOD files chosen in JAliEn,
- create and queue training tasks that will run on external machines (clusters, servers),
- keep logs, evaluation metrics and plots together with the exact dataset and config,
- browse the full history of training runs and share results with others.

## How it works (high level)

1. You log in to the web app from any device (desktop or mobile).
2. You create a **dataset** and then a **training task** in the UI.
3. A separate **training machine** runs the training module and regularly asks the web app:
   “Is there a job for me?”
4. When a job is available, the training machine:
   - downloads all dataset files,
   - prepares the dataset using the O2Physics task,
   - trains the model,
   - evaluates it and uploads logs, plots and metrics back to the web app.
5. The web app shows the status of the task at every stage and lets you review results.
6. When you are happy with the model, you can upload it to the production CCDB for use
   in O2Physics tasks with a single click.

The code is open source:

- [AliceTraINT web app](https://github.com/mytkom/AliceTraINT)
- [AliceTraINT training module](https://github.com/mytkom/AliceTraINT_pidml_training_module)

## Creating a dataset

1. Go to the **Datasets** section.
2. Click **Create Training Dataset**. You will see a view similar to this:

   ![Create dataset view](/docs/static/1-user-guide/1-getting-started/training-task-new.png)

   On the left there is the ALICE file browser. On the right you will see AOD files
   found under the selected directory (initially it is empty).

3. AOD files are usually nested in many directories, so manually browsing the full tree
   would be slow. Instead:
   - navigate to the directory that represents your data taking period or run, then
   - click **Find AO2Ds** above the explorer.

   The application will search for AODs in the current directory and show them on the right:

   ![Create dataset after finding AO2Ds](/docs/static/1-user-guide/1-getting-started/training-task-new-2.png)

4. Click on an AOD file to add it to your dataset. Selected files appear in the list below.
   You can:
   - pick AODs from different directories (e.g. different LHC periods or runs),
   - remove a file from the selection by clicking it again in the selected list.

   ![Create dataset with AO2Ds and title](/docs/static/1-user-guide/1-getting-started/training-task-new-3.png)

5. When you are satisfied with the selection:
   - enter a **title** for the dataset,
   - make sure at least one AOD is selected,
   - click **Submit**.

   The title must be unique in the whole application.


## Queueing a training task

This part of the documentation will be updated after the new training module and its
configuration are finalized.

## Registering a new training machine

To register a new training machine:

1. Go to the **Training Machines** section.

   ![Training machines listing](/docs/static/1-user-guide/1-getting-started/register-machine-0.png)

2. Click **Register Training Machine**.

   ![Training machine name form](/docs/static/1-user-guide/1-getting-started/register-machine-1.png)

3. Fill in the **name** of the machine and submit the form. The application will show:
   - the **machine ID**,
   - the **secret key**.

   ![Training machine ID and secret key](/docs/static/1-user-guide/1-getting-started/register-machine-2.png)

4. Copy these values and configure your training module:
   - set `MACHINE_ID` and `MACHINE_SECRET_KEY` in the `.env` file or as environment variables,
   - make sure `ALICETRAINT_BASE_URL` points to `https://alicetraint.app.cern.ch`
     in the training environment.

After that, the training machine can securely request jobs from the web app and run them. 