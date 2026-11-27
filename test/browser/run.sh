#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."

mode=${1:-e2e}
case "$mode" in
    e2e|baseline|serve) ;;
    *) echo "Usage: $0 [e2e|baseline|serve]" >&2; exit 2 ;;
esac
if (( $# )); then shift; fi

if command -v pg_config >/dev/null; then
    export PATH="$(pg_config --bindir):$PATH"
fi

# Use a fresh private cluster, with no TCP listener and no application .env.
root=$(mktemp -d /tmp/alicetraint-browser.XXXXXXXX)
app_pid=
cleanup() {
    if [[ -n "$app_pid" ]]; then
        kill "$app_pid" 2>/dev/null || true
        wait "$app_pid" 2>/dev/null || true
    fi
    if [[ -f "$root/pg/postmaster.pid" ]]; then
        pg_ctl -D "$root/pg" -m immediate stop
    fi
    rm -rf -- "$root"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

artifact_dir="$PWD/artifacts/browser/$mode"
mkdir -p "$artifact_dir" "$root/socket" "$root/fixtures"
initdb --pgdata="$root/pg" --username=browser_fixture --auth=trust --no-locale --encoding=UTF8 > "$artifact_dir/initdb.log" 2>&1
pg_ctl -D "$root/pg" -l "$artifact_dir/postgres.log" \
    -o "-c listen_addresses='' -c unix_socket_directories=$root/socket" -w start
env -u PGHOSTADDR -u PGSERVICE -u PGOPTIONS createdb \
    --host="$root/socket" --port=5432 --username=browser_fixture --maintenance-db=postgres alicetraint_browser_fixture

go build -o "$root/app" ./cmd/AliceTraINT_test
"$root/app" "$root" "$mode" > "$artifact_dir/app.log" 2>&1 &
app_pid=$!
for ((attempt = 0; attempt < 120; attempt++)); do
    [[ -s "$root/url" ]] && break
    kill -0 "$app_pid" 2>/dev/null || break
    sleep 0.5
done
if [[ ! -s "$root/url" ]]; then
    cat "$artifact_dir/app.log" >&2
    exit 1
fi

CYPRESS_BASE_URL=$(cat "$root/url")
export CYPRESS_BASE_URL
if [[ "$mode" == e2e ]]; then
    ./node_modules/.bin/cypress run "$@"
elif [[ "$mode" == baseline ]]; then
    node test/browser/baseline.mjs "$root/chrome" | tee "$artifact_dir/results.json"
else
    printf "Fixture app listening on %s\n" "$CYPRESS_BASE_URL"
    wait "$app_pid"
fi
