# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down...\n"
    kill -TERM 0
    exit 1
}

if ! go build -o go_server; then
    exit 2
fi

./go_server --reload &
if ! ps -p $! >&-; then
    echo "Failed to start application."
    exit 3
fi

if getopts "o" novar; then
    open "http://localhost:8080"
fi
wait %1
