# Cara Mengamankan Aplikasi Golang dalam Kontainer dengan Nginx, Let's Encrypt, dan Docker Compose

### Initial Server Setup with Ubuntu

- sudo apt-get update && sudo apt-get upgrade
- sudo adduser sammidev
- sudo usermod -aG sudo sammidev

Setting Up a Basic Firewall
- sudo ufw app list
- sudo ufw allow OpenSSH
- sudo ufw enable `Type “y” and press ENTER to proceed.`
- sudo ufw status

If the Root Account Uses SSH Key Authentication
- sudo rsync --archive --chown=sammidev:sammidev ~/.ssh /home/sammidev
- ssh sammidev@server_ip

Install docker & docker compose
- sudo apt-get install curl apt-transport-https ca-certificates software-properties-common
- curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
- sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
- sudo apt update
- sudo apt install docker-ce
- sudo systemctl status docker
- docker --version
- sudo apt-get update
- sudo apt-get install docker-compose-plugin
- apt-cache madison docker-compose-plugin
- sudo apt-get install docker-compose-plugin=`version`
- docker compose version

Mengeksekusi Perintah Docker Tanpa Sudo
- sudo usermod -aG docker ${USER}
- su - ${USER}
- id -nG

Testing the Golang App

- git clone https://github.com/SemmiDev/gogreet.git
- cd gogreet
- cat Dockerfile
- docker build -t go-demo .
- docker images
- docker run --name go-demo -p 80:8080 -d go-demo
- docker ps
- docker stop `CONTAINER ID`
- docker system prune -a
  
- mkdir nginx-conf
- nano nginx-conf/nginx.conf
```nginx
server {
        listen 80;
        listen [::]:80;

        root /var/www/html;
        index index.html index.htm index.nginx-debian.html;

        server_name sammidev.codes www.sammidev.codes;

        location / {
                proxy_pass http://goapp:8080;
        }

        location ~ /.well-known/acme-challenge {
                allow all;
                root /var/www/html;
        }
}
```

- nano docker-compose.yml
```docker compose
version: '3'

services:
  goapp:
    build:
      context: .
      dockerfile: Dockerfile
    image: golang
    container_name: goapp
    restart: unless-stopped
    networks:
      - app-network

  webserver:
    image: nginx:mainline-alpine
    container_name: webserver
    restart: unless-stopped
    ports:
      - "80:80"
    volumes:
      - web-root:/var/www/html
      - ./nginx-conf:/etc/nginx/conf.d
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
    depends_on:
      - goapp
    networks:
      - app-network

  certbot:
    image: certbot/certbot
    container_name: certbot
    volumes:
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
      - web-root:/var/www/html
    depends_on:
      - webserver
    command: certonly --webroot --webroot-path=/var/www/html --email sammidev4@gmail.com --agree-tos --no-eff-email --staging -d sammidev.codes  -d www.sammidev.codes 

volumes:
  certbot-etc:
  certbot-var:
  web-root:
    driver: local

networks:
  app-network:
    driver: bridge
```

