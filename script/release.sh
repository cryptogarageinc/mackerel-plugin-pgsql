#!/bin/sh

# Usage:
# $ script/release # Setting github.token in .gitconfig is required
# $ GITHUB_TOKEN=... script/release

set -e
latest_tag=$(git describe --abbrev=0 --tags)
goxz -d dist/$latest_tag -z -os windows,darwin,linux -arch amd64,386
ghr -replace -u cryptogarageinc -r mackerel-plugin-pgsql $latest_tag dist/$latest_tag
