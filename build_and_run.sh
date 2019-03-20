#!/bin/bash

id="$(docker build --quiet .)"
if [[ $id == sha256* ]] ;
then
    echo "Running docker image ${id}"
    docker run -it \
        -e "az_storage_name=${SAMPLE_STORAGE_ACCOUNT_NAME}" \
        -e "az_storage_key=${SAMPLE_STORAGE_ACCOUNT_KEY}" \
        "${id}" /bin/bash
else
    echo $id
fi
