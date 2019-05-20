#!/usr/bin/env sh

cat <<EOF > ~/.netrc
machine github.com
    login $GITHUB_USER
    password $GITHUB_PWD
EOF

export GITHUB_USER=
export GITHUB_PWD=
/app/bin/goproxy -cacheDir=/go