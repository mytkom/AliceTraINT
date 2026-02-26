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

<ol>
  <li>
    Go to the <strong>Datasets</strong> section.
  </li>
  <li>
    Click <strong>Create Training Dataset</strong>. You will see a view similar to this:
    <p>
      <img src="/docs/static/0-getting-started/training-task-new.png" alt="Create dataset view">
    </p>
    <p>
      On the left there is the ALICE file browser. On the right you will see AOD files
      found under the selected directory (initially it is empty).
    </p>
  </li>
  <li>
    AOD files are usually nested in many directories, so manually browsing the full tree
    would be slow. Instead:
    <ul>
      <li>navigate to the directory that represents your data taking period or run, then</li>
      <li>click <strong>Find AO2Ds</strong> above the explorer.</li>
    </ul>
    <p>
      The application will search for AODs in the current directory and show them on the right:
    </p>
    <p>
      <img src="/docs/static/0-getting-started/training-task-new-2.png" alt="Create dataset after finding AO2Ds">
    </p>
  </li>
  <li>
    Click on an AOD file to add it to your dataset. Selected files appear in the list below.
    You can:
    <ul>
      <li>pick AODs from different directories (e.g. different LHC periods or runs),</li>
      <li>remove a file from the selection by clicking it again in the selected list.</li>
    </ul>
    <p>
      <img src="/docs/static/0-getting-started/training-task-new-3.png" alt="Create dataset with AO2Ds and title">
    </p>
  </li>
  <li>
    When you are satisfied with the selection:
    <ul>
      <li>enter a <strong>title</strong> for the dataset,</li>
      <li>make sure at least one AOD is selected,</li>
      <li>click <strong>Submit</strong>.</li>
    </ul>
    <p>
      The title must be unique in the whole application.
    </p>
  </li>
</ol>


## Queueing a training task

This part of the documentation will be updated after the new training module and its
configuration are finalized.

## Registering a new training machine

To register a new training machine:

<ol>
  <li>
    Go to the <strong>Training Machines</strong> section.
    <p>
      <img src="/docs/static/0-getting-started/register-machine-0.png" alt="Training machines listing">
    </p>
  </li>
  <li>
    Click <strong>Register Training Machine</strong>.
    <p>
      <img src="/docs/static/0-getting-started/register-machine-1.png" alt="Training machine name form">
    </p>
  </li>
  <li>
    Fill in the <strong>name</strong> of the machine and submit the form. The application will show:
    <ul>
      <li>the <strong>machine ID</strong>,</li>
      <li>the <strong>secret key</strong>.</li>
    </ul>
    <p>
      <img src="/docs/static/0-getting-started/register-machine-2.png" alt="Training machine ID and secret key">
    </p>
  </li>
  <li>
    Copy these values and configure your training module:
    <ul>
      <li>set <code>MACHINE_ID</code> and <code>MACHINE_SECRET_KEY</code> in the <code>.env</code> file or as environment variables,</li>
      <li>make sure <code>ALICETRAINT_BASE_URL</code> points to <code>https://alicetraint.app.cern.ch</code>
          in the training environment.</li>
    </ul>
  </li>
</ol>

After that, the training machine can securely request jobs from the web app and run them. 