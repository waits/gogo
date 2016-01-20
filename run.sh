# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down...\n"
    kill -TERM 0
    exit 1
}

if ! go build; then
    exit 2
fi

./playgo --reload &
if ! ps -p $! >&-; then
    echo "Failed to start application."
    exit 3
fi

open "http://localhost:8080"
wait %1
