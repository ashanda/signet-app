# S3: KYC images and daily database backups

Signet uses one **private** S3 bucket, split into two prefixes:

| Prefix | Contents | Written by |
|---|---|---|
| `kyc/` | KYC images (`kyc/nic_front/…`, `kyc/nic_back/…`, `kyc/passport/…`) | the API, on every KYC upload |
| `db-backups/` | `signet_last_2026-09-26T203000Z.sql.gz` … | `signet-db-backup.timer`, daily |

The bucket is never public. The API streams KYC images back through the
same `/storage/kyc/...` URLs it already uses, so the frontend and the paths
stored in the `kycs` table don't change.

The server's IAM key can read and write these two prefixes but **cannot
delete anything**. If the server were compromised, the key still couldn't
wipe the backups.

Commands below assume the server layout from [LIGHTSAIL_SETUP.md](LIGHTSAIL_SETUP.md):
the repo at `/opt/signet` and config in `/opt/signet/backend/.env.production`.

---

## 1. Create the bucket (one time, AWS console)

1. **S3 → Create bucket**
   - Name: e.g. `signet-prod-data` (globally unique)
   - Region: the same region as the Lightsail instance
   - **Block all public access: ON** (leave all four boxes ticked)
   - **Bucket Versioning: Enable**, so an overwritten or deleted object
     can be recovered
   - Default encryption: SSE-S3 (the default)
2. **Lifecycle rules.** Apply [deploy/aws/s3-lifecycle.json](../deploy/aws/s3-lifecycle.json)
   from any machine that has admin AWS credentials:
   ```bash
   aws s3api put-bucket-lifecycle-configuration --bucket signet-prod-data \
     --lifecycle-configuration file://deploy/aws/s3-lifecycle.json
   ```
   This keeps backups 90 days (moved to cheaper storage after 30), drops
   old object versions, and cleans up abandoned uploads. Edit the day
   counts if you need a different retention.

## 2. Create the IAM user for the server

1. **IAM → Policies → Create policy → JSON.** Paste
   [deploy/aws/signet-iam-policy.json](../deploy/aws/signet-iam-policy.json)
   and replace `YOUR_BUCKET` (3 places) with the bucket name. Name it
   `signet-server-s3`.
2. **IAM → Users → Create user** `signet-server` (no console access) and
   attach `signet-server-s3`.
3. **Security credentials → Create access key → "Application running
   outside AWS".** Copy the key ID and the secret.

Lightsail instances can't use IAM instance roles, so an access key is
required. Rotate it by creating a new key, updating the env file,
restarting, and then deleting the old key.

## 3. Configure the server

Add to `/opt/signet/backend/.env.production` (keep it `chmod 600`):

```bash
STORAGE_DRIVER=s3
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_DEFAULT_REGION=ap-south-1        # your bucket's region
AWS_BUCKET=signet-prod-data

# optional, for the backup (defaults shown)
# BACKUP_S3_URI=s3://signet-prod-data/db-backups
# BACKUP_KEEP_LOCAL_DAYS=3
# BACKUP_IGNORE_TABLES=jobs          # see "Large tables" below
```

Install the AWS CLI v2 (used by the backup and for copying files):

```bash
cd /tmp
curl -sSLo awscliv2.zip "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip"
sudo apt install -y unzip && unzip -q awscliv2.zip && sudo ./aws/install
aws --version
```

