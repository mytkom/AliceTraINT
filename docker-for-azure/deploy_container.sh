#!/usr/bin/env bash

# ENVs to fill
# RESOURCE_GROUP=
# STORAGE_ACCOUNT_NAME=
# LOCATION=
# SHARE_NAME=
# ALICETRAINT_ACR_NAME=
# ALICETRAINT_ACR_URL=
# IMAGE_NAME=
# STORAGE_KEY=$(az storage account keys list --resource-group $RESOURCE_GROUP --account-name $STORAGE_ACCOUNT_NAME --query "[0].value" --output tsv)
# SERVICE_PRINCIPAL_USERNAME=
# SERVICE_PRINCIPAL_PASSWORD=
# PGHOST=
# PGUSER=
# PGPORT=
# PGDATABASE=
# PGPASSWORD=
# CERN_REDIRECT_URL=
# SSL_ROOT_CERT_PATH=

# Before this script
# 1. Log In to acr
# 2. Set all envs
# 3. Run build.sh

# Deploy container
az container create --resource-group $RESOURCE_GROUP \
  --name alicetraint \
  --image $IMAGE_NAME \
  --dns-name-label alicetraint \
  --ports 80 \
  --cpu 1 \
  --memory 1 \
  --registry-username $SERVICE_PRINCIPAL_USERNAME \
  --registry-password $SERVICE_PRINCIPAL_PASSWORD \
  --azure-file-volume-account-name $STORAGE_ACCOUNT_NAME \
  --azure-file-volume-account-key $STORAGE_KEY \
  --azure-file-volume-share-name $SHARE_NAME \
  --azure-file-volume-mount-path /app/data/ \
  --secure-environment-variables \
    'DB_PASSWORD'=$PGPASSWORD \
  --environment-variables \
    'DB_HOST'=$PGHOST \
    'DB_PORT'=$PGPORT \
    'DB_USER'=$PGUSER \
    'DB_NAME'=$PGDATABASE \
    'CERN_REDIRECT_URL'=$CERN_REDIRECT_URL \
    'ALICETRAINT_PORT'='80' \
    'DB_SSL_CERT_PATH'=$SSL_ROOT_CERT_PATH \
    'DB_SSLMODE'='verify-full'
