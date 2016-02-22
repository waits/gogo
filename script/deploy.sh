# Builds the application, uploads it to the server, and restarts the daemon

sftp -i id_ecdsa -o StrictHostKeyChecking=no dylan@go.waits.io <<'EOF'
cd /srv
rename gogo gogo.old
put gogo
EOF

ssh -i id_ecdsa -o StrictHostKeyChecking=no dylan@go.waits.io <<'EOF'
cd /srv
git pull origin master
sudo service gogo restart
rm /srv/gogo.old
EOF
