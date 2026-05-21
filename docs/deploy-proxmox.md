# Deploy Zolix di Proxmox

Arsitektur target:

- `192.168.18.44`: Proxmox Host. Jangan install aplikasi langsung di host ini.
- `192.168.18.100`: VM App Server. Docker, Nginx Proxy Manager, WAHA, dan Zolix.
- `192.168.18.101`: LXC Database Server. PostgreSQL dan backup database.

Contoh di bawah memakai Debian 12 untuk VM/LXC.

## 1. Proxmox Host `192.168.18.44`

Buat dua guest:

- VM App Server: minimal 2 vCPU, 4 GB RAM, 40 GB disk, IP statis `192.168.18.100`.
- LXC Database Server: minimal 2 vCPU, 2 GB RAM, 30 GB disk, IP statis `192.168.18.101`.

Jangan install Docker, PostgreSQL, WAHA, atau aplikasi Zolix di host Proxmox.

## 2. LXC Database Server `192.168.18.101`

Install PostgreSQL:

```bash
sudo apt update
sudo apt install -y postgresql postgresql-client gzip
```

Buat user dan database:

```bash
sudo -u postgres psql
```

```sql
CREATE USER zolix_user WITH PASSWORD 'ganti-password-kuat';
CREATE DATABASE zolix_db OWNER zolix_user;
\q
```

Izinkan koneksi hanya dari App Server:

```bash
sudo sed -i "s/^#listen_addresses =.*/listen_addresses = '192.168.18.101'/" /etc/postgresql/*/main/postgresql.conf
echo "host zolix_db zolix_user 192.168.18.100/32 scram-sha-256" | sudo tee -a /etc/postgresql/*/main/pg_hba.conf
sudo systemctl restart postgresql
```

Tes port PostgreSQL dari App Server nanti:

```bash
nc -vz 192.168.18.101 5432
```

## 3. Backup Database di LXC

Salin `scripts/backup-postgres-lxc.sh` ke `/data/zolix/scripts/backup-postgres-lxc.sh`, lalu:

```bash
sudo mkdir -p /data/zolix/scripts /data/zolix/backups
sudo chmod +x /data/zolix/scripts/backup-postgres-lxc.sh
```

Buat file password PostgreSQL agar cron tidak meminta password:

```bash
sudo sh -c 'echo "127.0.0.1:5432:zolix_db:zolix_user:ganti-password-kuat" > /root/.pgpass'
sudo chmod 600 /root/.pgpass
```

Tambah cron backup harian jam 02:15:

```bash
sudo crontab -e
```

```cron
15 2 * * * POSTGRES_USER=zolix_user POSTGRES_DB=zolix_db /data/zolix/scripts/backup-postgres-lxc.sh >> /data/zolix/backups/backup.log 2>&1
```

Tes backup:

```bash
sudo POSTGRES_USER=zolix_user POSTGRES_DB=zolix_db /data/zolix/scripts/backup-postgres-lxc.sh
ls -lh /data/zolix/backups
```

## 4. VM App Server `192.168.18.100`

Install Docker Engine mengikuti repository resmi Docker untuk Debian:

```bash
sudo apt update
sudo apt install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
sudo tee /etc/apt/sources.list.d/docker.sources > /dev/null <<EOF
Types: deb
URIs: https://download.docker.com/linux/debian
Suites: $(. /etc/os-release && echo "$VERSION_CODENAME")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

Siapkan direktori aplikasi:

```bash
sudo mkdir -p /opt/zolix /data/zolix/uploads /data/waha /data/nginx-proxy-manager/data /data/nginx-proxy-manager/letsencrypt
sudo chown -R "$USER:$USER" /opt/zolix /data/zolix /data/waha /data/nginx-proxy-manager
```

Upload atau clone repository Zolix ke `/opt/zolix`, lalu buat env produksi:

```bash
cd /opt/zolix
cp .env.production.example .env
nano .env
```

Isi minimal:

```env
TZ=Asia/Makassar
JWT_SECRET=isi-dengan-secret-panjang
PUBLIC_BASE_URL=http://192.168.18.100:8090
WAHA_API_KEY=isi-dengan-api-key-panjang
WAHA_DASHBOARD_USERNAME=admin
WAHA_DASHBOARD_PASSWORD=isi-dengan-password-dashboard-panjang
WAHA_SESSION=default
DATABASE_URL=postgres://zolix_user:ganti-password-kuat@192.168.18.101:5432/zolix_db?sslmode=disable
```

Jalankan stack:

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f app
```

Endpoint awal:

- Zolix: `http://192.168.18.100:8090`
- Nginx Proxy Manager admin: `http://192.168.18.100:81`
- WAHA dashboard: `http://192.168.18.100:3000/dashboard`

## 5. Nginx Proxy Manager

Di Nginx Proxy Manager, buat Proxy Host untuk Zolix:

- Domain Names: isi domain atau hostname internal.
- Scheme: `http`
- Forward Hostname/IP: `zolix-app`
- Forward Port: `8080`
- Websockets Support: aktifkan bila dibutuhkan.

Jika belum memakai domain publik, akses langsung via `http://192.168.18.100:8090` tetap bisa.

Opsional, buat Proxy Host untuk WAHA dashboard:

- Domain Names: isi domain atau hostname internal WAHA.
- Scheme: `http`
- Forward Hostname/IP: `waha`
- Forward Port: `3000`

## 6. Operasi Harian

Update aplikasi:

```bash
cd /opt/zolix
git pull
docker compose up -d --build
```

Lihat log:

```bash
docker compose logs -f app
docker compose logs -f waha
```

Restart:

```bash
docker compose restart app
```

Restore database dari backup:

```bash
gunzip -c /data/zolix/backups/zolix-YYYYMMDD-HHMMSS.sql.gz | psql -h 127.0.0.1 -U zolix_user -d zolix_db
```
