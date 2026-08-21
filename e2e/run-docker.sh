#!/bin/sh
set -eu
main_was_running=false

cleanup() {
  npm run compose:down
  if [ "$main_was_running" = true ]; then
    docker compose -f ../docker-compose.yml start
  fi
}
trap cleanup EXIT

if [ -n "$(docker compose -f ../docker-compose.yml ps -q)" ]; then
  main_was_running=true
  docker compose -f ../docker-compose.yml stop
fi

npm run compose:down
npm run compose:up
if ! npx playwright test; then
  docker compose -p scaean-gate-e2e -f ../docker-compose.yml logs --no-color --tail=150 auth-provider sync-worker app-a app-b
  exit 1
fi
