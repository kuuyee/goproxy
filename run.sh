#!/usr/bin/env sh

cat <<EOF > ~/.netrc
machine github.com
    login a3V1eWVlCg==
    password bGludXhnejExMDQxMQo=
EOF

export GITHUB_USER=
export GITHUB_PWD=
/app/bin/goproxy -cacheDir=/go