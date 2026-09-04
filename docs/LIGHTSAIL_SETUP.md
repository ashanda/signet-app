# AWS Lightsail server setup (Ubuntu)

Step-by-step guide to take a fresh Ubuntu 22.04/24.04 Lightsail instance to
a fully working deployment of this app, plus MySQL and phpMyAdmin for
database administration. Run every command block on the server unless
noted otherwise.

Recommended instance size: **2 GB RAM minimum** (the $10/mo Lightsail
plan). Building the Go backend and the Vite frontend on a 512 MB–1 GB
instance will OOM-kill the build — either size up, or add a swap file
(step 1.3), or build on your own machine and `scp` the artifacts up
instead (see the note in step 6/7).

---

## 1. Initial server access and hardening

### 1.1 Open the ports you need (Lightsail console, not the OS)

In the Lightsail console → your instance → **Networking** tab, add firewall
rules for:

| Application | Protocol | Port |
|---|---|---|
| SSH | TCP | 22 (already open by default) |
| HTTP | TCP | 80 |
| HTTPS | TCP | 443 |

Leave MySQL's port **3306 closed** to the internet — the app and
phpMyAdmin both reach MySQL over `localhost`, so it never needs to be
publicly reachable. If you truly need remote MySQL access, restrict it to
your own IP in the Lightsail firewall rather than opening it to
`0.0.0.0/0`.

Also attach a **static IP** to the instance (Lightsail console → Networking
→ "Create static IP") so it survives reboots — otherwise every restart
changes your public IP and breaks DNS.

### 1.2 Connect

Either use the Lightsail console's browser-based SSH ("Connect using SSH"
button on the instance page), or from your own machine with the downloaded
key:

```bash
chmod 400 LightsailDefaultKey.pem
ssh -i LightsailDefaultKey.pem ubuntu@YOUR_STATIC_IP
```

### 1.3 Update the system, add swap (small instances only)

```bash
sudo apt update && sudo apt upgrade -y
```

Skip this if you sized up to 2 GB+ RAM. Otherwise, add a 2 GB swap file so
`go build`/`npm run build` don't get OOM-killed:

```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### 1.4 Basic firewall (ufw) matching the Lightsail rules above

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable
sudo ufw status
```

---

## 2. Install MySQL

```bash
sudo apt install -y mysql-server
sudo mysql_secure_installation
```

Answer the prompts: set a strong root password, remove anonymous users,
disallow remote root login, remove the test database, reload privileges —
yes to all.

Create the app's database and a dedicated (non-root) user for it:

```bash
sudo mysql -u root -p
```

```sql
CREATE DATABASE signet_last CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'signet'@'localhost' IDENTIFIED BY 'CHANGE_ME_TO_A_STRONG_PASSWORD';
GRANT ALL PRIVILEGES ON signet_last.* TO 'signet'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

Use the `signet` user (not `root`) in the backend's `.env` — see step 7.

---

## 3. Install nginx and PHP-FPM

nginx serves the frontend's static files, reverse-proxies `/api` and
`/storage` to the Go backend, **and** serves phpMyAdmin via PHP-FPM — no
Apache needed.

```bash
sudo apt install -y nginx php-fpm php-mysqli php-mbstring php-zip php-gd php-json php-curl
```

Note the installed PHP-FPM socket path (varies by PHP version bundled with
your Ubuntu release):

```bash
ls /run/php/
# e.g. php8.1-fpm.sock or php8.3-fpm.sock — use this in the nginx configs below
```

---

## 4. Install phpMyAdmin (manual install, no Apache)

Ubuntu's `phpmyadmin` apt package tries to auto-configure Apache; installing
it manually avoids pulling in a second web server.

```bash
cd /usr/share
sudo curl -LO https://www.phpmyadmin.net/downloads/phpMyAdmin-latest-all-languages.tar.gz
sudo tar -xzf phpMyAdmin-latest-all-languages.tar.gz
sudo mv phpMyAdmin-*-all-languages phpmyadmin
sudo rm phpMyAdmin-latest-all-languages.tar.gz

