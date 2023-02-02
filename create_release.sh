#!/bin/bash

TOKEN="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"
VER=v1.0.0

curl \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $TOKEN"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/OrShtrauch/RouterStress/releases \
  -d '{"tag_name":"v1.0.0","target_commitish":"master","name":"v1.0.0","body":"","draft":false,"prerelease":false,"generate_release_notes":false}'