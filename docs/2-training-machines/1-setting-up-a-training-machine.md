# Setting up a training machine

To register a new training machine:

1. Go to the **Training Machines** section.

   ![Training machines listing](/docs/static/2-training-machines/1-setting-up-a-training-machine/register-machine-0.png)

2. Click **Register Training Machine**.

   ![Training machine name form](/docs/static/2-training-machines/1-setting-up-a-training-machine/register-machine-1.png)

3. Fill in the **name** of the machine and submit the form. The application will show:
   - the **machine ID**,
   - the **secret key**.

   ![Training machine ID and secret key](/docs/static/2-training-machines/1-setting-up-a-training-machine/register-machine-2.png)

4. Copy these values and configure your training module:
   - set `MACHINE_ID` and `MACHINE_SECRET_KEY` in the `.env` file or as environment variables,
   - make sure `ALICETRAINT_BASE_URL` points to `https://alicetraint.app.cern.ch`
     in the training environment.

After that, the training machine can securely request jobs from the web app and run them. 