sudo mkdir -p /var/lib/phpmyadmin/tmp
sudo chown -R www-data:www-data /var/lib/phpmyadmin
sudo cp /usr/share/phpmyadmin/config.sample.inc.php /usr/share/phpmyadmin/config.inc.php
```

Generate a secret and edit the config:

```bash
openssl rand -hex 32
sudo nano /usr/share/phpmyadmin/config.inc.php
```

Set:
```php
$cfg['blowfish_secret'] = 'PASTE_THE_GENERATED_HEX_HERE';
$cfg['TempDir'] = '/var/lib/phpmyadmin/tmp';
```

### nginx server block for phpMyAdmin (subdomain recommended)

Point a DNS `A` record — e.g. `pma.your-domain.com` — at the instance's
static IP, then:

```nginx
# /etc/nginx/sites-available/phpmyadmin
server {
    listen 80;
    server_name pma.your-domain.com;
    root /usr/share/phpmyadmin;
    index index.php;

    location / {
        try_files $uri $uri/ =404;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;  # match the socket from step 3
    }

    location ~ /\.ht {
        deny all;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/phpmyadmin /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

**Lock this down further before going live** — phpMyAdmin is a common
attack target. At minimum, add HTTP basic auth in front of it:

```bash
sudo apt install -y apache2-utils
sudo htpasswd -c /etc/nginx/.htpasswd youradminuser
```

Add inside the `server { }` block above, before the `location /` block:
```nginx
auth_basic "Restricted";
auth_basic_user_file /etc/nginx/.htpasswd;
```

Even better: restrict the phpMyAdmin server block to your own IP via an
`allow`/`deny` pair, or only reach it through an SSH tunnel
(`ssh -L 8080:localhost:80 ubuntu@YOUR_IP`, then visit
`http://localhost:8080` locally) and never open it to the internet at all.

---

## 5. Install Go and Node.js (only needed if building on the server)

```bash
# Go
curl -LO https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version

# Node.js 20.x (NodeSource)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
node --version
```

If you'd rather not build on a small instance, build both `backend/signet-api`
(with `GOOS=linux GOARCH=amd64 go build`) and `frontend/dist/` on your own
machine and `scp` them to the server instead, skipping this step entirely.

---

## 6. Get the code onto the server

The GitHub repo is private, so cloning needs a deploy key or a token.
Simplest for a single server: generate a dedicated SSH key on the server
and add it as a **deploy key** (read-only) on the GitHub repo.

```bash
ssh-keygen -t ed25519 -C "lightsail-deploy" -f ~/.ssh/id_ed25519 -N ""
cat ~/.ssh/id_ed25519.pub
```

Copy that public key into GitHub → your repo → **Settings → Deploy keys →
Add deploy key** (no write access needed).

```bash
sudo mkdir -p /opt/signet
sudo chown ubuntu:ubuntu /opt/signet
git clone git@github.com:ashanda/signet-app.git /opt/signet
cd /opt/signet
```

---

## 7. Deploy the backend

```bash
cd /opt/signet/backend
cp .env.example .env.production
nano .env.production
```

Fill in at minimum:
```
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=signet_last
DB_USERNAME=signet
DB_PASSWORD=the-password-you-set-in-step-2
SESSION_SECRET=$(openssl rand -hex 32)
FRONTEND_ORIGIN=https://your-domain.com
```

```bash
chmod 600 .env.production
set -a; source .env.production; set +a
go run ./cmd/migrate            # bootstraps the schema (safe to re-run)
go build -o signet-api ./cmd/api
```

Install as a systemd service — see **DEPLOYMENT.md § 3** for the full
`signet-api.service` unit file and `systemctl enable --now` steps (same
process, just pointed at `/opt/signet/backend` and `.env.production`).

---

## 8. Deploy the frontend

```bash
cd /opt/signet/frontend
npm install
npm run build
```

`frontend/dist/` is now ready to serve.

### nginx server block for the app

Point your main domain (e.g. `your-domain.com`) at the instance, then use
the nginx config from **DEPLOYMENT.md § 5**, with `root
/opt/signet/frontend/dist;`.

```bash
sudo ln -s /etc/nginx/sites-available/signet /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

---

## 9. TLS (both domains)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com -d pma.your-domain.com
```

Certbot edits both nginx server blocks in place to add the `listen 443
ssl` directives and sets up auto-renewal (`systemctl status certbot.timer`
to confirm it's active).

---

## 10. Post-deploy checklist

- [ ] `sudo systemctl status signet-api` — backend running
- [ ] `curl -I https://your-domain.com` — frontend serving
- [ ] `curl -I https://your-domain.com/api/v1/login` — API reachable through the proxy
- [ ] phpMyAdmin reachable only behind basic auth / IP restriction / SSH tunnel — **not open to the world**
- [ ] MySQL `signet` user (not `root`) is what the backend `.env` uses
- [ ] `SESSION_SECRET` is a real random value, not the placeholder
- [ ] Static IP attached so a reboot doesn't change your DNS target
- [ ] Set up a MySQL backup (Lightsail's own automatic snapshots cover the
      whole instance disk, but a `mysqldump` cron job to offsite storage is
      cheap insurance — not covered here, ask if you want that added)
