#!/usr/bin/env sh

cat <<EOF > ~/.netrc
machine github.com
    login $GITHUB_USER
    password $GITHUB_PWD
EOF

/app/bin/goproxy -cacheDir=/go