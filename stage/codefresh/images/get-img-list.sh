#!/bin/bash

set -eo pipefail

while [[ $1 =~ ^(--(repo|version|show-digests|gcr-sa)) ]]
    do
        key=$1
        value=$2
        case $key in
            --repo)
                REPO_CHANNEL="$value"
                shift
            ;;
            --version)
                ONPREM_VERSION="--version $value"
                shift
            ;;
            --show-digests)
                SHOW_DIGESTS="true"
            ;;
            --gcr-sa)
                GCR_SA_FILE="$value"
            ;;
        esac
        shift
    done

REPO_CHANNEL=${REPO_CHANNEL:-"prod"}
CHART=codefresh-onprem-${REPO_CHANNEL}/codefresh
ONPREM_VERSION=${ONPREM_VERSION:-""}
SHOW_DIGESTS=${SHOW_DIGESTS:-"false"}
SKOPEO_IMAGE="quay.io/codefresh/skopeo"
SKOPEO_CONTAINER="cf-skopeo"

HELM_VALS="--set global.seedJobs=true --set global.certsJobs=true"

set -u 

function getHelmReleaseImages() {
    helm template ${LOCAL_CHART_PATH}/* ${HELM_VALS} | grep 'image:' | awk -F 'image: ' '{print $2}' | tr -d '"' | sort -u
}

function getRuntimeImages() {

    RUNTIME_IMAGES=(
        CONTAINER_LOGGER_IMAGE
        DOCKER_PUSHER_IMAGE
        DOCKER_TAG_PUSHER_IMAGE
        DOCKER_PULLER_IMAGE
        DOCKER_BUILDER_IMAGE
        GIT_CLONE_IMAGE
        COMPOSE_IMAGE
        KUBE_DEPLOY
        FS_OPS_IMAGE
        TEMPLATE_ENGINE
        PIPELINE_DEBUGGER_IMAGE
    )

    for k in ${RUNTIME_IMAGES[@]}; do
        helm template ${LOCAL_CHART_PATH}/* ${HELM_VALS} | grep "$k" | tr -d '"' | tr -d ',' | awk -F "$k: " '{print $2}' | sort -u
    done
    
    helm template ${LOCAL_CHART_PATH}/* ${HELM_VALS} | grep 'engine:' | tr -d '"' | tr -d ',' | awk -F 'image: ' '{print $2}'| sort -u # engine image
    helm template ${LOCAL_CHART_PATH}/* ${HELM_VALS} | grep '"dindImage"'  | tr -d '"' | tr -d ',' | awk -F ' ' '{print $2}' | sort -u # dind image
}

function getOtherImages() {
    
    OTHER_IMAGES=(
        quay.io/codefresh/cf-runtime-cleaner:1.2.0
        quay.io/codefresh/agent:stable
        gcr.io/codefresh-enterprise/codefresh/cf-k8s-monitor:4.6.3
        quay.io/codefresh/kube-helm:3.0.3
        quay.io/codefresh/hermes-store-backup:0.2.0
    )

    for i in ${OTHER_IMAGES[@]}; do
        echo $i
    done
}

function getImages() {
    getHelmReleaseImages
    getRuntimeImages
    getOtherImages
}

function getDigest() {
   local manifest
   local digest
   
   digest=$(docker exec $SKOPEO_CONTAINER skopeo inspect docker://$1 --format {{.Digest}} 2>&1)
   if [[ "$?" == "1" ]]; then
        echo "Error: $digest"
        return
   fi

   echo $digest
}

function printImage() {
    if [[ "$SHOW_DIGESTS" == "true" ]]; then
        local digest=$(getDigest $1)
        local space_width=$(( 80 - "$(echo $1 | wc -c)"  ))

        local spacing=$(awk "BEGIN{for(c=0;c<${space_width};c++) printf \" \"}")
        echo "$1${spacing}$digest"
    else
        echo "$1"
    fi
}

function initSkopeo() {

   docker run --rm -d \
         --name ${SKOPEO_CONTAINER} \
         --entrypoint sh \
         -w /skopeo \
         ${SKOPEO_IMAGE} \
         -c 'sleep 1000'

   local gcr_pass=$(cat ${GCR_SA_FILE})
   docker exec $SKOPEO_CONTAINER skopeo login -u _json_key -p "$gcr_pass" gcr.io
}

function stopSkopeo() {
   docker stop ${SKOPEO_CONTAINER} &> /dev/null
}

function printImages() {
    if [[ "$SHOW_DIGESTS" == "true" ]]; then
        trap stopSkopeo EXIT

        initSkopeo 1> /dev/null
    fi

    set +e
    local tmpfile=$(mktemp)
    for i in $IMAGES; do
        printImage $i >> ${tmpfile} &
    done

    wait

    cat ${tmpfile} | sort
}

LOCAL_CHART_PATH=$(mktemp -d)
helm repo add codefresh-onprem-${REPO_CHANNEL} http://charts.codefresh.io/${REPO_CHANNEL} &>/dev/null
helm fetch ${CHART} ${ONPREM_VERSION} -d ${LOCAL_CHART_PATH}

IMAGES=$(getImages | sort -u)

printImages