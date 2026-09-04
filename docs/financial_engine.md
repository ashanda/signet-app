# Financial / MLM commission engine — extracted verbatim logic

Source: `app/Services/WalletService.php`, `app/helpers.php`, `app/Console/Commands/*.php`.
This is the highest-risk area for a faithful port — preserve exact order of operations,
even where it looks like a bug (noted inline). All money fields appear to be stored as
plain numbers (no currency scaling) — DECIMAL/VARCHAR per schema.md, so use decimal/float64
consistently with the original (PHP does float arithmetic here, not BC Math — match that,
don't "fix" it into fixed-point unless asked).

## 1. WalletService::updateWallet($user_id, $amount, $description) — the ONLY sanctioned entry point for crediting a wallet

Callers: CalculateGlobalDirectorShare command, WeeklyActivePackageSum (via RocController approve flow — check RocController in api_spec.md), SalaryController, RocController, DirectShareController, LeaderController, ExecutiveController (per api_spec.md — confirm each call site there).

```
updateWallet(user_id, amount, description):
  myWallet = UserPackage.where(user_id, status='active').with(userpackage) -> get all
  totalValue = sum over myWallet of (pkg.userpackage.price * 4)   # "4x cap" ceiling per user
  wallet = Wallet.where(user_id).first()
  if wallet exists AND totalValue < wallet.balance:
      return   # NO-OP: user already at/above their 4x earnings cap, nothing credited, no log written
  creditWallet(user_id, amount, description)
```

Note the cap check compares `totalValue < wallet.balance` — i.e. it blocks *before* adding
`amount`, using the wallet balance as it stood before this credit. This is a coarse gate, not
a proportional/partial credit — port literally.

### creditWallet($user_id, $amount, $description) — private, only called from updateWallet

```
walletAmount = amount
wallet = Wallet.where(user_id).first()

packageMultipliers = {5000:1.5, 10000:1.5, 25000:1.5, 50000:1.5, 100000:1.5, 500000:1.5, 1000000:1.5}
# (all multipliers are 1.5 today; keep as a map, not a constant, in case values diverge later)

globalShareWallet = GlobalShareWallet.where(user_id).first()
if globalShareWallet is null:
    userPackages = UserPackage.with(userpackage).where(user_id).get()   # ALL packages, any status
    highestPrice = max(price of userPackages where price in packageMultipliers.keys)
    if highestPrice:
        globalShareWallet = GlobalShareWallet.create(user_id, balance=0, max_out = highestPrice * packageMultipliers[highestPrice])

if wallet exists:
    wallet.balance += walletAmount; save
else:
    wallet = Wallet.create(user_id, balance = walletAmount)

EarnLog.create(user_id, amount=walletAmount, description)

if globalShareWallet:
    remaining = globalShareWallet.max_out - globalShareWallet.balance
    if remaining > 0:
        credit = min(walletAmount, remaining)
        globalShareWallet.balance += credit; save
        GlobalShareWalletLog.create(user_id, amount=walletAmount, description='Credited to Global Share Wallet')
        # NOTE: logs the FULL walletAmount even when credit (added to balance) was capped to `remaining` — bug, preserve as-is

checkUserPackage = UserPackage.where(user_id).first()   # NOTE: first(), not filtered by status='active' — arbitrary/oldest package row
if checkUserPackage AND checkUserPackage.earn <= (checkUserPackage.userpackage.price * 4):
    checkUserPackage.earn += walletAmount; save
    # NOTE: earn column is VARCHAR per schema.md — PHP will type-juggle this addition; Go port should
    # parse to float, add, format back to string on write to match column type exactly.
else:
    companyWallet = Wallet.where(user_id=1).first()   # user id 1 = company/house account, assumed to always exist
    companyWallet.balance += walletAmount; save
    EarnLog.create(user_id=1, amount=walletAmount)   # description omitted here (null), unlike the user-side EarnLog above
```

## 2. Genealogy placement algorithm — `ParentFind` / `superParentFind` (app/helpers.php)

This is the binary/matrix placement + "gratitude" super-parent spillover algorithm run when a
user activates a package and needs a tree parent assigned. `user_parents.node` enum values:
`active`, `deactive`, `gratitude`, `correct` (see schema.md notes — widened twice via raw SQL).

Entry point signature: `ParentFind($id, $package, $currentUser, $i, &$createdIds = [])`
- `$id`: the virtual/current node id being evaluated (starts as the referrer/upline id)
- `$package`: package id being activated (used for commission-eligibility check via `checkWalet`)
- `$currentUser`: the user actually activating (the "gratitude_user" in logs)
- `$i`: recursion depth counter (0 or 1 controls whether `checkAndLogFirstTimeSuper` writes a
  `super_parent_logs` row — only writes when `$i == 1`, i.e. first hop into `superParentFind`)
- `$createdIds`: by-reference accumulator of `user_parents` row ids created during this pass

```
ParentFind(id, package, currentUser, i, &createdIds):
  if id == 1:
      return User.where(id=id).first()?.id       # root/company user — stop recursion, return as-is

  activeOrderCount = count UserParent where virtual_id=id and node in (active, gratitude, correct)
  predefinedValues = [1,4,9,14,19,29,39,49,74,99]
  parent = UserParent.where(virtual_id=id).whereIn(node, [active, deactive]).first()

  if parent exists:
      checkActive = User.where(id = parent.virtual_id).first()
      if checkActive exists and checkActive.status == 'inactive':
          return superParentFind(parent.virtual_id, package, currentUser, i, createdIds)

  if activeOrderCount in predefinedValues:
      if parent.virtual_id not in [2, 3, 4]:            # 2,3,4 = reserved/root accounts, excluded from gratitude spillover
          if parent exists:
              newRow = UserParent.create(user_id=id, virtual_id=parent.virtual_id, parent_id=parent.virtual_id, node='gratitude')
              createdIds.append(newRow.id)
              return superParentFind(parent.virtual_id, package, currentUser, i, createdIds)
      else:
          checkActivation = checkWalet(parent.virtual_id, package)
          return checkActivation == 1 ? parent?.virtual_id : (parent ? 1 : null)
              # i.e.: eligible (has capacity) -> place under the parent's virtual_id; not eligible -> fall back to root user id=1

  elseif activeOrderCount > 100:
      nextMultipleOf10 = ceil(activeOrderCount/10)*10 - 1
      if activeOrderCount == nextMultipleOf10:
          return superParentFind(parent.virtual_id, package, currentUser, i, createdIds)
      else:
          return User.where(id=id).first()?.id

  else:  # activeOrderCount not in predefinedValues, and <= 100
      if activeOrderCount == 0:
          return User.where(id=id).first()?.id
      else:
          checkActivation = checkWalet(parent.virtual_id, package)
          return checkActivation == 1 ? User.where(id=id).first()?.id : (parent ? 1 : null)
```

```
superParentFind(virtualParent, package, currentUser, i, &createdIds):
  activeOrderCount = count UserParent where virtual_id=virtualParent and node in (active, gratitude, correct)
  checkAndLogFirstTimeSuper(virtualParent, package, currentUser, i)   # side-effecting log write, see below
  predefinedValues = [2,5,10,15,20,30,40,50,75,100]
  parent = UserParent.where(user_id=virtualParent).first()   # NOTE: keyed by user_id here, not virtual_id
  i += 1

  if parent exists:
      checkActive = User.where(id=parent.user_id).first()
      if checkActive exists and checkActive.status == 'inactive':
          return ParentFind(parent.parent_id, package, currentUser, i, createdIds)

  if activeOrderCount in predefinedValues:
      return ParentFind(parent.parent_id, package, currentUser, i, createdIds)
  elseif activeOrderCount > 100:
      nextMultipleOf10 = ceil(activeOrderCount/10)*10 - 1
      if activeOrderCount == nextMultipleOf10:
          return ParentFind(parent.parent_id, package, currentUser, i, createdIds)
      else:
          return User.where(id=parent.parent_id).first()?.id
  else:
      checkActivation = checkWalet(parent.virtual_id, package)
      return checkActivation == 1 ? User.where(id=virtualParent).first()?.id : (parent ? 1 : null)
```

```
checkAndLogFirstTimeSuper(virtualParent, package, currentUser, i):
  existingLog = super_parent_logs row where current_user_id=virtualParent, package_id=package,
                gratitude_user=currentUser, created_at >= now()-5min
  if not existingLog:
      (in a DB transaction)
      if i == 1:
          insert super_parent_logs(current_user_id=virtualParent, package_id=package, gratitude_user=currentUser, created_at=now(), updated_at=now())
      return true
  return false
```

```
checkWalet(user_id, package_id):   # "eligibility" check — despite the name, this is a commission-affordability gate
  myWallet = UserPackage.where(user_id, status='active').with(userpackage).get()
  totalValue = sum(pkg.userpackage.price * 4 for pkg in myWallet)     # same 4x cap concept as WalletService
  parentWallet = walletBalance(user_id)                               # current wallet.balance, 0 if no wallet row
  parentPackage = UserPackage.where(user_id, status='active').with(userpackage).latest().first()
  feePercentage = parentPackage?.userpackage.commission ?? 0
  refGet = Package.where(id=package_id).first()
  commission = refGet.price * (feePercentage ?? 0) / 100
  predictedValue = parentWallet + commission
  return predictedValue > totalValue ? 0 : 1     # 0 = not eligible (would exceed cap), 1 = eligible
```

## 3. `tokenShare($package, $virtualParent)` — 20% activation bonus paid in TOKENS from the house account

Called on package activation (see TokenController in api_spec.md for the call site/order
relative to ParentFind).

```
tokenShare(package, virtualParent):
  packageData = packages row where id=package  (throw if missing)
  bonusTokens = floor(packageData.price * 0.20)

  tokensToTransfer = tokens where user_id=1 (NOTE: comment says "user_id = 2" but code filters user_id=1 — trust the code)
                     and status='active', order by id asc, limit bonusTokens
  if count(tokensToTransfer) < bonusTokens: throw "Not enough active tokens in user_id 2!"   # message text kept verbatim even though it references the wrong id — user-facing/log text, not logic

  update tokens set user_id = virtualParent, updated_at = now() where id in tokensToTransfer.ids
  # NOTE: token.status is NOT changed here — still 'active', just re-owned

  insert earn_logs(user_id=virtualParent, amount=bonusTokens, description='20% activation bonus')

  wallet = Wallet.where(user_id=virtualParent).first()
  if wallet: wallet.balance += bonusTokens; save
  # NOTE: unlike WalletService.creditWallet, if wallet is missing here NOTHING is created — no Wallet.create fallback. Preserve.

  # Global Share Wallet top-up — same packageMultipliers map and same remaining/credit logic as WalletService.creditWallet,
  # but computed independently here (duplicate logic, not a shared call) and uses bonusTokens as the amount:
  globalShareWallet = GlobalShareWallet.where(user_id=virtualParent).first()
  if null:
      userPackages = UserPackage.with(userpackage).where(user_id=virtualParent).get()
      highestPrice = max(price in packageMultipliers.keys among userPackages)
      if highestPrice: create GlobalShareWallet(user_id=virtualParent, balance=0, max_out=highestPrice*packageMultipliers[highestPrice])
  if globalShareWallet:
      remaining = max_out - balance
      if remaining > 0:
          credit = min(bonusTokens, remaining)
          balance += credit; save
          GlobalShareWalletLog.create(user_id=virtualParent, amount=bonusTokens, description='Credited to Global Share Wallet')
```

## 4. Misc helpers

- `walletBalance(user_id)` → `Wallet.where(user_id).first()?.balance ?? 0`
- `allUsers()` → count of `users` where `status='active' AND role='user'`
- `newActivations()` → count of `user_packages` where `company_status = 0`
- `roc($user_id)` → **echoes HTML directly** (not a return) showing the most recent
  `RocIncomeLog` for the user joined to its `WeeklyPackageSummary` (via `job_id`): week_start,
  week_end, per_week_total. Port as a Go handler/Vue component returning this data as JSON,
  not as server-rendered HTML.
- `rank($user_id)` → **returns an HTML string** (not echoed) computing MLM rank. Logic:
  - `activeUsersDirect` = user_ids where `user_parents.virtual_id = user_id AND node='active'`
  - `activeUsersGrant` = user_ids where `user_parents.parent_id = user_id AND node='gratitude' AND created_at >= 2025-11-22` (hardcoded cutoff date — preserve verbatim, don't make configurable unless asked)
  - `activeUsers` = union of the two (deduped)
  - `totalActivePackagesIds` = active `user_packages.package` for all of `activeUsers`
  - `totalActivePackagesDirect` (a DOLLAR total, confusingly similar name to `totalActivePackagesIds`) = SQL join `user_packages` + `user_parents as up` + `packages as p` where `up.virtual_id = user_id`, `up.node in (active, correct)`, `up.created_at >= 2025-10-01` (hardcoded), `user_packages.status='active'`, summed `p.price` — this is "Team Sales" in the UI
  - `totalActivePackages` = sum of `package.price * occurrence_count` for `totalActivePackagesIds` (this is the value actually compared against rank thresholds, NOT `totalActivePackagesDirect`)
  - `gratuityUsers` = `SuperParentLog.where(current_user_id=user_id).pluck(gratitude_user)`
  - `totalActiveSuperPackages` = sum of `(package.price * 0.2) * occurrence_count` for active packages of `gratuityUsers` — this is "Gratitude" in the UI, and is also the value compared against the `super` rank threshold
  - Rank ladder (name → {team threshold, super threshold}), first tier (in order) where BOTH `totalActivePackages >= team` AND `totalActiveSuperPackages >= super` wins; loop breaks at the first tier NOT met and records that as `nextRank` with `remainingTeam`/`remainingSuper` (max(target-current,0)):
    ```
    Crystal:                   team 5000,    super 100
    Jade:                      team 10000,   super 200
    Emerald:                   team 20000,   super 300
    Ruby:                      team 30000,   super 500
    Diamond:                   team 100000,  super 1000
    Senior Diamond:             team 250000,  super 2000
    Senior Executive Diamond:   team 500000,  super 5000
    Crown Diamond:               team 1000000, super 10000
    ```
    Default `currentRank = 'No Rank'` if no tier met.
  - Rendered fields (port as a JSON payload, values only, styling is a Vue concern): Team Sales
    (=`totalActivePackagesDirect`), Gratitude (=`totalActiveSuperPackages`), Current Rank, Next
    Rank (or "N/A"), Remaining Team, Remaining Gratitude.

## 5. Scheduled/console jobs

### `share:calculate` — CalculateGlobalDirectorShare (monthly Global Director profit-share payout)
```
period = current "Y-m"
if GlobalDirectorShareDistribution.where(period).exists(): abort "already distributed"
totalPool = sum(PackagePool.pool_amount) where created_at in current year+month AND pool_amount > 0
if totalPool <= 0: abort
totalShares = sum(User.global_director_share) where global_director_share_status = 1
if totalShares <= 0: abort
valuePerShare = totalPool / totalShares
for each user where global_director_share_status = 1:
    amount = round(user.global_director_share * valuePerShare, 2)
    WalletService.updateWallet(user.id, amount, "Global Director Share")   # subject to the 4x cap gate in section 1
    on failure: collect user id, log, continue (does not abort the whole run)
if no failures: insert GlobalDirectorShareDistribution(period, total_pool, total_shares, value_per_share, user_count=succeeded)
# if any failures: period is NOT marked distributed, safe to re-run (will re-pay everyone, including already-paid users — no per-user idempotency, only whole-period)
```

### `mining:send-webhook` — SendMiningWebhook (outbound webhook w/ retry backoff, run frequently e.g. every minute via scheduler — confirm actual cron in routes/console.php)
```
if !config(services.mining_webhook.enabled): skip
acquire Cache::lock('mining-webhook-lock', 55s) or skip (prevents overlap)
webhookUrl/secret from config(services.mining_webhook.*)  (env MINING_WEBHOOK_ENABLED/URL/SECRET)
maxAttempts = 5
records = user_minings where webhook_sent_at IS NULL
          AND (webhook_status='pending' OR (webhook_status='failed' AND (next_retry_at IS NULL OR next_retry_at <= now())))
          AND webhook_attempts < 5
          order by id, limit 50
for each record:
    POST webhookUrl, headers {Accept, Content-Type, X-Webhook-Secret: secret}, 20s timeout, body:
        {id, user_id: record.secretKey.secret_key (the OPAQUE hashed secret key string, not the numeric user id!), total_token, mining_token, daily_mining, status, created_at, updated_at}
    on 2xx: set webhook_sent_at=now, webhook_status='sent', webhook_response=body, next_retry_at=null
    on non-2xx or exception:
        attempts += 1
        status = attempts >= 5 ? 'permanently_failed' : 'failed'
        next_retry_at = attempts>=5 ? null : backoff(attempts)   # 1:+1m 2:+5m 3:+15m 4:+30m default:+1h
        save webhook_status, webhook_attempts, next_retry_at, webhook_response=error/body
```

### `mining:update` — UpdateMiningTokens (accrue mining tokens; run every minute)
```
for each user_minings row where status='active' AND daily_mining > 0:
    row = fresh fetch of that user's user_minings row (re-read, not the loop variable — matches if only one active row per user)
    perMinute = row.daily_mining / 1440
    newToken = row.mining_token + perMinute
    if newToken >= row.total_token:
        newToken = row.total_token
        update: mining_token=newToken, status='inactive'
    else:
        update: mining_token=newToken   (status unchanged)
    broadcast MiningUpdated(user_id, {mining_token: round(newToken,8), status: newToken>=total_token?'inactive':'active', progress: newToken/total_token*100})
    # broadcast channel/payload shape — see app/Events/MiningUpdated.php; Pusher-based in original,
    # port as a WS/SSE push or poll-based endpoint (see architecture notes)
```

### `packages:weekly-sum` — WeeklyActivePackageSum (weekly ROC distribution, Monday-Sunday)
```
jobId = "JOB-{year}W{isoWeekNumber}-{5 random uppercase chars}"
weekTotal = sum(packages.price) joined user_packages where status='active' and activated_at in [Monday 00:00, Sunday 23:59:59]
perWeekTotal = weekTotal * 0.05                 # 5% of weekly activated volume
total = perWeekTotal                            # balance_forward carry-over is CODE-DISABLED (commented out) — do not carry forward, even though weekly_package_summaries.balance_forward column exists and is still written
if total <= 0: skip, no row written

activeRoc = users where roc_status='active' (with wallet)
rocPotation = total / count(activeRoc)   (0 if count=0)

for each user in activeRoc:
    package = user's active user_packages row (skip user if none)
    packagePrice = package.userpackage.price
    maxPayout = match packagePrice: 100->5, 500->25, 1000->50, default->0   # per-package weekly payout cap
    if user.wallet.roc_balance >= packagePrice * 5:
        set user.roc_status = 'stopped'   # lifetime 5x package price ROC cap reached
        continue   # no payout this run
    if rocPotation > maxPayout:
        wallet.roc_balance += maxPayout        # capped payout
        remainingBalance += (rocPotation - maxPayout)   # accumulates but is NOT reused this run (see total above) — still stored on the summary row for visibility/audit
        storeEarnLogs(user.id, maxPayout, jobId)
    else:
        wallet.roc_balance += rocPotation      # full share
        storeEarnLogs(user.id, rocPotation, jobId)

insert weekly_package_summaries(job_id, week_start, week_end, per_week_total=weekTotal, balance_forward=remainingBalance, roc_potation=floor(rocPotation), total_amount=total)

storeEarnLogs(userId, amount, jobId):
    insert earn_logs(user_id, amount, description='ROC Income')  -> capture insert id
    insert roc_income_log(job_id, user_id, amount, description='ROC Income', status='pending', earn_log_id=<that id>)
    # NOTE: this credits wallet.roc_balance directly via raw update earlier, NOT via WalletService — the
    # 4x cap gate in section 1 does NOT apply to ROC income. roc_balance is a separate ledger from wallet.balance;
    # check RocController (api_spec.md) for how/when roc_balance gets moved into spendable wallet.balance (likely an "approve" action).
```

### `users:generate-secret-keys` — GenerateOldUserSecretKeys (one-off backfill)
```
for each user in chunks of 100:
    if UserSecretKey exists for user: skip
    plainKey = "USER-{id}-" + random(40)
    UserSecretKey.create(user_id, secret_key = bcrypt(plainKey))   # NOTE: the plaintext key is never stored/shown here — this command as written has no way to hand the plaintext back to the user. Check UserController/AuthController (api_spec.md) for the real key-issuance flow used at registration; this command is a backfill for pre-existing users only and its output plaintext is effectively lost. Port behavior as-is; flag to product owner.
    else: skip count++
```
