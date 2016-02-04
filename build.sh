# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down...\n"
    kill -TERM 0
    rm gogo
    exit 1
}

while getopts "op" opt; do
    case $opt in
        o) open=true;;
        p) production=true;;
    esac
done

if [ $production ]; then
    env GOOS=linux GOARCH=amd64 go build -o gogo
    sftp go <<'EOF'
cd /srv
put gogo gogo
EOF
    rm gogo
else
    if ! go build -o gogo; then
        exit 2
    fi
    ./gogo --reload &
    if ! ps -p $! >&-; then
        echo "Failed to start application."
        exit 3
    fi

    if [ $open ]; then
        open "http://localhost:8080"
    fi
    wait %1
fi
