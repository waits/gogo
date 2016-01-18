# Builds server and runs it in development mode

trap 'quit' INT

quit() {
    trap '' INT TERM
    echo "Shutting down..."
    kill -TERM 0
    rm playgo
    wait
}

go build
./playgo --reload &
open "http://localhost:8080"
wait %1
