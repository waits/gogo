# Uploads the application to the server

sftp -i id_ecdsa dylan@go.waits.io <<'EOF'
cd /usr/local/bin
rename gogo gogo.old
put gogo
EOF

ssh -i id_ecdsa dylan@go.waits.io <<'EOF'
cd /srv
git pull origin master
sudo service gogo restart
rm /usr/local/bin/gogo.old
EOF
