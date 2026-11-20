#!/usr/bin/env bash
# builds the web ui, installs aoc-runner with go install and starts the server.
# any arguments are passed through to `aoc-runner http`, e.g. ./start.sh --config ./config.yaml
set -euo pipefail

cd "$(dirname "$0")"

for tool in go npm; do
	if ! command -v "$tool" >/dev/null; then
		echo "missing $tool, install it and try again" >&2
		exit 1
	fi
done

# only reinstall node modules when the lockfile changed
if [ ! -d ui/node_modules ] || [ ui/package-lock.json -nt ui/node_modules ]; then
	echo "installing ui dependencies"
	(cd ui && npm ci)
	touch ui/node_modules
fi

echo "building ui"
(cd ui && npm run build --silent)

echo "installing aoc-runner"
go install .

bin="$(go env GOBIN)"
bin="${bin:-$(go env GOPATH)/bin}"

echo "installed to $bin/aoc-runner"

exec "$bin/aoc-runner" http "$@"
