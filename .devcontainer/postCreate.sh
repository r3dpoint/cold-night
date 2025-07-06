#!/usr/bin/env bash

set -euxo pipefail

go version
npm install -g @anthropic-ai/claude-code

# set a well known password for jetbrains ssh
# new_passwd="password"
# echo -e "${new_passwd}\n${new_passwd}" | sudo passwd "$(whoami)"

# nvm alias default lts/iron
# npm install
# go install github.com/air-verse/air@latest

# git config --global branch.autosetupmerge always
# git config --global branch.autosetuprebase always

# sudo apt-get update && apt-get install google-cloud-sdk-gke-gcloud-auth-plugin
