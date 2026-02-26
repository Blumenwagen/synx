#!/bin/bash

set -e

# Calculate dynamic version based on Git commit history
# 1. Count commits that touched non-markdown files (main logic changes)
# 2. Get the current short hash for uniqueness

# Move to the git repo root to execute git commands safely
cd "$(dirname "$0")/.."

# Safely check if we are in a git repository
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    COM_COUNT=$(git rev-list HEAD -- . ":(exclude)*.md" | wc -l | tr -d ' ')
    COM_HASH=$(git rev-parse --short HEAD)
    VERSION="v${COM_COUNT}-${COM_HASH}"
else
    # Fallback to dev version if not downloaded via Git
    VERSION="dev"
fi

# Move back to synx-go for building
cd synx-go

echo "Building synx version $VERSION..."
go build -ldflags="-X 'github.com/Blumenwagen/synx/cmd.Version=$VERSION'" -o synx .
echo "Build complete."