Check that the key works (the "list" should succeed, even if it's empty):

```bash
set -a; source /opt/signet/backend/.env.production; set +a
aws s3 ls "s3://$AWS_BUCKET/"
```

## 4. Switch KYC images to S3

Deploy the new backend build and restart:

```bash
cd /opt/signet && git pull
cd backend && go build -o signet-api ./cmd/api
sudo systemctl restart signet-api
sudo journalctl -u signet-api -n 20 | grep storage
# expect:  storage: KYC files → s3://signet-prod-data (fallback: local disk)
# a "storage: WARNING: cannot reach ..." line means the key, region or bucket is wrong
```

From here on, **new uploads go to S3**. Existing images still work
because the API falls back to `storage/kyc/` on disk for anything not
in the bucket yet. Now copy the existing images up:

```bash
sudo -E aws s3 sync /opt/signet/backend/storage/kyc "s3://$AWS_BUCKET/kyc" --sse AES256
```

(If images from the old Laravel server haven't been copied to
`/opt/signet/backend/storage/kyc` yet, sync them straight from
`<laravel>/storage/app/public/kyc` instead.)

Check the migration:

1. Open an existing KYC record in the admin panel. The images should load.
2. Submit a new KYC and confirm it landed in S3:
   `aws s3 ls "s3://$AWS_BUCKET/kyc/nic_front/" | tail -3`
3. `sudo journalctl -u signet-api | grep "not in s3"`: after the sync,
   no new lines should appear. Each such line is an image still being
   served from local disk.

Keep `storage/kyc/` on disk until you're satisfied, then archive it. The
fallback only kicks in for files that are missing from S3.

## 5. Enable the daily database backup

```bash
sudo cp /opt/signet/deploy/backup/signet-db-backup.service \
        /opt/signet/deploy/backup/signet-db-backup.timer /etc/systemd/system/
sudo chmod +x /opt/signet/deploy/backup/signet-db-backup.sh
sudo systemctl daemon-reload
sudo systemctl enable --now signet-db-backup.timer

# run one backup right now and watch it
sudo systemctl start signet-db-backup.service
sudo journalctl -u signet-db-backup -n 20 --no-pager
```

A good run ends with:

```
backup: dump ok: 48213377 bytes in 21s
backup: upload verified (48213377 bytes)
backup: done: s3://signet-prod-data/db-backups/signet_last_2026-09-26T203000Z.sql.gz
```

It runs daily at **02:00 Sri Lanka time**. To change that, edit
`OnCalendar=` in the timer, then run `daemon-reload`.

The script:
- dumps with `--single-transaction`, so the site keeps running
- verifies the dump is complete before uploading anything
- checks the uploaded size
- keeps 3 days of local copies in `/var/backups/signet`

If a dump fails or is incomplete, **nothing is uploaded** and the unit is
marked failed.

### Checking it keeps working

```bash
systemctl list-timers signet-db-backup.timer          # NEXT / LAST run times
systemctl status signet-db-backup.service              # last result
aws s3 ls "s3://$AWS_BUCKET/db-backups/" | tail -5     # one new file per day
```

For an alert, Lightsail/CloudWatch can't see systemd. The simplest option
is a weekly glance at the `aws s3 ls` above, or add
`OnFailure=` pointing at a unit that emails you.

### Large tables

The old Laravel app filled `jobs` with millions of unprocessed
`MiningUpdated` rows, and the Go app never reads that table. Set
`BACKUP_IGNORE_TABLES=jobs` to skip it (comma-separate more tables if
needed). A skipped table is left out of the dump completely, and
`go run ./cmd/migrate` recreates it empty after a restore.

## 6. Restore a backup

**Test this on a scratch database first. Never restore over production
without taking a fresh backup of it.**

```bash
set -a; source /opt/signet/backend/.env.production; set +a
aws s3 ls "s3://$AWS_BUCKET/db-backups/" | tail        # pick a file
F=signet_last_2026-09-26T203000Z.sql.gz

sudo mysql -e "CREATE DATABASE signet_restore_test CHARACTER SET utf8mb4"
aws s3 cp "s3://$AWS_BUCKET/db-backups/$F" - | gunzip | sudo mysql signet_restore_test
sudo mysql signet_restore_test -e "SELECT COUNT(*) FROM users; SELECT MAX(created_at) FROM users;"
```

Restore a KYC image version that was overwritten or deleted with:
**S3 console → the object → Versions**.

---

## Local development

Nothing changes locally. `STORAGE_DRIVER` defaults to `local`, so
uploads go to `backend/storage/kyc/` as before. To try S3 locally, set the
same variables in `backend/.env`.
