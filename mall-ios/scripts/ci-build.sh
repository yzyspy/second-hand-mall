#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

SIMULATOR_NAME="${MALL_IOS_SIMULATOR:-iPhone 17}"

xcodebuild -project MallApp.xcodeproj -scheme MallApp \
  -destination "platform=iOS Simulator,name=${SIMULATOR_NAME}" \
  build
