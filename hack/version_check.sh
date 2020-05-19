#!/usr/bin/env bash

set -eo pipefail

log() { echo -e "\e[1mboldINFO [$(date +%F\ %T)] ---> $1\e[0m"; }
success() { echo -e "\e[32mSUCCESS [$(date +%F\ %T)] ---> $1\e[0m"; }
err() { echo -e "\e[31mERR [$(date +%F\ %T)] ---> $1\e[0m" ; exit 1; }

function checkInstallerVersion() {
    local last_installer_ver=$(git describe --abbrev=0 --tags master | cut -d '-' -f 1)
    local curr_installer_ver=$(head -n1 version)

    log "Last installer version is: ${last_installer_ver}"
    log "Current installer version is: ${curr_installer_ver}"

    if $(semver-cli greater ${curr_installer_ver} ${last_installer_ver}); then
        success "Installer version check is ok"
    else
        err "Installer version check failed. Please update the version file"
    fi
}

checkInstallerVersion