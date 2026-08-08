#!/usr/bin/env bash
set -e
REPO_URL=${1:-""}
if [ -z "$REPO_URL" ]; then
  echo "Usage: ./scripts/init_github.sh <repo-url>"
  exit 1
fi
git init
git add .
git commit -m "feat: initial Go v2.2.1 SRE hardening"
git branch -M main
git remote add origin $REPO_URL
git push -u origin main
echo "Push OK. Go to GitHub -> Code -> Codespaces -> Create codespace"
