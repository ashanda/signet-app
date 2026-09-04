# API Specification — crypto-app-src (Laravel 12)

Source of truth: `routes/web.php`, `routes/api.php`, `app/Http/Controllers/**`, `app/Http/Middleware/RoleMiddleware.php`.

Scope note: this document covers routing, auth/role gating, request validation, business-logic sequencing, and response shapes for every controller method wired to a route. The underlying commission/wallet/mining/ROC math formulas are being extracted by a separate agent — here every place such a calculation is invoked is flagged with the call and its arguments, not re-derived.

Helper functions referenced by controllers (defined in `app/helpers.php`, autoloaded globally via composer `files`): `ParentFind($id,$package,$currentUser,$i,&$createdIds)`, `superParentFind(...)`, `checkAndLogFirstTimeSuper(...)`, `tokenShare($package,$virtualParent)`, `walletBalance($user_id)`, `checkWalet($user_id,$package_id)`, `rank($user_id)`, `allUsers()`, `roc($user_id)`, `newActivations()`. These are treated as black boxes here (math owned by another agent) — only call sites and arguments are noted.

---

## Route table overview

### `routes/web.php`

| Method | URL | Name | Controller@Method | Middleware / role |
|---|---|---|---|---|
| GET | /dashboard | (none, closure) | closure — role-based redirect to `<role>.dashboard` or `guest` | none (checks `auth()->check()` itself) |
| GET | / | guest | AuthController@guest | none |
| GET | /login | login | AuthController@showLoginForm | none |
| POST | /login | login.post | AuthController@login | none |
| POST | /logout | logout | AuthController@logout | auth |
| GET | password/reset | password.request | PasswordResetController@showResetRequestForm | none |
| POST | password/email | password.email | PasswordResetController@sendResetLink | none |
| GET | password/reset/{token} | password.reset | PasswordResetController@showResetForm | none |
| POST | password/reset | password.update | PasswordResetController@reset | none |
| GET | /register | register.step1 | AuthController@showStep1 | none |
| POST | /register/step1 | register.processStep1 | AuthController@processStep1 | none |
| GET | /register/step2/{id} | register.step2 | AuthController@showStep2 | none |
| POST | /register/step2 | register.processStep2 | AuthController@processStep2 | none |
| GET | /register/step3 | register.step3 | AuthController@showStep3 | none |
| **Group: `auth`,`role:company`** | | | | |
| GET | /company/dashboard | company.dashboard | CompanyController@index | auth, role:company |
| POST | /company/activate/{user} | company.activateUser | CompanyController@activateUser | auth, role:company |
| GET | /view-tokens/{userId} | view.tokens | TokenController@viewTokens | auth, role:company |
| POST | /generate-tokens/{userId} | generate.tokens | TokenController@generateTokens | auth, role:company |
| GET | /company/pending-activation | company.pending.activation | CompanyController@pendingActivation | auth, role:company (route declared twice, identical) |
| POST | /company/new_active-package | company.newActivepackage | CompanyController@newActivepackage | auth, role:company |
| GET | /admin/{userId}/setup-google-auth | setup.google.auth | GoogleAuthenticatorController@setupGoogleAuthenticator | auth, role:company |
| resource | packages | packages.* | PackageController (index/create/store/show*/edit/update/destroy) | auth, role:company |
| GET | /kyc/verified | kyc.verified | KycController@verified | auth, role:company |
| GET | /mining/users | mining.users | MiningController@index | auth, role:company |
| GET | /mining/search/{id} | (none) | MiningController@search | auth, role:company |
| POST | /mining/update/{id} | (none) | MiningController@update | auth, role:company |
| POST | /users/update/{id} | (none) | UserController@update | auth, role:company |
| POST | /users/update-roc/{id} | (none) | UserController@updateRocStatus | auth, role:company |
| POST | /users/update-leader-status/{id} | (none) | UserController@updateLeaderStatus | auth, role:company |
| POST | /users/update-global-director-share/{id} | (none) | UserController@updateGlobalDirectorShare | auth, role:company |
| GET | /roc | company.roc | RocController@rocIncome | auth, role:company |
| POST | /company/roc/status-update | company.updateRocStatus | RocController@updateRocStatus | auth, role:company |
| GET | /direct-share | company.direct_share | DirectShareController@directShare | auth, role:company |
| GET | /direct-share-log | company.direct_share_log | DirectShareController@directShareLog | auth, role:company |
| resource | package-pools | package-pools.* | DirectShareController (index/store/show*/update/destroy — no dedicated `index`/`create`/`edit` methods defined, so those verbs 404 at runtime) | auth, role:company |
| GET | user-parent-logs | userparentlogs.index | UserParentMapsLogController@index | auth, role:company |
| DELETE | user-parent-logs/{log} | userparentlogs.destroy | UserParentMapsLogController@destroy | auth, role:company |
| GET | /new-activations | new.activations | CompanyController@newActivations | auth, role:company |
| GET | /salaries | salaries.index | SalaryController@index | auth, role:company |
| GET | /salaries/search-users | salaries.searchUsers | SalaryController@searchUsers | auth, role:company |
| POST | /salaries | salaries.store | SalaryController@store | auth, role:company |
| DELETE | /salaries/{id} | salaries.destroy | SalaryController@destroy | auth, role:company |
| GET | /leaders/gain | leaders.gain | LeaderController@index | auth, role:company |
| GET | /executives/gain | executives.gain | ExecutiveController@index | auth, role:company |
| GET | /leadership-bonus-log | company.leadership_bonus_log | ExecutiveController@leadershipBonusLog | auth, role:company |
| resource | /countries | countries.* | CountriesController (index/store/update/destroy defined; create/edit/show not defined → 404) | auth, role:company |
| **Group: `auth`,`role:company,admin`** | | | | |
| GET | /users | company.users | UserController@allUsers | auth, role:company OR admin |
| GET | /users/search/{id} | (none) | UserController@search | auth, role:company OR admin |
| POST | /users/update-leader-code/{id} | (none) | UserController@updateLeaderCode | auth, role:company OR admin |
| POST | /users/update-executive-code/{id} | (none) | UserController@updateExecutiveCode | auth, role:company OR admin |
| GET | /leader-code-logs | leader.code.logs | UserController@leaderCodeLogs | auth, role:company OR admin |
| GET | /executive-code-logs | executive.code.logs | UserController@executiveCodeLogs | auth, role:company OR admin |
| **Group: `auth`,`role:admin`** | | | | |
| GET | /admin/dashboard | admin.dashboard | AdminController@index | auth, role:admin |
| POST | /admin/activate/{user} | admin.activateUser | AdminController@activateUser | auth, role:admin |
| **Group: `auth`,`role:agent`** | | | | |
| GET | /agent/dashboard | agent.dashboard | AgentController@index | auth, role:agent |
| POST | /admin/activate/{user} | admin.activateUser (duplicate name — see note) | AdminController@activateUser | auth, role:agent |
| **Group: `auth`,`role:user`** | | | | |
| GET | /user/dashboard | user.dashboard | UserController@index | auth, role:user |
| POST | /admin/activate/{user} | admin.activateUser (duplicate name — see note) | AdminController@activateUser | auth, role:user |
| **Ungrouped (only `auth` via `Auth::check()` implicitly inside controller methods — NOT enforced by route middleware)** | | | | |
| GET | /earn/history | earn.history | EarnLogController@index | none declared — relies on `auth()->user()->id` inside; would 500/error for guests rather than redirect |
| POST | /active-package | active.package | TokenController@activePackage | none declared — same risk |
| GET | /token-shares | token.share | TokenController@shareToken | none declared |
| POST | /token/share | token.shares | TokenController@shareTokens | none declared |
| GET | /token/share/logs | token.share.log | TokenController@shareTokensLog | none declared |
| GET | /buy-package-history | buy.package.history | PackageController@buyPackageHistory | none declared |
| GET | /buy-package | buy.package | PackageController@buyPackage | none declared |
| POST | /buy-packages | buy.packages | PackageController@buyPackages | none declared |
| GET | /my-geneology | my.geneology | GeneologyController@index | none declared |
| GET | /geneology/{userId} | geneology.show | GeneologyController@viewGeneology | none declared |
| GET | /kyc | kyc.index | KycController@index | none declared |
| GET | /kyc/show | kyc.show | KycController@show | none declared |
| GET | /kyc/create | kyc.create | KycController@create | none declared |
| GET | /kyc/{id}/edit | kyc.edit | KycController@edit | none declared |
| POST | /kyc | kyc.store | KycController@store | none declared |
| PUT | /kyc/{id} | kyc.update | KycController@update | none declared |
| DELETE | /kyc/{id} | kyc.destroy | KycController@destroy | none declared |
| POST | /kyc/{id}/verify | kyc.verify | KycController@verify | none declared — anyone logged in (or not) can verify KYC! |
| POST | /kyc/{id}/unverify | kyc.unverify | KycController@unverify | none declared |
| POST | /toggle-vacation | user.toggleVacation | UserController@toggleVacation | none declared |
| GET | /mining/user/{userId} | (none) | MiningController@getUserMiningData | none declared |

