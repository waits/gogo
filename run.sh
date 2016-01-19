# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down..."
    kill -TERM 0
    rm playgo
    wait
}

if go build; then
    ./playgo --reload &
    open "http://localhost:8080"
    wait %1
fi
