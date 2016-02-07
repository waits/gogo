# Uploads the application to the server

GOOS=linux GOARCH=amd64 go build -o gogo

sftp -i id_ecdsa -o StrictHostKeyChecking=no dylan@go.waits.io <<'EOF'
cd /usr/local/bin
rename gogo gogo.old
put gogo
EOF

ssh -i id_ecdsa -o StrictHostKeyChecking=no dylan@go.waits.io <<'EOF'
cd /srv
git pull origin master
sudo service gogo restart
rm /usr/local/bin/gogo.old
EOF