**IMPORTANT — route-name collision:** `admin.activateUser` is registered three times (role:admin group, role:agent group, role:user group), all pointing at `AdminController@activateUser` and all POST `/admin/activate/{user}`. In Laravel, the last-registered route with a given name wins for `route('admin.activateUser')` URL generation, but path-based dispatch matches the first route whose pattern+method fits — since all three have an identical path/verb, only the **first registered** (`role:admin` group's) is ever reachable; the `role:agent` and `role:user` copies are dead routes. Flag this for the Go port: only admins can hit this endpoint in practice.

**IMPORTANT — many endpoints have no `auth` middleware at all** (everything after the four role groups, from `/earn/history` down to `/mining/user/{userId}`). These rely entirely on `auth()->user()` / `Auth::user()` being non-null inside the controller body; for a guest, PHP will throw (calling a method on null) rather than a clean redirect/403. **In the Go rewrite, decide explicitly whether these require session auth** — the Laravel behavior is "crash if not logged in," not "any role can access," but there is no controlled 401/redirect. Treat this as an inconsistency to resolve, not a spec to literally replicate.

### `routes/api.php` (Sanctum, `ApiUser` guard)

| Method | URL | Controller@Method | Middleware |
|---|---|---|---|
| POST | /api/token | Api\AuthController@token | none (public) |
| POST | /api/check-user | Api\AuthController@checkUser | auth:sanctum |
| GET | /api/users | Api\AuthController@getAllUsers | auth:sanctum |
| POST | /api/user/details | Api\AuthController@getSpecificUser | auth:sanctum |

---

## RoleMiddleware

`app/Http/Middleware/RoleMiddleware.php`:
```php
public function handle(Request $request, Closure $next, $role)
{
    $allowedRoles = explode(',', $role);
    if (!Auth::check() || !in_array(Auth::user()->role, $allowedRoles, true)) {
        abort(403, 'Unauthorized action.');
    }
    return $next($request);
}
```
- Registered as route middleware alias `role`, invoked as `role:company`, `role:company,admin`, etc. Comma-separated role list, strict (`===`) string comparison against `User::role`.
- Unauthenticated OR wrong-role → HTTP 403 with message "Unauthorized action." (Laravel's default abort page/JSON depending on `Accept` header).
- No middleware registration file (`Kernel.php` / `bootstrap/app.php`) is present in this trimmed repo — assume standard Laravel 12 `bootstrap/app.php` `->withMiddleware()` aliasing `'role' => RoleMiddleware::class`. Go port: implement as a per-route/group guard checking session user's `role` column against an allow-list, returning 403 with that exact message string if replicating error bodies.

---

## `guards` / auth config

- `web` guard: session, `App\Models\User` provider — used by all `routes/web.php` endpoints (cookie/session auth via `Auth::attempt`, `Auth::user()`, `Auth::check()`).
- `api` guard: `sanctum` driver, provider `api_users` → `App\Models\ApiUser` model (separate from `User`!). Used only by `routes/api.php`. Tokens minted via `Sanctum`'s `createToken()->plainTextToken` in `Api\AuthController@token`.

---

## AuthController

### GET / (guest) — AuthController@guest
- **Auth/role required:** none (public)
- **Request input:** none
- **Business logic steps:** none
- **Response:** view `welcome`, no data.

### GET /login (login) — AuthController@showLoginForm
- **Auth/role required:** none
- **Request input:** none
- **Business logic steps:** none
- **Response:** view `auth.login`.

### POST /login (login.post) — AuthController@login
- **Auth/role required:** none (guest)
- **Request input:**
  - `email`: `required|email`
  - `password`: `required`
- **Business logic steps:**
  1. `Auth::attempt($credentials)`.
  2. On success: `User::where('email', $request->email)->first()`.
  3. `UserPackage::where('user_id', $checkUser->id)->count()` → `$checkPackage`.
     - If count `> 1`: redirect to `route(Auth::user()->role . '.dashboard')`.
     - If count `== 1`: look up `UserPackage::where('user_id',...)->where('status','pending')->first()`.
       - If a pending package exists: `Alert::error('Error', "You're package is pending")`, redirect back.
       - Else: redirect to `route(Auth::user()->role . '.dashboard')`.
     - Else (count `== 0`): `Alert::error('Error', 'You do not have any package')`, redirect back. **Note: user IS authenticated at this point (session established) even though they're redirected back with an error — this is not a logout.**
  4. If `Auth::attempt` fails: `back()->with('error', 'Invalid credentials')`.
- **Response:** redirect to `{role}.dashboard` route, or redirect back with flash `error` message / SweetAlert error. No JSON.

### POST /logout (logout) — AuthController@logout
- **Auth/role required:** auth (any logged-in user)
- **Request input:** none
- **Business logic steps:** `Auth::logout()`.
- **Response:** redirect to `login`.

### GET /register (register.step1) — AuthController@showStep1
- **Auth/role required:** none (guest, but reads referral code)
- **Request input:** query param `ref` (referral code, optional in signature but effectively required for success path)
- **Business logic steps:**
  1. `ReferralCode::where('code', $referralCode)->first()` → `$checkReferral`.
  2. `Countries::all()`.
  3. `User::where('leader_status','active')->get()` → `$leaders`.
  4. If no matching referral code: `Alert::error('Error','Invalid referral code')`, redirect to `login`.
  5. Else: find the newest user referred by that code: `User::where('referred_by', $checkReferral->user_id)->latest('id')->first()` → `$lastRefUser`.
     - If found, check `UserPackage::where('user_id',$lastRefUser->id)->where('status','active')->first()`.
       - If none: `Alert::error('Error', 'The last user in this referral does not have an active package. Please contact your referral person.')`, redirect to `login`.
- **Response:** view `auth.register_step1` with `referralCode`, `countries`, `leaders` (only reached if referral valid and gate above passed).

### POST /register/step1 (register.processStep1) — AuthController@processStep1
- **Auth/role required:** none (guest)
- **Request input (validated):**
  - `email`: `required|email`
  - `password`: `required|confirmed|min:6`
  - `whatsapp_number`: `required`
  - `country_code`: `required`
  - `binance_pay_id`: `required`
  - `referral_code`: `required|exists:referral_codes,code`
  - `country`: `required`
  - `leader_code`: `nullable|exists:users,id`
  - `executive_code`: `nullable|exists:users,id|different:leader_code`
  - (`name` is used when creating the User but is **not validated** at all — silently null if omitted)
- **Business logic steps:**
  1. `$whatsapp = $request->country_code . $request->whatsapp_number`.
  2. `ReferralCode::where('code', $request->referral_code)->first()` → `$virtualParentUser`; if missing, error + redirect back (defensive; validation already guarantees existence via `exists:referral_codes,code`).
  3. Duplicate whatsapp check: `User::where('whatsapp_number', $whatsapp)->first()`; if found → `Alert::error('Error','The WhatsApp number is already registered.')`, redirect back.
  4. Duplicate binance pay id check: `User::where('binance_pay_id', $request->binance_pay_id)->first()`; if found → error "The Binance Pay ID is already registered.", redirect back.
  5. Duplicate email check: `User::where('email', $request->email)->first()`; if found → error "Email already exists", redirect back. (Dead code follows: `if ($checkUser && $checkUser->id)` branch is unreachable because the `else` above only runs when `$checkUser` is falsy — noted as inert legacy logic, not to be replicated.)
  6. `$referredBy = $virtualParentUser->user_id`.
  7. Create `User`: `name` (unvalidated/likely null), `email`, `password` = `bcrypt($request->password)`, `whatsapp_number` = concatenated, `binance_pay_id`, `status = 'pending'`, `leader_code` = `ltrim($request->leader_code,'0') ?: ''`, `executive_code` = `ltrim($request->executive_code,'0') ?: ''`, `referred_by` = `$referredBy`, `country_id` = `$request->country`.
  8. Create `Wallet` for the new user with `balance = 0`.
  9. Generate referral code: `strtoupper(Str::random(6))`, create `ReferralCode` row for the new user.
  10. Generate secret key: `'USER-' . $user->id . '-' . Str::random(40)`, store `Hash::make($plainKey)` in `UserSecretKey`.
  11. Redirect to `register.step2` with `id = $newUserID`.
- **Response:** redirect to `register.step2` route (GET) with the new user id, or various error redirects-back with SweetAlert flash.

### GET /register/step2/{id} (register.step2) — AuthController@showStep2
- **Auth/role required:** none
- **Request input:** route param `id`
- **Business logic steps:** `Package::where('status','active')->get()`.
- **Response:** view `auth.register_step2` with `id`, `package`.

### POST /register/step2 (register.processStep2) — AuthController@processStep2
- **Auth/role required:** none
- **Request input:**
  - `package`: `required`
  - `newUserID`: `required`
- **Business logic steps (long — MLM tree placement):**
  1. `$id = $request->newUserID`; load `$userData = User::where('id',$id)->first()`.
  2. Call helper `ParentFind($userData->referred_by, $request->package, $request->newUserID, 1, $createdIds)` → `$parent_id`. **Flagged for the math-extraction agent**: this determines tree placement and is central to commission structure.
  3. If `$parent_id == 1` (company root sentinel): `$parentId = 1`, `$userparent = $userData->referred_by`.
     Else: `$parentId = $parent_id`, `$userparent = $parentId`; also computes `$myWallet`/`$totalValue` (sum of `price*4` across parent's active packages) and `$parentWalet = walletBalance($parent_id)` — **these two locals are computed but never used afterward** (dead code, but flag: might indicate incomplete wallet-cap check that was removed).
  4. Create `UserParent` row: `user_id=$id`, `virtual_id=$userData->referred_by`, `parent_id=$userparent`.
  5. If `$userData->referred_by` is set, run inside `DB::transaction`:
     - Query `UserParent::where('virtual_id', $userData->referred_by)->where('created_at', $createdAt)->orderBy('created_at','asc')->lockForUpdate()->get()`.
     - If `count() >= 2`: treat first as "gratitude" (oldest), last as "active" (newest). Update the gratitude row's `user_id`/`virtual_id`/`parent_id` to copy the active row's values (`updated_at = now()`). Create a new `UserParent` row with `node = 'correct'` copying active row's `user_id`/`parent_id` (as `virtual_id` and `parent_id`). Delete the original "active" row.
     - This is the two-child-per-parent "binary spillover" placement correction — **flag for math agent**, this affects who ultimately earns commission.
  6. Reload `$findParent = UserParent::where('user_id',$user)->first()`, `$parentData = User::where('id',$parentId)->first()`.
  7. Guard: if user already has an `active` `UserPackage`, error "You already have an active package", redirect to `login`.
  8. `$UserBuyPackage = Package::where('id',$request->package)->first()`.
  9. `UserPackage::firstOrNew(['user_id'=>$user])`: set `package = $UserBuyPackage->id`, `status = 'pending'`, `ref_id = $findParent->parent_id`, `sale = 'first'`, save.
  10. If `SuperParentLog` row exists for `gratitude_user = $request->newUserID` created within the last minute, update its `user_package` to the new `UserPackage` id (links a just-created super-parent gratitude log entry to this package).
  11. Create `UserParentMapsLog`: `user_id`, `parent_id`, `created_row_ids` (array of all UserParent ids touched above), `note = 'UserParent rows created in processStep2'`.
  12. Global Share Wallet bootstrap: hardcoded `$packageMultipliers` map (5000..1,000,000 → 1.5). If the new user has no `GlobalShareWallet` row yet, compute their highest eligible package price from `UserPackage::with('userpackage')` and, if it matches a multiplier key, create a `GlobalShareWallet` with `balance=0`, `max_out = highestPrice * multiplier`.
  13. `Mail::to($parentData->email)->queue(new NewUserPackageMail($userData, $parentData, $UserBuyPackage))`.
  14. Return view `auth.register_step3` with `parentData`, `user` (does NOT redirect, renders directly — form re-post safe issue for the Go/Vue rewrite to consider).
- **Response:** view `auth.register_step3` compact `parentData`, `user`; or early redirect to `login` with error alert if duplicate active package.

### GET /register/step3 (register.step3) — AuthController@showStep3
- **Auth/role required:** none
- **Request input:** `id` (read via `$request->id`, i.e. query string `?id=`)
- **Business logic steps:** `User::find($user_id)`.
- **Response:** view `auth.register_step4` (note: named step3 route renders the `register_step4` blade) with `user`.

### (unrouted) showStep4 — AuthController@showStep4
- Not wired to any route in `web.php`. Reads `$request->id`, loads user; if `status == 'active'` redirects to `login`; else returns view `auth.register_step3`. Dead/unreachable controller method — flag as such (may have been superseded by showStep3).

---

## Auth/PasswordResetController

### GET password/reset (password.request) — PasswordResetController@showResetRequestForm
- **Auth/role required:** none
- **Response:** view `auth.password_reset`.

### POST password/email (password.email) — PasswordResetController@sendResetLink
- **Auth/role required:** none
- **Request input:** `email`: `required|email|exists:users,email`
- **Business logic steps:**
  1. Generate `$token = Str::random(60)`.
  2. `DB::table('password_reset')->updateOrInsert(['email'=>...], ['token'=>$token,'created_at'=>now()])` — raw table, **not** Laravel's built-in `password_resets` broker table/hashing; no expiry check anywhere in this controller.
  3. `Mail::to($request->email)->send(new PasswordResetMail($token))` — sent synchronously (not queued).
- **Response:** redirect back with flash `status` = "We have emailed your password reset link!".

### GET password/reset/{token} (password.reset) — PasswordResetController@showResetForm
- **Auth/role required:** none
- **Response:** view `auth.password_reset_form` with `token`.

### POST password/reset (password.update) — PasswordResetController@reset
- **Auth/role required:** none
- **Request input:**
  - `email`: `required|email`
  - `password`: `required|confirmed|min:8`
  - `token`: `required`
- **Business logic steps:**
  1. `DB::table('password_reset')->where('token', $request->token)->first()` → `$reset`; if not found, `back()->withErrors(['token' => 'This password reset token is invalid.'])`.
  2. `User::where('email', $reset->email)->first()`; if not found, `back()->withErrors(['email' => 'No user found with that email address.'])`. Note: the submitted `email` field is validated but never actually compared to `$reset->email` — token alone determines the target user.
  3. `$user->update(['password' => Hash::make($request->password)])`.
  4. Delete all `password_reset` rows for that email.
  5. `Alert::toast('Your password has been reset!', 'success')`.
- **Response:** redirect to `login` with flash `status` = "Your password has been reset!".

---

## AdminController

### GET /admin/dashboard (admin.dashboard) — AdminController@index
- **Auth/role required:** auth, role:admin
- **Request input:** none
- **Business logic steps:**
  1. Build referral link from `ReferralCode` for the current admin user.
  2. `Token::selectRaw(...)` grouped by `user_id` with active/deactive counts, `with('user')`, paginate 10 → `$tokenCounts`.
  3. `$myTokens` = count of current user's active tokens.
  4. `$myWallet` = current user's active `UserPackage`s with `userpackage` relation; `$totalValue` = sum(`price*4`).
  5. `$myPackage` = current user's highest-price active package (join on packages, order by price desc). `$feePercentage` = that package's `commission`.
  6. `$poolAmount` = sum of `PackagePool.pool_amount` for the current month/year.
  7. `$totalPoolshareValue` = sum of `User.global_director_share` where `global_director_share_status = 1`.
  8. `$myshareValue` = `($poolAmount / $totalPoolshareValue) * auth()->user()->global_director_share` if `$totalPoolshareValue > 0`, else `0`. **Flag for math agent.**
  9. `$myGlobleDirectorShare` = `GlobalShareWalletLog::where('user_id', auth()->id())->first()` (note: `->first()` not filtered/ordered — arbitrary single row, likely a bug vs. intending latest).
  10. `$activations` = pending `UserPackage`s where `ref_id = auth()->id()`, with `user`,`userpackage`, paginate 10.
- **Response:** view `admin.dashboard` with all the above compacted.

### POST /admin/activate/{user} (admin.activateUser) — AdminController@activateUser
- **Auth/role required:** auth, role:admin (only the first-registered route with this path/verb is reachable — see routing note above)
- **Request input:** route-model-bound `User $user`
- **Business logic steps:** `$user->status = 'active'; $user->save();`
- **Response:** redirect to `admin.dashboard`. No wallet/commission side-effects here (unlike `TokenController::activePackage` / `processPackageActivation*`) — this only flips the `User.status`, not the package. Flag: this looks like a legacy/parallel activation path distinct from the token-based `active.package` flow.

### GET /company/pending-activation (also registered under admin file's method) — AdminController@pendingActivation
- Not routed in `web.php` (route commented out under `role:admin` group) — dead/unreachable code (there IS a routed method with the same name on `CompanyController`, don't confuse them).
- Logic: current user's active `UserPackage` + `feePercentage`; `$activations` = pending UserPackages where `ref_id = auth()->id()`, paginate 10. View `admin.pending-activation`.

---

## AgentController

### GET /agent/dashboard (agent.dashboard) — AgentController@index
- **Auth/role required:** auth, role:agent
- **Request input:** none
- **Business logic steps:** identical shape to `AdminController@index` minus the `Token::selectRaw` paginated breakdown (`$tokenCounts` not computed here): ref link, `$myTokens`, `$myWallet`/`$totalValue`, `$myPackage`/`$feePercentage`, `$poolAmount` (current month), `$totalPoolshareValue`, `$myshareValue` (same formula, flagged), `$myGlobleDirectorShare = GlobalShareWallet::where('user_id',auth()->id())->first()` (note: queries `GlobalShareWallet`, not `GlobalShareWalletLog` as Admin/User controllers do — inconsistent model use, verify intent), `$activations` = pending UserPackages with `ref_id = auth()->id()`.
- **Response:** view `agent.dashboard` compact.

---

## CompanyController

### GET /company/dashboard (company.dashboard) — CompanyController@index
- **Auth/role required:** auth, role:company
- **Request input:** optional query `from_date`, `to_date`
- **Business logic steps:**
  1. Ref link via `ReferralCode`.
  2. `$tokenCounts` — same grouped/paginated token breakdown as Admin.
  3. `$myTokens` — active token count for current user.
  4. `$myWallet = Wallet::where('user_id', auth()->id())->first()` (different from Admin/Agent — here it's the raw Wallet row, not the active-UserPackage collection).
  5. `$packageWiseQuery` = active `UserPackage`s, optionally filtered by `activated_at` between `from_date`/`to_date` (start/end of day).
  6. `$packageWiseCounts` = join to `packages`, group by package/name/price, select `active_count` and `total_value = price * COUNT(*)`.
  7. `$grandTotal` = sum of `total_value` across `$packageWiseCounts`.
- **Response:** view `company.dashboard` with `refLink,tokenCounts,myTokens,myWallet,packageWiseCounts,grandTotal`.

### POST /company/activate/{user} (company.activateUser) — CompanyController@activateUser
- **Auth/role required:** auth, role:company
- **Business logic steps:** `$user->status='active'; $user->save();`
- **Response:** redirect to `company.dashboard`. Same caveat as Admin's version — no package/wallet side effects.

### GET /company/pending-activation (company.pending.activation) — CompanyController@pendingActivation
- **Auth/role required:** auth, role:company
- **Business logic steps:**
  1. Declares (but does not use in the view) a hardcoded `$packageFees` map keyed by package price → fee % (10→20, 100→40, 1000→60, 5000→70, 10000→80, 100000→90, 1000000→90, 5000000→90, 10000000→90). Dead code — computed and discarded.
  2. `$myPackage` computed twice, second overwrites first (`UserPackage::where('user_id',auth()->id())->first()` — no status filter on the second query, overwriting the first `status='active'`-filtered lookup). `$myPackage` is never passed to the view anyway.
  3. `$activations = UserPackage::with('userpackage')->where('status','pending')->paginate(10)` — **note: unlike Admin/Agent this is NOT filtered by `ref_id`; shows ALL pending packages company-wide.**
- **Response:** view `company.pending-activation` compact `activations` only (`feePercentage` referenced in compact() call but never defined in this method — will raise an undefined-variable warning/null in the view under Laravel).

### GET /new-activations (new.activations) — CompanyController@newActivations
- **Auth/role required:** auth, role:company
- **Business logic steps:** `UserPackage::with('userpackage')->where('company_status', 0)->paginate(10)`.
- **Response:** view `company.new-activations` with `activations`.

### POST /company/new_active-package (company.newActivepackage) — CompanyController@newActivepackage
- **Auth/role required:** auth, role:company
- **Request input:** `package_id`: `required|integer|exists:user_packages,id`
- **Business logic steps:**
  1. `UserPackage::with('userpackage')->where('company_status', 0)->findOrFail($packageId)`.
  2. Set `company_status = 1`, save. (This just flips a "seen/approved by company" flag; it does NOT set `status = 'active'` — separate from the `active.package` wallet/commission flow.)
- **Response (JSON, wrapped in try/catch):**
  - Success 200: `{success:true, message:'Package activated successfully!'}`
  - `ModelNotFoundException` → 404: `{success:false, message:'Package not found or already activated.'}`
  - `ValidationException` → 422: `{success:false, message:'Invalid package ID.', errors:{...}}`
  - Any other `\Exception` → 500: `{success:false, message:'Something went wrong. Please try again.'}`

---

## CountriesController

### GET /countries (countries.index) — CountriesController@index
- **Auth/role required:** auth, role:company
- **Business logic:** `Countries::latest()->paginate(10)`.
- **Response:** view `countries.index` with `countries`.

### POST /countries (countries.store) — CountriesController@store
- **Auth/role required:** auth, role:company
- **Request input:** `code`: `required|string|max:10|unique:countries,code`; `name`: `required|string|max:255`
- **Business logic:** `Countries::create($validated)`.
- **Response:** `Alert::success('Success','Country Create Success')`, redirect to `countries.index`.

### PUT/PATCH /countries/{country} (countries.update) — CountriesController@update
- **Auth/role required:** auth, role:company
- **Request input:** `code`: `required|string|max:10|unique:countries,code,{id}` (excludes current row); `name`: `required|string|max:255`
- **Business logic:** `$country->update($validated)`.
- **Response:** success alert, redirect to `countries.index`.

### DELETE /countries/{country} (countries.destroy) — CountriesController@destroy
- **Auth/role required:** auth, role:company
- **Business logic:** `$country->delete()`.
- **Response:** success alert, redirect to `countries.index`.

(Route::resource registers full CRUD, but `create`/`edit`/`show` methods are not implemented on this controller — those verbs will 404/error if hit.)

---

## DirectShareController

### GET /direct-share (company.direct_share) — DirectShareController@directShare
- **Auth/role required:** auth, role:company
- **Request input:** optional `start_date`, `end_date` (query)
- **Business logic steps:**
  1. `PackagePool::with(['user','package'])`, optionally filtered `whereDate('created_at', >=/<=)`.
  2. `$companyPool` = sum of `pool_amount` where `user_id = 1` (company sentinel user).
  3. `$salesPool` = sum of `pool_amount` where `user_id != 1`.
  4. `$totalPool = companyPool + salesPool`.
  5. `$pools` = paginated (20) list, `withQueryString()`.
- **Response:** view `company.direct_share` with `companyPool,salesPool,totalPool,startDate,endDate,pools`.

### GET /direct-share-log (company.direct_share_log) — DirectShareController@directShareLog
- **Auth/role required:** auth, role:company
- **Request input:** optional `start_date`/`end_date`, defaulting to **previous** calendar month's start/end (`subMonthNoOverflow()->startOfMonth()/endOfMonth()`).
- **Business logic steps:** `GlobalShareWalletLog::with('user')` filtered by date range, paginate 20.
- **Response:** view `company.direct_share_log` with `pools,startDate,endDate`.

### POST /package-pools (package-pools.store) — DirectShareController@store
- **Auth/role required:** auth, role:company
- **Request input:** `user_id`: `required|exists:users,id`; `pool_amount`: `required|numeric|min:0`
- **Business logic:** `PackagePool::create(['user_id'=>1,'package_id'=>1,'pool_amount'=>$request->pool_amount])` — **note: ignores the submitted `user_id` and always hardcodes `user_id=1, package_id=1`** despite validating `user_id` against the users table. Flag as likely bug / intentional company-only pool add.
- **Response:** redirect back with flash `success` = "Pool added successfully."

### PUT/PATCH /package-pools/{package_pool} (package-pools.update) — DirectShareController@update
- **Auth/role required:** auth, role:company
- **Request input:** `pool_amount`: `required|numeric|min:0`
- **Business logic:** `$packagePool->update(['pool_amount' => $request->pool_amount])`.
- **Response:** redirect back, flash `success` = "Pool Values updated successfully."

### DELETE /package-pools/{package_pool} (package-pools.destroy) — DirectShareController@destroy
- **Business logic:** `$packagePool->delete()`.
- **Response:** redirect back, flash `success` = "Pool Values deleted successfully."

(`index`/`create`/`show`/`edit` not implemented on this controller — those resource verbs 404.)

---

## EarnLogController

### GET /earn/history (earn.history) — EarnLogController@index
- **Auth/role required:** none declared on the route (relies on `auth()->user()->id` — see cross-cutting note)
- **Request input:** optional `date_from`, `date_to` (query, checked via `filled()`)
- **Business logic steps:**
  1. `EarnLog::where('user_id', auth()->id())->orderBy('created_at','desc')`, optional date-range filters.
  2. `$totalAmount` = sum of `amount` on the (cloned, unpaginated) filtered query.
  3. `$earns` = paginate(10), `withQueryString()`.
- **Response:** view `earn.index` with `earns,totalAmount`.

---

## ExecutiveController

### GET /executives/gain (executives.gain) — ExecutiveController@index
- **Auth/role required:** auth, role:company
- **Request input:** optional `from`,`to` (default: current month start/end), optional `executive_id`
- **Business logic steps:**
  1. `$executives` = Users whose `id` appears as some other user's `executive_code` (distinct, `executive_code IS NOT NULL AND > 0`), ordered by name.
  2. Base query same set; if `executive_id` filled, narrow to that one user.
  3. `selectSub`: per executive, `total_package` = `COALESCE(SUM(packages.price),0)` from `user_packages` joined to `users AS members` (member's `executive_code = users.id`) and `packages`, where `user_packages.status IN ('active','deactivate')` **(note: 'deactivate', not 'deactive' — check against actual status enum used elsewhere, which is 'deactive' in Token logic; possible typo/inconsistency to verify against the DB schema)**, `sale='first'`, `created_at` between `from 00:00:00` and `to 23:59:59`.
  4. `withCount('executiveMembers as total_members')`.
  5. Paginate 10, order by name.
- **Response:** view `executives.index` with `executives,users,from,to`.

### GET /leadership-bonus-log (company.leadership_bonus_log) — ExecutiveController@leadershipBonusLog
- **Auth/role required:** auth, role:company
- **Request input:** optional `start_date`/`end_date`, default previous month.
- **Business logic:** `EarnLog::where('description','Leadership Bonus')->with('user')`, date-filtered, paginate 20. (Matches the `'Leadership Bonus'` description string used in `TokenController::processPackageActivation` when crediting leader/executive code bonuses.)
- **Response:** view `company.leadership_bonus_log` with `pools,startDate,endDate`.

---

## GeneologyController

### GET /my-geneology (my.geneology) — GeneologyController@index
- **Auth/role required:** none declared (relies on `Auth()->user()`)
- **Business logic:** `UserParent::where('parent_id', auth()->id())->with('user')->whereIn('node',['active','gratitude'])->where('user_id','!=',auth()->id())->get()`.
- **Response:** view `geneology.index` with `childerns` (sic).

### GET /geneology/{userId} (geneology.show) — GeneologyController@viewGeneology
- **Auth/role required:** none declared
- **Business logic:** loads `User::find($userId)`, `UserPackage::where('user_id',$userId)->with('userpackage')->get()`, `User::where('id', $userdata->referred_by)->first()` as `$parentData`.
- **Response:** view `geneology.show` with `userdata,userPackage,parentData`.

---

## GoogleAuthenticatorController

### GET /admin/{userId}/setup-google-auth (setup.google.auth) — GoogleAuthenticatorController@setupGoogleAuthenticator
- **Auth/role required:** auth, role:company (registered only in the company group despite the `/admin/...` URL prefix)
- **Request input:** route param `userId`
- **Business logic steps:**
  1. `User::find($userId)`; if not found, redirect back with error "User not found".
  2. `new GoogleAuthenticator()`; `$secret = $gAuth->generateSecret()`.
  3. Save `$secret` to `$user->google_authenticator_secret`.
  4. Build `otpauth://totp/...` URL with app name `signetint`, the user's email, and the secret.
  5. Render QR via BaconQrCode SVG renderer (300px), base64-encode, wrap as `data:image/svg+xml;base64,...`.
- **Response:** view `admin.setup_google_authenticator` with `qrCodeImageBase64,secret`.
- **Note:** this endpoint regenerates (overwrites) the secret every time it's visited — visiting twice invalidates a previously configured authenticator app. It sets the secret on the *target* `$userId`'s record, but `TokenController::generateTokens` checks the *authenticated* user's (`auth()->user()->google_authenticator_secret`) — i.e. a company user must run this against **their own** id to set up 2FA for token generation.

---

## KycController

### GET /kyc (kyc.index) — KycController@index
- **Auth/role required:** none declared (uses `auth()->check()` defensively)
- **Business logic steps:**
  - If authenticated and role !== 'company': fetch the current user's single `Kyc` row. Alert `info` if none submitted, `success` if `is_verified`, `warning` if pending. Pass `kycs` as a 0-or-1-element collection.
  - Else (guest OR role === 'company'): `Kyc::where('is_verified',0)->whereHas('user', fn($q)=>$q->where('role','!=','company'))->paginate(10)`. Alert info/success based on emptiness.
- **Response:** view `kyc.index` with `kycs` (shape differs by branch — single-collection vs. paginator).

### GET /kyc/verified (kyc.verified) — KycController@verified
- **Auth/role required:** auth, role:company (route-gated, unlike sibling `kyc.index`)
- **Business logic:** `Kyc::where('is_verified',1)->whereHas('user', fn($q)=>$q->where('role','!=','company'))->paginate(10)`. Alerts as above.
- **Response:** view `kyc.verified` with `kycs`.

### GET /kyc/create (kyc.create) — KycController@create
- **Auth/role required:** none declared
- **Response:** view `kyc.create`.

### GET /kyc/{id}/edit (kyc.edit) — KycController@edit
- **Auth/role required:** none declared
- **Business logic:** `Kyc::where('user_id', auth()->id())->findOrFail($id)`.
- **Response:** view `kyc.edit` with `kyc`.

### POST /kyc (kyc.store) — KycController@store
- **Auth/role required:** none declared
- **Request input:**
  - `full_name`: `required|string`
  - `email`: `required|email`
  - `contact_number1`: `required`
  - `contact_number2`: `nullable`
  - `document_type`: `required|in:nic,passport`
  - `document_number`: `required|unique:kycs,document_number`
  - `nic_front`: `required_if:document_type,nic|image`
  - `nic_back`: `required_if:document_type,nic|image`
  - `passport_image`: `required_if:document_type,passport|image`
- **Business logic:**
  1. `new Kyc($request->except(['nic_front','nic_back','passport_image']))`, set `user_id = auth()->id()`.
  2. Store each uploaded file (if present) to `public` disk under `kyc/nic_front`, `kyc/nic_back`, `kyc/passport` respectively; assign returned paths.
  3. Save.
- **Response:** `Alert::success('Success','KYC submitted. Pending verification.')`, redirect to `kyc.index`.

### PUT /kyc/{id} (kyc.update) — KycController@update
- **Auth/role required:** none declared
- **Request input:** same fields as store, but `document_number` uses `Rule::unique('kycs','document_number')->ignore($kyc->id)`, and the three file fields are `nullable|image` (optional on update).
- **Business logic:** loads `Kyc::where('user_id',auth()->id())->findOrFail($id)`; `$kyc->update(...)` on non-file fields; re-store any newly uploaded files (old files NOT deleted from disk — flag as potential storage leak to replicate-or-fix consciously); save.
- **Response:** success alert "KYC updated", redirect to `kyc.index`.

### DELETE /kyc/{id} (kyc.destroy) — KycController@destroy
- **Business logic:** `Kyc::where('user_id',auth()->id())->findOrFail($id)->delete()`.
- **Response:** success alert "KYC deleted.", redirect to `kyc.index`.

### POST /kyc/{id}/verify (kyc.verify) — KycController@verify
- **Auth/role required:** none declared at the route (!) — any caller, including guests, can hit this endpoint since `Kyc::findOrFail($id)` performs no ownership/role check.
- **Business logic:** `$kyc->is_verified = true; $kyc->save();`
- **Response:** success alert "KYC verified successfully.", redirect back with flash `success`.

### POST /kyc/{id}/unverify (kyc.unverify) — KycController@unverify
- **Auth/role required:** none declared (same gap as verify)
- **Business logic:** `$kyc->is_verified = false; $kyc->save();`
- **Response:** warning alert "KYC unverified successfully.", redirect back with flash `success`.

### GET /kyc/show (kyc.show) — KycController@show
- **Auth/role required:** none declared
- **Business logic:**
  - If `auth()->user()->role !== 'company'`: fetch current user's `Kyc` (same alert logic as index), plus `auth()->user()->packages()->where('status','active')->with('userpackage')->get()` → `$userPackages` **but note: `$userPackages` is passed via a 3rd positional `compact()` argument to `view()`, which is invalid Laravel usage (`view($name, $data, $mergeData)` expects an array, not `compact(...)` called separately) — this will not actually inject `userPackages` into the view correctly; likely a bug.** Also **this method calls `auth()->user()->role` directly with no null-check, unlike `index()`/`store()` which guard with `auth()->check()` — a guest hitting `/kyc/show` will get a fatal error, not a redirect.**
  - Else (company): same paginated non-verified-KYC list as `index()`'s company branch.
- **Response:** view `kyc.show` with `kycs` (and, buggy, attempted `userPackages`).

---

## LeaderController

### GET /leaders/gain (leaders.gain) — LeaderController@index
- **Auth/role required:** auth, role:company
- **Request input:** optional `from`,`to` (default current month), optional `leader_id`
- **Business logic:** Mirrors `ExecutiveController@index` exactly but keyed on `leader_code`/`leader_status` instead of `executive_code`: `$leaders` = users referenced as someone's `leader_code`; `$users` = same set (optionally filtered to one `leader_id`) with subselect `total_package` = sum of first-sale active/deactivate package prices for that leader's direct members within `[from,to]`, and `withCount('downline as total_members')`.
- **Response:** view `leaders.index` with `leaders,users,from,to`.

---

## MiningController

### GET /mining/users (mining.users) — MiningController@index
- **Auth/role required:** auth, role:company
- **Response:** view `company.mining` (no data passed — page is presumably AJAX-driven via `search`/`update` below).

### GET /mining/search/{id} — MiningController@search
- **Auth/role required:** auth, role:company
- **Request input:** route param `id`, accepts either raw numeric id or `SIG-00{n}` formatted id (regex `^SIG-0*(\d+)$` strips prefix/leading zeros).
- **Business logic:**
  1. `User::with(['packages.userpackage' => orderBy created_at asc, 'mining'])->where('id',$id)->first()`.
  2. `$userparent` = count of `UserParent` rows with `virtual_id = $id` and `node IN ('active','gratitude')`.
  3. If user not found: `{success:false, message:'User not found'}` (200, implicit).
  4. `$firstPackage`/`$lastPackage` = first/last package's `userpackage->name`.
- **Response:** `{success:true, user:{id:'SIG-00'.$id, name, email}, packages:{first,last}, mining:{total_token,mining_token,daily_mining,status(default 'inactive')}, sales:{total_sales: $userparent}}`.

### POST /mining/update/{id} — MiningController@update
- **Auth/role required:** auth, role:company
- **Request input:** `daily_mining`, `total_token`, `status` — **read via `$request->` directly, no `validate()` call at all**; any type/absence is accepted as-is (including null).
- **Business logic:** id normalized from `SIG-00N` if needed; `User::find($id)`; if missing → `{success:false, message:'User not found'}`. Else `$mining = $user->mining ?? new UserMining()`; set `user_id`, `daily_mining`, `total_token`, `status` from request; save.
- **Response:** `{success:true}`.

### GET /mining/user/{userId} — MiningController@getUserMiningData
- **Auth/role required:** none declared on the route
- **Business logic:** raw query builder join `user_minings AS m` to `users AS u` on `m.user_id=u.id`, filtered `m.user_id = $userId`, select mining columns; if none found → 404 `{success:false, error:'Mining data not found'}`.
- **Response (200):** `{success:true, data:{mining_token: round(x,8), total_token, daily_mining, status, progress: (mining_token/total_token)*100, updated_at}}`. On any exception: 500 `{success:false, error: <exception message>}`. **Division by zero risk** if `total_token` is 0 — PHP would emit a warning/INF, not throw; not caught by the try/catch as an exception.

---

## PackageController

### GET /packages (packages.index) — PackageController@index
- **Auth/role required:** auth, role:company
- **Business logic:** `Package::latest()->get()`.
- **Response:** view `packages.index` with `packages`.

### GET /packages/create (packages.create) — PackageController@create
- **Response:** view `packages.create`.

### POST /packages (packages.store) — PackageController@store
- **Request input:**
  - `name`: `required`
  - `price`: `required|numeric`
  - `commission`: `required|numeric`
  - `rank`: `required`
  - `telegram_link`: `required`
  - `status`: `required|in:active,deactive`
- **Business logic:** `Package::create($request->all())` — **mass-assigns the entire request body**, not just validated fields (relies on `Package` model's `$fillable`/`$guarded` to prevent injection of unexpected columns — verify the model, not deep-dived here per scope).
- **Response:** redirect to `packages.index`, flash `success` "Package created successfully."

### GET /packages/{package}/edit (packages.edit) — PackageController@edit
- **Response:** view `packages.edit` with `package`.

### PUT/PATCH /packages/{package} (packages.update) — PackageController@update
- **Request input:** identical rules to store.
- **Business logic:** `$package->update($request->all())`.
- **Response:** redirect to `packages.index`, flash "Package updated successfully."

### DELETE /packages/{package} (packages.destroy) — PackageController@destroy
- **Business logic:** `$package->delete()`.
- **Response:** redirect to `packages.index`, flash "Package deleted successfully."

### GET /buy-package-history (buy.package.history) — PackageController@buyPackageHistory
- **Auth/role required:** none declared (relies on `auth()->user()`)
- **Business logic:** `UserPackage::where('user_id',auth()->id())->with('user')->paginate(10)` → `$packages`; `$activePackage` = current user's active `UserPackage` (first).
- **Response:** view `packages.buy-packageHistory` with `packages,activePackage`.

### GET /buy-package (buy.package) — PackageController@buyPackage
- **Auth/role required:** none declared
- **Business logic:** `Package::where('status','active')->get()`.
- **Response:** view `packages.buy-package` with `package`.

### POST /buy-packages (buy.packages) — PackageController@buyPackages
- **Auth/role required:** none declared
- **Request input:** `package` (id) — **not validated at all** (no `$request->validate()` call; `Package::find($request->package)` will just be null on bad input and is handled).
- **Business logic steps (additional-package-purchase flow, distinct from registration's first package):**
  1. `$loggedUserId = auth()->id()`.
  2. Walk up the `UserParent` chain (max 50 hops) from the logged-in user until an ancestor `User` with `status == 'active'` is found → `$activeParentId`.
  3. `$checkActivation = $activeParentId ? checkWalet($activeParentId, $request->package) : 0` — **flagged for math agent** (wallet-capacity/eligibility check).
  4. `$saveUser = $loggedUserId` (the purchasing user, always).
  5. If `$checkActivation == 1` and an active parent was found: `$user = $activeParentId` (referrer for commission purposes); else `$user = 1` (company fallback).
  6. Load `$parentData = User::find($user)`; if missing → error "User not found", redirect back.
  7. Load `$UserBuyPackage = Package::find($request->package)`; if missing → error "Package not found", redirect back.
  8. `$ref_id = ($checkActivation == 1) ? $user : 0`.
  9. Create new `UserPackage`: `user_id = $saveUser`, `package = $UserBuyPackage->id`, `status = 'pending'`, `ref_id = $ref_id`, `sale = 'other'` (distinguishes from the registration-time `sale='first'`).
- **Response:** view `packages.buy-package-done` with `parentData` (no redirect — direct render).

---

## RocController

### GET /roc (company.roc) — RocController@rocIncome
- **Auth/role required:** auth, role:company
- **Request input:** optional `job_id` (query)
- **Business logic:**
  1. `$jobs = WeeklyPackageSummary::orderBy('id','desc')->get()`.
  2. `$selectedJobId = $request->job_id ?? $jobs->first()?->job_id`.
  3. If none: alert (mislabeled as success, message text "No ROC job found.") + redirect back with `error` flash "No ROC job found."
  4. `$weeklySummary = WeeklyPackageSummary::where('job_id',$selectedJobId)->first()`.
  5. `$rocIncomeLogs = RocIncomeLog::where('job_id',$selectedJobId)->with('user')->orderBy('created_at','desc')->paginate(10)`.
- **Response:** view `company.roc_income` with `rocIncomeLogs,jobs,selectedJobId,weeklySummary`.

### POST /company/roc/status-update (company.updateRocStatus) — RocController@updateRocStatus
- **Auth/role required:** auth, role:company
- **Request input:** `id`: `required|integer`; `status`: `required|in:pending,paid`
- **Business logic steps:**
  1. `RocIncomeLog::with('earnLog')->find($request->id)`; if not found → `{success:false, message:'Record not found.'}` (200 implicit).
  2. `$shouldCreditWallet = ($request->status === 'paid') && ($roc->status !== 'paid')` — computed **before** mutating `$roc->status` (correct idempotency guard: won't double-credit if already paid).
  3. Set `$roc->status = $request->status`, save.
  4. If related `earnLog` exists, update its `description` to `'ROC Income paid'` if now paid, else `'ROC Income'`.
  5. If `$shouldCreditWallet`: call `$this->walletService->updateWallet($roc->user_id, $roc->amount, $roc->description)` — **flagged for math agent** (this is `WalletService::updateWallet`, which itself gates crediting by wallet-cap `totalValue < wallet->balance`, then credits Wallet, EarnLog, GlobalShareWallet up to `max_out`, and rolls overflow to company wallet (`user_id=1`) — see `app/Services/WalletService.php`).
- **Response:** `{success:true, message:'Status updated successfully.'}`.

---

## SalaryController

Constructor-injects `WalletService`.

### GET /salaries (salaries.index) — SalaryController@index
- **Auth/role required:** auth, role:company
- **Request input:** optional `date_from`,`date_to` (`filled()`-checked, filters on `salary_date`)
- **Business logic:** `Salary::with('user')->orderByDesc('salary_date')`, filtered; `$totalAmount` = sum of `amount` on filtered (unpaginated clone); `$salaries` = paginate 15.
- **Response:** view `salaries.index` with `salaries,totalAmount`.

### GET /salaries/search-users (salaries.searchUsers) — SalaryController@searchUsers
- **Auth/role required:** auth, role:company
- **Request input:** `term` (query, default `''`, trimmed)
- **Business logic:** Select2-style AJAX search: `User::query()`, when `term` non-empty, matches `name LIKE %term%` OR `whatsapp_number LIKE %term%`, plus if the term ends in digits, also `OR id = <numeric part with leading zeros stripped>` (handles "SIG-00N" search input). Limit 15, select `id,name,whatsapp_number`.
- **Response (JSON array):** `[{id, text: "{name} (SIG-00{id})[ - {whatsapp_number}]"}]`.

### POST /salaries (salaries.store) — SalaryController@store
- **Auth/role required:** auth, role:company
- **Request input:**
  - `user_id`: `required|exists:users,id`
  - `amount`: `required|numeric|min:0`
  - `salary_date`: `required|date`
  - `remarks`: `nullable|string|max:255`
- **Business logic:** inside `DB::transaction`: `Salary::create($validated)`, then `$this->walletService->updateWallet($salary->user_id, $salary->amount, 'Salary credited')` — **flagged for math agent** (same `WalletService::updateWallet` gate/credit logic as ROC).
- **Response:** success 200 `{success:true, message:'Salary added successfully'}`; on any `\Throwable`, 500 `{success:false, message:'Could not add salary: '.$e->getMessage()}` (leaks internal exception text to the client — flag for the Go port to NOT do this, or to keep parity deliberately).

### DELETE /salaries/{id} (salaries.destroy) — SalaryController@destroy
- **Business logic:** `Salary::find($id)`; if missing → 404 `{success:false, message:'Record not found'}`; else delete.
- **Response:** `{success:true}` (200 implicit).

---

## UserParentMapsLogController

### GET user-parent-logs (userparentlogs.index) — UserParentMapsLogController@index
- **Auth/role required:** auth, role:company
- **Business logic:** `UserParentMapsLog::with('user')->whereHas('user', fn($q)=>$q->where('status','pending'))->where('created_at','<=', now()->subHours(10))->orderBy('created_at','desc')->paginate(10)`. (Comment says "deactive users only" but the filter is actually `status = 'pending'`.)
- **Response:** view `company.user_parent_logs.index` with `logs`.

### DELETE user-parent-logs/{log} (userparentlogs.destroy) — UserParentMapsLogController@destroy
- **Auth/role required:** auth, role:company
- **Business logic steps (destructive — inside `DB::transaction`, wrapped in try/catch):**
  1. `UserParentMapsLog::with('user')->lockForUpdate()->findOrFail($id)`.
  2. Delete the user's `ReferralCode`.
  3. Delete all `UserParent` rows whose `id` is in `$log->created_row_ids` (JSON array logged at tree-placement time).
  4. Delete the log row itself.
  5. Delete the `User` row (`User::where('id',$log->user_id)->delete()`) — **hard-deletes the user account**, presumably to let a stuck/never-activated pending registration be purged and re-registered.
  - On any `\Throwable`: logged via `Log::error` with `log_id`, error message, and trace; transaction auto-rolls back.
- **Response:** redirect back with flash `success` "Related UserParent records, referral code, user and log were deleted successfully." or flash `error` "Something went wrong while deleting records. No changes were saved."

---

## WalletController

Empty stub — `App\Http\Controllers\WalletController` has no methods and is not referenced by any route. Not wired anywhere; no endpoint to port.

---

## UserController

### GET /user/dashboard (user.dashboard) — UserController@index
- **Auth/role required:** auth, role:user
- **Business logic:** same dashboard-aggregation shape as Admin/Agent (`refLink`, `myTokens`, `myWallet`/`totalValue`, `myPackage`/`feePercentage` computed twice — second overwrite adds `.latest()`, `activations` filtered by `ref_id=auth()->id()` and `status='pending'`, `poolAmount` current month, `totalPoolshareValue`, `myshareValue` (same formula, flagged), `myGlobleDirectorShare = GlobalShareWallet::where('user_id',auth()->id())->first()`).
- **Response:** view `user.dashboard` with all compacted.

### POST /toggle-vacation (user.toggleVacation) — UserController@toggleVacation
- **Auth/role required:** none declared (relies on `auth()->user()`)
- **Business logic:** `$user->on_vacation = !$user->on_vacation; $user->save();`.
- **Response:** `{status:true, on_vacation: <new bool>}`.

### GET /users (company.users) — UserController@allUsers
- **Auth/role required:** auth, role:company OR admin
- **Request input:** optional `search` (query)
- **Business logic:** `User::query()`, if `search` present: OR-matches `name`, `whatsapp_number`, `status` (all `LIKE %term%`), `whereHas('country', fn($q)=>$q->where('code','like',...))`, and `CONCAT('SIG-00',id) LIKE %term%` (raw SQL). Then `whereIn('status', ['active','inactive','pending'])`, eager-load `leader`,`executive`,`country`, paginate 10, `appends(['search'=>...])`. Also `$leaders = User::where('leader_status','active')->orderBy('name')->get()`.
- **Response:** view `company.users` with `allUsers,leaders`.

### GET /users/search/{id} — UserController@search
- **Auth/role required:** auth, role:company OR admin
- **Business logic:** id normalized from `SIG-00N`; `User::whereIn('status',['active','inactive'])->with(['packages.userpackage' (ordered asc), 'mining','wallet'])->where('id',$id)->first()`. `$userparent` = count of `UserParent` rows with `virtual_id=$id`. `$directSale` = count of `UserParent` rows with `parent_id=$id`, `node IN ('active','gratitude')`, `user_id != $id`. `$totalSales` = sum of joined `packages.price` for that user's `UserPackage`s with `status IN ('active','deactive')`. If user not found → `{success:false, message:'User not found'}`. `$wallet = $totalSales*4 - ($user->wallet?->balance ?? 0)` — **note operator-precedence: PHP evaluates `?->` before `??`, so this is `$totalSales*4 - ($user->wallet?->balance) ?? 0`, i.e. if `wallet` exists this is fine, but if `$user->wallet` is null the whole subtraction becomes `null`, and only THEN does `?? 0` kick in — meaning `$totalSales*4` is discarded and `$wallet` becomes exactly `0` when there's no wallet row, not `$totalSales*4 - 0`.** Flag for math agent / Go port to replicate this precedence quirk exactly or knowingly fix it.
- **Response:** `{success:true, user:{id,name,email,status,roc_status,global_director_share,global_director_share_status}, packages:{first,last}, mining:{total_token,mining_token,daily_mining}, sales:{total_sales,direct_sales}, wallet:{total_wallet: (int)($wallet ?? 0)}}`.

### POST /users/update/{id} — UserController@update
- **Auth/role required:** auth, role:company
- **Request input:** `status`: `required|in:active,inactive`
- **Business logic:** id normalized; `User::find($id)`; 404-style JSON if missing; else `$user->status = $request->status; $user->save();`.
- **Response:** `{success:true}` or `{success:false, message:'User not found'}`.

### POST /users/update-roc/{id} — UserController@updateRocStatus
- **Auth/role required:** auth, role:company
- **Request input:** `status`: `required|in:active,inactive`
- **Business logic:** same pattern, sets `$user->roc_status`.
- **Response:** same shape as above.

### POST /users/update-leader-status/{id} — UserController@updateLeaderStatus
- **Auth/role required:** auth, role:company
- **Request input:** `status`: `required|in:active,inactive`
- **Business logic:** sets `$user->leader_status`.
- **Response:** same shape.

### POST /users/update-leader-code/{id} — UserController@updateLeaderCode
- **Auth/role required:** auth, role:company OR admin
- **Request input:** `leader_code`: `nullable|exists:users,id`
- **Business logic:**
  1. Reject if `leader_code == $user->id` (self-leader) → `{success:false, message:'A user cannot be their own leader'}`.
  2. `$newLeaderCode = $request->leader_code ?: null`; `$oldLeaderCode = $user->leader_code ?: null`. If unchanged, short-circuit `{success:true}` with no log entry.
  3. Else: create `LeaderCodeLog` (`user_id`,`old_leader_code`,`new_leader_code`,`changed_by`=auth id), then set `$user->leader_code = $newLeaderCode`, save.
- **Response:** `{success:true, leader: <new leader's name or null>}`.

### POST /users/update-executive-code/{id} — UserController@updateExecutiveCode
- **Auth/role required:** auth, role:company OR admin
- **Request input:** `executive_code`: `nullable|exists:users,id`
- **Business logic:** mirrors `updateLeaderCode` exactly, logging to `ExecutiveCodeLog`.
- **Response:** `{success:true, executive: <new executive's name or null>}`.

### GET /leader-code-logs (leader.code.logs) — UserController@leaderCodeLogs
- **Auth/role required:** auth, role:company OR admin
- **Request input:** optional `search`,`from`,`to`
- **Business logic:** `LeaderCodeLog::with(['user','oldLeader','newLeader','changedBy'])`, `search` matches user name or `SIG-00{id}`; date-range filters on `created_at`; paginate 15, `appends($request->query())`.
- **Response:** view `leader_code_logs.index` with `logs`.

### GET /executive-code-logs (executive.code.logs) — UserController@executiveCodeLogs
- Mirrors `leaderCodeLogs`, view `executive_code_logs.index`.

### POST /users/update-global-director-share/{id} — UserController@updateGlobalDirectorShare
- **Auth/role required:** auth, role:company
- **Request input:** `value`: `required|numeric|min:0`; `status`: `required|in:1,0`
- **Business logic:** sets `$user->global_director_share = $request->value`, `$user->global_director_share_status = (bool)$request->status`, save.
- **Response:** `{success:true}` or not-found JSON.

---

## TokenController

Core of the wallet/commission activation flow. **All financial multipliers/percentages/formulas below are flagged for the dedicated math-extraction agent** — only the call sequence, arguments, and gating conditions are documented here.

Large block of **dead/commented-out legacy code** (original `activePackage` implementation, roughly 300 lines) precedes the live implementation in the source file — not analyzed line-by-line since it is inert, but noted because it shows the pre-refactor version of the same logic (useful cross-reference if the live version's behavior looks incomplete vs. intent).

### POST /generate-tokens/{userId} (generate.tokens) — TokenController@generateTokens
- **Auth/role required:** auth, role:company. **Plus 2FA:** requires a valid Google Authenticator TOTP code from the *authenticated* user (see below) — this is the only endpoint in the app enforcing 2FA at the controller level.
- **Request input:**
  - `token_count`: `required|integer|min:1|max:500`
  - `google_auth_code`: `required|numeric`
- **Business logic steps:**
  1. `User::find($userId)` (the target user tokens are generated FOR); if missing, redirect back with error "User not found".
  2. Validate `token_count` (separate `validate()` call).
  3. Validate `google_auth_code` (separate `validate()` call).
  4. `$secret = auth()->user()->google_authenticator_secret` — **note: reads the secret of the currently logged-in company user, not of `$userId`.** If empty, redirect back with error "Google Authenticator secret not set for the user".
  5. `new GoogleAuthenticator(); $gAuth->checkCode($secret, $request->google_auth_code)`. If invalid, redirect back with error "Invalid Google Authenticator code".
  6. Loop `token_count` times, building `{user_id: $userId, token: bin2hex(random_bytes(32)), status:'active', created_at, updated_at}`.
  7. `Token::insert($tokens)` (bulk insert, bypasses model events/timestamps auto-fill since it's raw `insert()`, timestamps set manually above).
- **Response:** redirect to `view.tokens` (with `$userId`) and flash `success` "{$count} tokens generated successfully"; or various redirect-back error flashes.

### GET /view-tokens/{userId} (view.tokens) — TokenController@viewTokens
- **Auth/role required:** auth, role:company
- **Business logic:** `User::find($userId)`; if missing, error+back. `Token::where('user_id',$userId)->paginate(10)`.
- **Response:** view `tokens.index` with `user,tokens`.

### POST /active-package (active.package) — TokenController@activePackage
- **Auth/role required:** none declared on the route (relies on `auth()->user()`) — this is a highly sensitive financial endpoint with no route-level auth guard; flag prominently for the Go port (must require session auth, presumably company/admin-only given it's driven from admin/company package-activation UIs, though as written any authenticated user hitting it would run the logic against `auth()->user()->id`).
- **Request input:** `package_id` (read via `$request->input('package_id')`, **not validated** with `validate()`).
- **Business logic steps (wrapped in `DB::transaction`):**
  1. `UserPackage::where('id',$packageId)->with('userpackage')->lockForUpdate()->first()` → `$package` (row-level lock guards against concurrent double-activation of the same package).
  2. If found, also load `$packageData = Package::where('id',$package->package)->first()`.
  3. If `$package` not found → `{message:'Package not found.'}` (404).
  4. Compute `$totalValue` = sum(`price*4`) over the *authenticated* user's own active `UserPackage`s (used only in the non-company branch below).
  5. **Branch A — `auth()->user()->id == 1`** (company/system sentinel account): call `processPackageActivationCompany($package)` directly, skip all wallet/token checks.
  6. **Branch B — everyone else:**
     a. Load `$wallet = Wallet::where('user_id',auth()->id())->first()`.
     b. If `$totalValue < $wallet->balance` → `{message:'Please top up your wallet.'}` (400). **Note: NPE risk if `$wallet` is null** (no wallet row) — `$wallet->balance` on null throws in PHP 8, not caught here.
     c. Else: recompute the authenticated user's highest active package + its `commission` → `$feePercentage`.
     d. `$needTokens = $package->userpackage->price - ($package->userpackage->price * ($feePercentage ?? 0)/100)` — **flag for math agent.**
     e. `$tokensCount = Token::where('user_id',auth()->id())->where('status','active')->count()`.
     f. If `$tokensCount < $needTokens`: if `auth()->id() !== 1`, return `{message:'Not enough tokens.'}` (400) (note: this branch is unreachable for id==1 since that's handled in Branch A already — dead condition).
     g. Else: `$checkActivation = checkWalet(auth()->id(), $packageData->id)` — **flagged for math agent.** If `== 1`: call `processPackageActivation($package, $needTokens, $feePercentage)` then `deactivateTokens($needTokens)`. Else: `{message:'Not enough wallet.'}` (400).
- **Response:** `{message:'Package and wallet updated successfully.'}` (200, default status) on the success path reached at the end of the transaction closure — **note this success message is returned unconditionally after the transaction closure runs, even along code paths that already returned an error response earlier inside the closure; those early `return response()->json(...)` calls DO exit the closure correctly in PHP, so this is fine, just worth verifying in tests.**

#### private processPackageActivation($package, $needTokens, $feePercentage)
Not directly routed; called from `activePackage` Branch B. Sequence:
1. `UserParent::where('user_id',$package->user_id)->first()` → `$findUser`; 404 JSON if missing (return value discarded by caller since this is inside a closure calling a non-static private method — **flag: this `return response()->json(...)` inside a private helper does NOT actually short-circuit `activePackage`'s outer transaction/response; it just returns a value that's ignored, so execution silently continues to the "success" response even when the UserParent relation is missing.** Same footgun applies to the two `return response()->json(...)` calls inside `processPackageActivationCompany`.)
2. Set `$findUser->node` = `'gratitude'` if it was already `'gratitude'`, else `'active'`; save.
3. Set `$package->status='active'`, `$package->activated_at = Carbon::now('Asia/Colombo')`, save.
4. Set the package owner's `User.status = 'active'`.
5. If `$package->sale == 'other'` (i.e., a subsequent/extra package purchase, not first registration): find the user's prior active `UserPackage`, look up a `SuperParentLog` by `user_package = $oldPackage->id`; if found, call `updateWallet($discount->current_user_id, $package, 20)` and `tokenTransfer(1, $discount->current_user_id, $package->userpackage->price)` — **flagged for math agent** (20% figures, token transfer from company).
6. If `$findUser->node == 'gratitude'`: `tokenTransfer(1, $findUser->virtual_id, $package->userpackage->price)`, then `updateWallet($findUser->virtual_id, $package, 20)` (`$feePercentagegt = 20` inline).
7. `updateWallet(auth()->id(), $package, $feePercentage)` — credits the direct referrer/activator.
8. If `$package->user_id` NOT IN `[2,3,4,5]` (hardcoded excluded user ids — likely seed/test/company accounts): `packagePool($package->user_id, $package->package, $package->userpackage->price * 0.05)`.
9. Load the package owner `User`; if `leader_code` set, `updateWallet($users->leader_code, $package, 5, 'Leadership Bonus')`; if `executive_code` set, `updateWallet($users->executive_code, $package, 5, 'Leadership Bonus')`.

#### private tokenTransfer(int $fromUserId, int $toUserId, $package): bool
- `$amount = $package * (20/100)` (here `$package` param is actually a raw price value passed by callers, not a model — naming collision, be careful in the Go port).
- If `$amount <= 0`, return false.
- In `DB::transaction`: lock up to `$amount` of `$fromUserId`'s active tokens (`lockForUpdate()->limit($amount)`); if fewer than `$amount` exist, return false (no partial transfer); else bulk `Token::whereIn('id',...)->update(['user_id'=>$toUserId])`; `Log::info(...)`; return true.

#### private processPackageActivationCompany($package)
- Called only when `auth()->id() == 1`.
- `UserParent::where('user_id',$package->user_id)->first()` → `$findUser`. **Odd guard:** the null-check (`if (!$findUser) return 404 json (ignored, see above)`) is wrapped in `if (auth()->user()->id != 1) { ... }` — since this method is only ever called when `auth()->id() == 1`, the null-check block is **dead code**; if `$findUser` is actually null here, the very next line `$findUser->node = 'active'` will throw a fatal error (calling property on null) rather than being caught by the never-executing guard. Flag as a real bug.
- Sets `$findUser->node='active'`, save.
- Activates `$package` (`status='active'`, `activated_at`), activates owner `User.status='active'`.
- If package owner id not in `[2,3,4,5]`: `packagePool(...)` same 0.05 multiplier as the non-company path. **No `updateWallet` call at all in this branch** — company-initiated activations do not credit any commission wallet.

#### private updateWallet($parentId, $package, $feePercentage, $description=null)
(Distinct from — and duplicated with — `WalletService::updateWallet`/`creditWallet`; this private copy lives directly in `TokenController` and is used by all the calls above.)
- `$walletAmount = $package->userpackage->price * ($feePercentage/100)`.
- Same `GlobalShareWallet` bootstrap logic as `AuthController::processStep2` (hardcoded `$packageMultipliers` 5000..1,000,000 → 1.5×; create wallet with `max_out = highestPrice * multiplier` if none exists yet and user has an eligible package).
- **Unlike `WalletService::updateWallet`, this version has NO up-front wallet-cap gate** (`WalletService` checks `totalValue < wallet->balance` before crediting at all; this private method always credits). Flag as a real behavioral divergence between the two "wallet credit" implementations in the codebase — the Go port needs one canonical implementation and a decision on which behavior is authoritative.
- Credit `Wallet` (create if missing) by `$walletAmount`; create `EarnLog` (`description` defaults to `''` if null passed).
- If a `GlobalShareWallet` exists (existing or just bootstrapped) and `remaining = max_out - balance > 0`: credit `min($walletAmount, $remaining)` to it, log to `GlobalShareWalletLog` (**logs the full `$walletAmount`, not the capped `$credit` actually applied** — discrepancy between wallet balance change and log amount, flag for math agent).
- Earnings cap check: `$checkUserPackage = UserPackage::where('user_id',$parentId)->first()` (no status filter — first `UserPackage` row for that user regardless of status); if `earn <= userpackage.price * 4`: increment `checkUserPackage.earn += walletAmount`. Else: redirect the overflow to the company wallet (`user_id=1`) — increments company `Wallet.balance` and creates an `EarnLog` for user 1.

#### private deactivateTokens($needTokens)
- `Token::where('user_id',auth()->id())->where('status','active')->limit($needTokens)->get()`, set each to `status='deactive'`, save individually (N+1 saves, not bulk update — perf note only).

#### private packagePool($userId, $packageId, $amount)
- `PackagePool::create(['user_id'=>$userId,'package_id'=>$packageId,'pool_amount'=>$amount])`.

### GET /token-shares (token.share) — TokenController@shareToken
- **Auth/role required:** none declared (relies on `auth()->user()`)
- **Business logic:** `Token::where('user_id',auth()->id())->where('status','active')->count()`.
- **Response:** view `tokens.share` with `tokens` (count).

### POST /token/share (token.shares) — TokenController@shareTokens
- **Auth/role required:** none declared
- **Request input:** `tokenValue`: `required|integer|min:1`; `user_id`: `required|exists:users,id`
- **Business logic:**
  1. `$tokenCount = Token::where('user_id',$currentUser->id)->where('status','active')->count()` (computed but not directly compared — see step 3, which recomputes via the actual limited query).
  2. `$recipient = User::findOrFail($validatedData['user_id'])`.
  3. `$tokensToUpdate = Token::where('user_id',$currentUser->id)->where('status','active')->limit($tokenValue)->get()`.
  4. If `$tokensToUpdate->count() < $tokenValue`: `Alert::error('Error','You do not have enough active tokens to share.')`, redirect back.
  5. Else: for each fetched token, set `user_id = $recipient->id`, save (N+1 saves).
  6. `TokenLogs::create(['user_id'=>$recipient->id, 'shared_by'=>$currentUser->id, 'amount'=>$tokenValue])`.
- **Response:** success alert "Tokens sent successfully!", redirect back.

### GET /token/share/logs (token.share.log) — TokenController@shareTokensLog
- **Auth/role required:** none declared
- **Business logic:** `TokenLogs::with(['user','sharedBy'])->where('shared_by',auth()->id())->orderBy('created_at','desc')->paginate(10)`.
- **Response:** view `tokens.share-log` with `tokenLogs`.

---

## Api\AuthController (Sanctum API — `/api/*`)

### POST /api/token — Api\AuthController@token
- **Auth/role required:** none (public — issues the token)
- **Request input:** `username`: `required`; `password`: `required`
- **Business logic:** `ApiUser::where('username',$request->username)->first()`; `Hash::check($request->password, $user->password)`. If either fails → 401 `{message:'Invalid credentials'}`. Else `$user->createToken('api_token')->plainTextToken`.
- **Response:** 200 `{token: "<sanctum plaintext token>"}`. 422 on validation failure (standard Laravel `{message, errors:{field:[...]}}`).

### POST /api/check-user — Api\AuthController@checkUser
- **Auth/role required:** `auth:sanctum` (Bearer token from `ApiUser`)
- **Request input:** `email`: `required|email`
- **Business logic:** `User::where('email',$request->email)->first()` — **note: looks up the `User` table (end customers), not `ApiUser`**. If not found → 200 `{status:'fail', message:'User not found'}` (not 404). Else `$user->packages()->orderByDesc('created_at')->first()` → `$lastPackage`.
- **Response:** 200 `{status:'success', email, has_package: bool}` or 200 fail body above. 401 `{message:'Unauthenticated.'}` if token missing/invalid (Sanctum default). 422 on validation failure.

### GET /api/users — Api\AuthController@getAllUsers
- **Auth/role required:** `auth:sanctum`
- **Request input:** optional `per_page` (query, default 20)
- **Business logic:** `User::with(['secretKey','packages','mining'])->orderByDesc('id')->paginate($perPage)`; transform each row to `{id, user_id: secretKey.secret_key (bcrypt hash, used as an opaque external user identifier — see cross-cutting note), name, email, has_package: packages.count()>0, booked_token: mining.total_token ?? null}`.
- **Response:** 200 `{status:'success', users: <laravel paginator JSON with transformed `data`>}`. 500 `{status:'error', message:'Something went wrong.', error:<exception message>}` on exception (leaks internals — same pattern as `SalaryController::store`).

### POST /api/user/details — Api\AuthController@getSpecificUser
- **Auth/role required:** `auth:sanctum`
- **Request input:** `user_id`: `required|string` — **this is NOT the numeric `users.id`; it is the bcrypt `secret_key` string returned as `user_id` by `getAllUsers`/registration** (see `UserSecretKey` created in `AuthController::processStep1`).
- **Business logic:** `User::with(['secretKey','packages','mining'])->whereHas('secretKey', fn($q)=>$q->where('secret_key',$request->user_id))->first()`. **Important: `secret_key` is stored via `Hash::make()` (bcrypt) — bcrypt hashes are salted/non-deterministic, so comparing `secret_key = <given string>` with plain `where()` equality will only ever match if the caller round-trips the exact stored hash string verbatim (as returned by `getAllUsers`), not a freshly-hashed value of the same plaintext key. This is effectively using the hash as an opaque token, not as a verified secret — flag for the Go port: replicate as "opaque string key" lookup, not real hash verification.**
- **Response:** 200 `{status:'success', user:{id,user_id,name,email,has_package,booked_token}}` or 200 `{status:'fail', message:'User not found'}`. 500 on exception (`{status:'error', message:'Something went wrong.', error:<msg>}`). 422 on validation failure.

---

## Cross-cutting notes

### 1. Session (web) vs. token (API) auth are two entirely separate identity systems
- Web routes authenticate via the `web` guard (`session` driver) against `App\Models\User`, using `Auth::attempt()`/`Auth::user()`/`Auth::check()`/`Auth::logout()`. This is cookie/CSRF-based, standard Blade-form login.
- API routes (`routes/api.php`) authenticate via the `api` guard, `sanctum` driver, against a **completely different model**, `App\Models\ApiUser` (with its own `username`/`password`, presumably a small set of B2B integration accounts, not end customers). `POST /api/token` exchanges `username`+`password` for a Sanctum personal access token; all other `/api/*` routes require `Authorization: Bearer <token>` validated by Sanctum against tokens issued to `ApiUser` rows.
- Endpoints under `auth:sanctum` then operate on the **`User`** model (end customers) for their actual queries (`checkUser`, `getAllUsers`, `getSpecificUser`) — i.e., `ApiUser` is purely the API-caller's identity/credential, `User` is the business data being queried. The Go port needs two distinct principal types: an "API client" principal (bearer-token, maps to `ApiUser`) and an "end-user" session principal (cookie/JWT, maps to `User`), with the API surface always operating on end-user data regardless of which `ApiUser` called it (no per-`ApiUser` scoping/tenancy visible in the code).
- `getAllUsers`/`getSpecificUser` expose each `User`'s bcrypt `secret_key` (from `UserSecretKey`, generated once at registration in `AuthController::processStep1`) as an opaque `user_id` field, and `getSpecificUser` looks a user up **by that exact hash string**, not by re-hashing a plaintext credential. Since bcrypt hashes are unique per generation (random salt), this only works because the same stored hash string is being echoed back — it is being used purely as an opaque per-user token, not as a verified secret. Replicate this as a stored-string lookup, not a hash comparison, in Go.

### 2. RoleMiddleware / role enforcement
- Single middleware, alias `role`, parameterized with a comma-separated allow-list (`role:company`, `role:company,admin`, etc.). Strict string match on `User::role` (values seen in code: `admin`, `company`, `agent`, `user` — plus the sentinel behavior around `User.id == 1` used pervasively in `TokenController`/`CompanyController`/`DirectShareController` as an implicit "the platform/company operator account", independent of `role`).
- Failure mode is a single `abort(403, 'Unauthorized action.')` for both "not logged in" and "wrong role" — the Go port should likely distinguish 401 (no session) vs 403 (wrong role) unless byte-for-byte behavior parity is required, in which case collapse both to 403 with this message.
- **A large fraction of financially-sensitive endpoints have NO `auth` middleware at all** at the route level (`/active-package`, `/buy-packages`, `/token/share`, `/toggle-vacation`, `/kyc/*` including `verify`/`unverify`, `/earn/history`, `/my-geneology`, `/geneology/{id}`, `/mining/user/{id}`, etc. — see the route table above for the full list). They rely on `auth()->user()`/`Auth::user()` being truthy inside the method body, which for a guest either raises a PHP error (calling a method on null) or, in the read-only GET cases, may silently execute with a null "current user" causing further null-pointer errors downstream. **This is very likely unintentional** (i.e., a missing `->middleware(['auth'])` on these route definitions) rather than a deliberate "public" design, given that `/kyc/{id}/verify` and `/kyc/{id}/unverify` let anyone (including an unauthenticated caller) flip any KYC record's verification flag. Recommend the Go rewrite require authenticated sessions on all of these by default and flag this list explicitly to the product owner as a security gap being fixed, not silently ported.
- Route name collision: `admin.activateUser` is defined identically (same path/verb) in three different role groups (`admin`, `agent`, `user`); only the first-registered (`role:admin`) is actually dispatchable — the other two are unreachable dead routes, though `route('admin.activateUser')` name resolution in Blade/redirects would still work (last-registered name wins for URL generation, which happens to still resolve to the same URL anyway since all three share the same path).

### 3. Google Authenticator (2FA) flow
- Library: `sonata-project/google-authenticator` (`Sonata\GoogleAuthenticator\GoogleAuthenticator`) + `bacon/bacon-qr-code` for QR rendering.
- Setup: `GET /admin/{userId}/setup-google-auth` (role:company only) generates a **new** secret via `$gAuth->generateSecret()`, overwrites `User.google_authenticator_secret` for the **target `$userId`** unconditionally (no "already configured" guard — repeat visits invalidate any previously-scanned QR/app entry), builds an `otpauth://totp/signetint:{email}?secret={secret}&issuer=signetint` URL, and renders it as an SVG QR code (base64 data URI) for display — the secret itself is also shown in the view (`compact('qrCodeImageBase64','secret')`), presumably as a manual-entry fallback.
- Verification: only `TokenController::generateTokens` actually checks a TOTP code, via `$gAuth->checkCode($secret, $code)`, and it checks the **currently authenticated user's own** `google_authenticator_secret` — not the secret of the `$userId` route param whose tokens are being generated. So in practice, a company operator must set up their *own* 2FA (by visiting `/admin/{their-own-id}/setup-google-auth`) for `generate-tokens` to ever succeed; visiting the setup route for a different `$userId` sets 2FA for that other account, which is never checked by any endpoint in this codebase. This looks like a latent bug/incomplete feature — flag explicitly; the Go port should decide whether 2FA is meant to gate the *acting* company user (current behavior) or the *target* user, and implement deliberately rather than replicate the apparent bug silently.
- No 2FA challenge exists anywhere in the login flow (`AuthController::login`) — 2FA in this codebase is scoped solely to token generation, not general authentication.

### 4. Two independent, divergent "credit a wallet" implementations
- `App\Services\WalletService::updateWallet()` (used by `RocController::updateRocStatus` and `SalaryController::store`): gates the credit behind a check that the target user's `totalValue` (sum of `price*4` over their active packages) is **not** less than their current wallet balance — i.e., it silently no-ops (grants nothing) if the user is already at/above their earning cap.
- `TokenController::updateWallet()` (private method used throughout `processPackageActivation`/`processPackageActivationCompany`/`tokenTransfer`-adjacent flows): has **no such up-front cap gate** — it always credits the wallet amount, and only applies a *different*, per-package "earn" cap afterward (`UserPackage.earn <= price*4`) to decide whether the overflow routes to the user's own `UserPackage.earn` counter or diverts to the company wallet (`user_id=1`).
- Both versions independently reimplement: `GlobalShareWallet` bootstrap (same hardcoded `$packageMultipliers` map, duplicated verbatim in three places: `AuthController::processStep2`, `TokenController::updateWallet`, `WalletService::creditWallet`), `GlobalShareWalletLog` capped-credit logic, and `EarnLog` creation.
- **For the Go rewrite, unify these into one canonical wallet-crediting service** and explicitly decide which gating behavior (cap-before-crediting vs. cap-after via UserPackage.earn) is correct — do not port both divergent implementations as-is without a product decision, since they will produce different wallet balances for functionally identical "credit user X with amount Y" calls depending on which code path triggered the credit.

### 5. `SIG-00{id}` external ID formatting is convention, not a column
- Throughout `MiningController`, `UserController`, `SalaryController`, the "Signet ID" shown to users/searched by is always the literal string `'SIG-00' . $user->id}` (note: **not** zero-padded to a fixed width — `SIG-00` is a fixed 2-zero prefix concatenated directly with the raw numeric id, so id 5 → `SIG-005`, id 123 → `SIG-00123`). Inbound lookups strip this via the regex `^SIG-0*(\d+)$` (case-insensitive) to recover the numeric id, additionally stripping *any* number of leading zeros (not just the conventional two) via `0*` in the regex. Replicate both the display format and the lenient parback-parsing exactly.

### 6. `user_id == 1` as an implicit "company/system" sentinel
- Pervasive convention, not an explicit role: id `1` is treated as the company/root account in `TokenController` (`processPackageActivation` skip branch, `processPackageActivationCompany`, overflow-routing target `Wallet::where('user_id',1)`), `DirectShareController::directShare` (`companyPool` vs `salesPool` split), `DirectShareController::store` (hardcodes `user_id=1` on every pool row regardless of input), `PackageController::buyPackages` (fallback referrer when no active ancestor found). Also user ids `2,3,4,5` are hardcoded as excluded from `packagePool()` contributions in both `processPackageActivation` and `processPackageActivationCompany` — likely other seed/system accounts (leadership pool, admin test accounts, etc.) whose real identity should be confirmed against the seeders/DB before the Go port hardcodes the same ids.

### 7. Validation style is inconsistent across the codebase
- Most controllers use inline `$request->validate([...])` (throws `ValidationException` → Laravel's default redirect-back-with-errors for web routes, or 422 JSON for JSON/API requests) with rules documented per-endpoint above.
- Several endpoints accept `$request->input(...)` / `$request->field` with **no validation at all** (`TokenController::activePackage`'s `package_id`, `MiningController::update`'s three fields, `PackageController::buyPackages`'s `package`), meaning malformed/missing input silently becomes `null` and propagates into queries (`Model::find(null)` → null, `Model::where('col', null)` → matches NULL rows) rather than failing fast with a 422. The Go port should decide whether to add real validation here (recommended) or deliberately preserve the "silently do the wrong thing with null" behavior for parity — flag to the product owner either way.
- `RealRashid\SweetAlert` (`Alert::success/error/info/warning/toast`) is used for web-flash-message UX throughout; these have no JSON equivalent and are irrelevant to a pure API — the Go/Vue rewrite should replace them with structured error/success responses the Vue frontend renders itself, but must preserve the exact message text captured above where flows are behaviorally significant (e.g., "You already have an active package").

### 8. Pagination conventions
- All paginated endpoints use Laravel's default `paginate($n)`, `$n` varying per endpoint (10 most commonly; 15 for salaries/leader-executive logs; 20 for direct-share/leadership-bonus logs). The Go port should standardize on Laravel's paginator JSON shape (`current_page,data,first_page_url,from,last_page,last_page_url,links,next_page_url,path,per_page,prev_page_url,to,total`) for any endpoint currently rendering via Blade+paginator links, if those pages become API-driven Vue views, to minimize frontend rework — or explicitly document the new shape if diverging.

### 9. Mass-assignment via `$request->all()`
- `PackageController::store`/`update` call `Package::create($request->all())` / `$package->update($request->all())` directly — relies entirely on the `Package` model's `$fillable`/`$guarded` configuration (not deep-dived here; covered by the models agent) to prevent unexpected columns being set. Flag as a pattern to avoid in the Go port (bind to an explicit typed request struct with only the validated fields).

### 10. Views vs. JSON — no API versioning, mixed response types per controller
- The app is a monolithic Blade+SweetAlert2 server-rendered app with a handful of AJAX-style JSON endpoints sprinkled in (`Mining*`, `Salary*`, `UserController`'s AJAX update/search endpoints, `TokenController::activePackage`, `RocController::updateRocStatus`, `CompanyController::newActivepackage`). For the Go/Vue rewrite, essentially **every** endpoint needs to become JSON (the "Response" fields documented per-entry above list the current Blade view name + compacted variables where relevant, which map to what data the new Vue page/component needs) — the view names and their variable payloads throughout this document are the effective "page data contract" to replicate on the new frontend.
