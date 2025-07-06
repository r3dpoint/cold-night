#!/usr/bin/env bash

set -euxo pipefail

go version
# Install Air for live reloading
go install github.com/air-verse/air@latest

# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install gotest.tools/gotestsum@latest

npm install -g @anthropic-ai/claude-code

# nvm alias default lts/iron
# npm install
# go install github.com/air-verse/air@latest

# git config --global branch.autosetupmerge always
# git config --global branch.autosetuprebase always

# sudo apt-get update && apt-get install google-cloud-sdk-gke-gcloud-auth-plugin