- docker compose up -d
- docker compose ps
- docker compose logs `service_name`
- docker compose exec webserver ls -la /etc/letsencrypt/live
- nano docker-compose.yml
```
version: '3'

services:
  goapp:
    build:
      context: .
      dockerfile: Dockerfile
    image: golang
    container_name: goapp
    restart: unless-stopped
    networks:
      - app-network

  webserver:
    image: nginx:mainline-alpine
    container_name: webserver
    restart: unless-stopped
    ports:
      - "80:80"
    volumes:
      - web-root:/var/www/html
      - ./nginx-conf:/etc/nginx/conf.d
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
    depends_on:
      - goapp
    networks:
      - app-network

  certbot:
    image: certbot/certbot
    container_name: certbot
    volumes:
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
      - web-root:/var/www/html
    depends_on:
      - webserver
    command: certonly --webroot --webroot-path=/var/www/html --email sammidev4@gmail.com --agree-tos --no-eff-email --force-renewal -d sammidev.codes  -d www.sammidev.codes 

volumes:
  certbot-etc:
  certbot-var:
  web-root:
    driver: local

networks:
  app-network:
    driver: bridge
```
- docker compose up --force-recreate --no-deps certbot
- docker compose stop webserver
- mkdir dhparam
- sudo openssl dhparam -out /home/sammidev/gogreet/dhparam/dhparam-2048.pem 2048
- rm nginx-conf/nginx.conf
- nano nginx-conf/nginx.conf
```nginx
server {
        listen 80;
        listen [::]:80;
        server_name sammidev.codes www.sammidev.codes;

        location ~ /.well-known/acme-challenge {
          allow all;
          root /var/www/html;
        }

        location / {
                rewrite ^ https://$host$request_uri? permanent;
        }
}

server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name sammidev.codes www.sammidev.codes;

        server_tokens off;

        ssl_certificate /etc/letsencrypt/live/sammidev.codes/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/sammidev.codes/privkey.pem;

        ssl_buffer_size 8k;

        ssl_dhparam /etc/ssl/certs/dhparam-2048.pem;

        ssl_protocols TLSv1.2 TLSv1.1 TLSv1;
        ssl_prefer_server_ciphers on;

        ssl_ciphers ECDH+AESGCM:ECDH+AES256:ECDH+AES128:DH+3DES:!ADH:!AECDH:!MD5;

        ssl_ecdh_curve secp384r1;
        ssl_session_tickets off;

        ssl_stapling on;
        ssl_stapling_verify on;
        resolver 8.8.8.8;

        location / {
                try_files $uri @goapp;
        }

        location @goapp {
                proxy_pass http://goapp:8080;
                add_header X-Frame-Options "SAMEORIGIN" always;
                add_header X-XSS-Protection "1; mode=block" always;
                add_header X-Content-Type-Options "nosniff" always;
                add_header Referrer-Policy "no-referrer-when-downgrade" always;
                add_header Content-Security-Policy "default-src * data: 'unsafe-eval' 'unsafe-inline'" always;
                # add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
                # enable strict transport security only if you understand the implications
        }

        root /var/www/html;
        index index.html index.htm index.nginx-debian.html;
}
```
- nano docker-compose.yml
```docker compose
version: '3'

services:
  goapp:
    build:
      context: .
      dockerfile: Dockerfile
    image: golang
    container_name: goapp
    restart: unless-stopped
    networks:
      - app-network

  webserver:
    image: nginx:mainline-alpine
    container_name: webserver
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - web-root:/var/www/html
      - ./nginx-conf:/etc/nginx/conf.d
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
      - dhparam:/etc/ssl/certs
    depends_on:
      - goapp
    networks:
      - app-network

  certbot:
    image: certbot/certbot
    container_name: certbot
    volumes:
      - certbot-etc:/etc/letsencrypt
      - certbot-var:/var/lib/letsencrypt
      - web-root:/var/www/html
    depends_on:
      - webserver
    command: certonly --webroot --webroot-path=/var/www/html --email sammidev4@gmail.com --agree-tos --no-eff-email --force-renewal -d sammidev.codes  -d www.sammidev.codes 

volumes:
  certbot-etc:
  certbot-var:
  web-root:
    driver: local
  dhparam:
    driver: local
    driver_opts:
      type: none
      device: /home/sammidev/gogreet/dhparam/
      o: bind

networks:
  app-network:
    driver: bridge
```
- docker compose up -d --force-recreate --no-deps webserver
- docker compose ps
- try to access!!

- nano ssl_renew.sh
```bash
#!/bin/bash

COMPOSE="/usr/local/bin/docker compose --no-ansi"
DOCKER="/usr/bin/docker"

cd /home/sammy/node_project/
$COMPOSE run certbot renew && $COMPOSE kill -s SIGHUP webserver
$DOCKER system prune -af
```
- chmod +x ssl_renew.sh
- sudo crontab -e 
- tail -f /var/log/cron.log

To run the script every day at noon
```crob
0 12 * * * /home/sammidev/go-greet/ssl_renew.sh >> /var/log/cron.log 2>&1
```
