#!/usr/bin/env bash

# Copyright maksim.paskal@gmail.com
#
# Licensed under the Apache License, Version 2.0 (the "License")
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -euo pipefail

export CGO_ENABLED=0
export GO111MODULE=on
export TAGS=""
export GOFLAGS=""
export LDFLAGS="-X main.buildTime=$(date +\"%Y%m%d%H%M%S\")"
export TARGETS="darwin/amd64 linux/amd64"
export BINNAME="telegram-gateway"
export GOX="go run github.com/mitchellh/gox"

rm -rf _dist

$GOX -parallel=3 -output="_dist/{{.OS}}-{{.Arch}}/$BINNAME" -osarch="$TARGETS" $GOFLAGS -tags "$TAGS" -ldflags "$LDFLAGS"

for dir in _dist/*
do
  base=$(basename "$dir")
  tar -czf "_dist/$BINNAME-${base}.tar.gz" "$dir"
done