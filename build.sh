# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down...\n"
    kill -TERM 0
    rm go_dev
    exit 1
}

while getopts "do" opt; do
    case $opt in
        d) deploy=true;;
        o) open=true;;
    esac
done

if [ $deploy ]; then
    env GOOS=linux GOARCH=amd64 go build -o go_server
else
    if ! go build -o go_dev; then
        exit 2
    fi
    ./go_dev --reload &
    if ! ps -p $! >&-; then
        echo "Failed to start application."
        exit 3
    fi

    if [ $open ]; then
        open "http://localhost:8080"
    fi
    wait %1
fi
