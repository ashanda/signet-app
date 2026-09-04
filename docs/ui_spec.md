# Signet MLM/Crypto Platform — UI Specification

Reverse-engineered from `resources/views/` in the Laravel 12 source, for a 1:1 Vue 3 SPA rebuild.

## Styling stack (verified, not assumed)

- **Bootstrap 5** is the styling framework actually used on every authenticated/dashboard page and every auth page. Views use Bootstrap utility/component classes throughout: `row`, `col-*`, `card`, `card-header`, `card-body`, `table table-flush`, `btn btn-primary/success/danger/warning/secondary/outline-*`, `badge bg-*`, `modal fade`, `form-control`, `form-select`, `form-check form-switch`, `input-group`, `alert alert-*`, `progress`/`progress-bar`, `d-flex`, `collapse`, etc. Bootstrap JS (`data-bs-toggle="modal"`, `data-bs-dismiss`) is used for all modals.
- **Tailwind CSS v4** is wired into the build (`resources/css/app.css` = `@import 'tailwindcss'` plus `@theme { --font-sans: 'Instrument Sans' }`) and configured in `vite.config.js` via `@tailwindcss/vite`, but **no Tailwind utility classes were found in any Blade view**. Tailwind appears to be scaffolding left over from Laravel's default installer — it is not actually driving any visible UI. Treat the real design system as 100% Bootstrap 5 (+ vendor "Volt" dashboard theme CSS) for the rebuild.
- **Two separate front-end "skins"** exist:
  1. **Dashboard/app skin** (`layouts/app.blade.php`): loads `resources/css/volt.css` (the "Volt" Bootstrap 5 admin dashboard theme by Themesberg) and `resources/css/custome.css` (project overrides), plus SweetAlert2 CSS/JS, Notyf CSS, Font Awesome 5 (via CDN), Bootstrap JS bundle, Popper, jQuery 3.6, and the Pusher JS client (loaded but not observably wired up in the views read). This is used by literally every dashboard/auth/register/kyc/package/token/etc. page (i.e. everything that does `@extends('layouts.app')`).
  2. **Public marketing skin** (`layouts/main.blade.php`): a full separate "Signet"-branded one-page marketing template with its own Bootstrap build (`resources/main/css/bootstrap.min.css`), SlickNav, Swiper, Animate.css, Magnific Popup, Google Fonts (`DM Sans`), a "magic cursor" effect, WOW.js scroll animations, and a large custom footer. Used only by `welcome.blade.php` (the public landing page at `/`).
- **SweetAlert2** (`realrashid/sweet-alert` package, `@include('sweetalert::alert')` in `layouts/app.blade.php` for flashed alerts) is used everywhere for confirm dialogs and toasts — loaded both as a local vendor asset in the layout and re-loaded via CDN (`sweetalert2@11`) in many individual view `@section('scripts')` blocks (redundant double-load, but functionally consistent — same API used everywhere: `Swal.fire(...)`, `Swal.mixin({ toast: true, position: 'top-end', ... })` for toast-style notices).
- **jQuery 3.6/3.7** is used for AJAX calls (`$.ajax`) and DOM ready handlers on many pages; native `fetch()` is used interchangeably on others (inconsistently — different pages use different patterns for what is functionally the same "activate package" AJAX call).
- **Select2** (`select2@4.1.0-rc.0`, via CDN) is used for the "Leaders Gain"/"Executives Gain" scripts sections and the Salaries "Add Salary" modal user-search dropdown (AJAX-backed autocomplete).
- **Font Awesome 5** (`fa-*`, `fas`/`fab`/`fa-solid`/`fa-brands`) and **Bootstrap Icons** (`bi bi-*`) are both used for icons, plus a large number of hand-inlined SVG icon paths (heroicons-style) directly in Blade for sidebar nav icons and stat-card icons.
- No CSS-in-JS, no component library beyond Bootstrap. No custom design tokens/utility classes of note in `custome.css`/`volt.css` were inspected beyond being vendor CSS (not reproduced here since out of scope of Blade views, but note their existence as the true visual source — recommend pulling `resources/css/volt.css` and `resources/css/custome.css` for exact colors/spacing during the Vue rebuild).

## Roles & route map

Roles: `admin`, `company`, `agent`, `user` (from `role:` middleware in `routes/web.php`). `/dashboard` redirects to the role-specific dashboard route.

| Route name | Path | Controller@method | Role(s) | Blade view |
|---|---|---|---|---|
| `guest` | `/` | AuthController@guest | public | `welcome.blade.php` |
| `login` | `/login` | AuthController@showLoginForm | public | `auth/login.blade.php` |
| `login.post` | `/login` (POST) | AuthController@login | public | — |
| `logout` | `/logout` (POST) | AuthController@logout | auth | — |
| `password.request` | `/password/reset` | PasswordResetController@showResetRequestForm | public | `auth/password_reset.blade.php` |
| `password.reset` | `/password/reset/{token}` | PasswordResetController@showResetForm | public | `auth/password_reset_form.blade.php` |
| `register.step1` | `/register` | AuthController@showStep1 | public | `auth/register_step1.blade.php` |
| `register.step2` | `/register/step2/{id}` | AuthController@showStep2 | public | `auth/register_step2.blade.php` |
| `register.step3` | `/register/step3` | AuthController@showStep3 | public | `auth/register_step3.blade.php` (labelled "Step 4" in-page — see note) |
| — | (upliner wait screen) | AuthController@processStep2 (redirect target) | public | `auth/register_step2.blade.php`'s sibling content is actually served for the "wait for upliner" screen — see Auth section below for the file/label mismatch |
| `admin.dashboard` | `/admin/dashboard` | AdminController@index | admin | `admin/dashboard.blade.php` |
| `agent.dashboard` | `/agent/dashboard` | AgentController@index | agent | `agent/dashboard.blade.php` |
| `user.dashboard` | `/user/dashboard` | UserController@index | user | `user/dashboard.blade.php` |
| `company.dashboard` | `/company/dashboard` | CompanyController@index | company | `company/dashboard.blade.php` |
| `company.pending.activation` | `/company/pending-activation` | CompanyController@pendingActivation | company | `company/pending-activation.blade.php` |
| `new.activations` | `/new-activations` | CompanyController@newActivations | company | `company/new-activations.blade.php` |
| `setup.google.auth` | `/admin/{userId}/setup-google-auth` | GoogleAuthenticatorController@setupGoogleAuthenticator | company | `admin/setup_google_authenticator.blade.php` |
| `company.users` / also shared | `/users` | UserController@allUsers | company, admin | `company/users.blade.php` |
| `leader.code.logs` | `/leader-code-logs` | UserController@leaderCodeLogs | company, admin | `leader_code_logs/index.blade.php` |
| `executive.code.logs` | `/executive-code-logs` | UserController@executiveCodeLogs | company, admin | `executive_code_logs/index.blade.php` |
| `mining.users` | `/mining/users` | MiningController@index | company | `company/mining.blade.php` |
| `company.roc` | `/roc` | RocController@rocIncome | company | `company/roc_income.blade.php` |
| `company.direct_share` | `/direct-share` | DirectShareController@directShare | company | `company/direct_share.blade.php` |
| `company.direct_share_log` | `/direct-share-log` | DirectShareController@directShareLog | company | `company/direct_share_log.blade.php` |
| `userparentlogs.index` | `/user-parent-logs` | UserParentMapsLogController@index | company | `company/user_parent_logs/index.blade.php` |
| `salaries.index` | `/salaries` | SalaryController@index | company | `salaries/index.blade.php` |
| `leaders.gain` | `/leaders/gain` | LeaderController@index | company | `leaders/index.blade.php` |
| `executives.gain` | `/executives/gain` | ExecutiveController@index | company | `executives/index.blade.php` |
| `company.leadership_bonus_log` | `/leadership-bonus-log` | ExecutiveController@leadershipBonusLog | company | `company/leadership_bonus_log.blade.php` |
| `countries.*` (resource) | `/countries` | CountriesController | company | `countries/index.blade.php` |
| `packages.*` (resource) | `/packages` | PackageController | company | `packages/index.blade.php`, `create.blade.php`, `edit.blade.php` (+ `form.blade.php` partial) |
| `buy.package` | `/buy-package` | PackageController@buyPackage | any authenticated | `packages/buy-package.blade.php` |
| `buy.package.history` | `/buy-package-history` | PackageController@buyPackageHistory | any authenticated | `packages/buy-packageHistory.blade.php` |
| — (upliner wait) | (post buy-package redirect) | PackageController@buyPackages | any authenticated | `packages/buy-package-done.blade.php` |
| `earn.history` | `/earn/history` | EarnLogController@index | any authenticated | `earn/index.blade.php` |
| `token.share` | `/token-shares` | TokenController@shareToken | admin, user, agent | `tokens/share.blade.php` |
| `token.share.log` | `/token/share/logs` | TokenController@shareTokensLog | admin, user, agent | `tokens/share-log.blade.php` |
| `view.tokens` | `/view-tokens/{userId}` | TokenController@viewTokens | company | `tokens/index.blade.php` |
| `my.geneology` | `/my-geneology` | GeneologyController@index | any (non-company only shown in nav) | `geneology/index.blade.php` |
| `geneology.show` | `/geneology/{userId}` | GeneologyController@viewGeneology | any | `geneology/show.blade.php` |
| `kyc.index` | `/kyc` | KycController@index | company (list) | `kyc/index.blade.php` |
| `kyc.show` | `/kyc/show` | KycController@show | non-company (own KYC) | `kyc/show.blade.php` |
| `kyc.create` | `/kyc/create` | KycController@create | any | `kyc/create.blade.php` |
| `kyc.edit` | `/kyc/{id}/edit` | KycController@edit | any | `kyc/edit.blade.php` |
| `kyc.verified` | `/kyc/verified` | KycController@verified | company | `kyc/verified.blade.php` |

## Register flow file/label mismatch (verbatim quirk worth knowing before rebuilding)

The four `auth/register_step*.blade.php` files do **not** map 1:1 to their filenames' step numbers in their on-page copy:
- `register_step1.blade.php` → "Step 1 of 2 — Create your account" (the real account-details form).
- `register_step2.blade.php` → "Step 2: Select a Package" (package picker).
- `register_step3.blade.php` → "Step 3: Get Your Upliner Details" (shows the referring upliner's Binance ID / WhatsApp, a "Your Upliner is On Vacation" warning badge if applicable, a Call-Now button, and a Next button that links to `register.step3` again with the new user id — i.e. this view is actually shown as an intermediate "wait" screen, likely rendered by `processStep2`).
- `register_step4.blade.php` → "Step 4: Account Status" (final status screen: pending/active/inactive messaging + Go to Login button) — this file's Blade comment header literally still says `{{-- resources/views/auth/register_step3.blade.php --}}` (copy-paste leftover), even though its route is `register.step3` per web.php and its heading says "Step 4". Reproduce the on-screen labels/copy as written; don't "fix" the mismatch unless product asks.

---

## Public site

### `welcome.blade.php` — used at `/` (route `guest`)
- **Layout/role:** `@extends('layouts.main')` — the separate marketing skin, public, unauthenticated.
- **Navigation:** Top nav (`layouts/main.blade.php`) — logo, "Home", "About us", "Contact" (all `javascript:void(0)` except Home), "Login" → `route('login')`.
- **Page structure (in order):**
  - **Hero**: eyebrow "Welcome to Signet", H1 "Easy Way to **Signet** Key To Success.", paragraph, "Join for Free" button.
  - **Coin ticker**: horizontally auto-scrolling marquee of 14 crypto logos+names (Bitcoin, Etherium, Tether, BNB, Solana, USD Coin, Cardano ×2, Dogecoin, Tron, Polygon, Shiba INU, Lite Coin, Stacks), duplicated twice for seamless loop.
  - **About Signet**: eyebrow "About Signet", H2 "Simple. Faster. Secure", two stacked images, body copy about the $10 signal service and $5,000 growth pitch, bullet list, "Read More" button linking to a static PDF asset.
  - **How It Works** ("SIGNET Fund Building Plan"): 3 columns — subscription tiers & profit %, reinvestment explanation, referral-commission pattern explanation (1st/3rd commission to you, 2nd/4th to your introducer, repeating).
  - **Price section**: eyebrow "Price", H2 "Explore Cryptocurrency Price", two Swiper carousels (left/right) of ~24 static coin price cards (icon, name, static `$12.185` price, green/red % badge) — purely decorative/static data, not live.
  - **Why Choose Us**: 3 icon items — "Safe & Secure", "Early Bonus", "Several Profit" (no body copy, headings only).
  - **Footer** (see Shared Components).
  - Several sections are HTML-commented-out in source (Exclusive Partners, Our Services, Download Apps, Crypto Calculator, Our Team, Testimonials, Latest Posts) — present as templated markup but not rendered; skip in the rebuild unless asked to restore them.
- **Interactions:** WOW.js scroll-reveal (`wow fadeInUp` + `data-wow-delay`), `text-anime` heading animation class, Swiper carousels auto-init, magic cursor effect (`#magic-cursor`/`#ball`), preloader overlay on load.

---

## Auth pages

### `auth/login.blade.php` — `route('login')` / `route('login.post')`
- **Layout/role:** `@extends('layouts.app')`, public, uses only `@section('content')` (no sidebar/topbar).
- **Page structure:** Centered white card (`fmxw-500`) over `bg-soft` section. "Back to Homepage" link with left-arrow icon above the card. Logo (80px). H1 "Sign in to Signetint platform". Session `error` flash shown as `alert-danger`.
- **Form fields:**
  - Email — `type=email`, name `email`, icon-prefixed input-group, placeholder `example@company.com`, `autofocus`, `required`.
  - Password — `type=password`, name `password`, icon-prefixed input-group, placeholder `Password`, `required`, plus a clickable eye-icon toggle (`togglePassword()`) that swaps input type and swaps between an open-eye/closed-eye inline SVG.
  - "Remember me" checkbox, name `remember`.
  - "Lost password?" link → `route('password.request')`.
  - Submit button "Sign in" (`btn btn-gray-800`, full width `d-grid`).
- **Interactions:** inline `togglePassword()` JS toggles password visibility and swaps the eye icon paths' `display`.

### `auth/password_reset.blade.php` — `route('password.request')` / `route('password.email')`
- **Layout/role:** `@extends('layouts.app')`, public, content-only.
- **Page structure:** "Back to log in" link, H1 "Forgot your password?", helper paragraph, `session('status')` success alert.
- **Form:** single Email field (`@error('email')` inline validation styling with `.invalid-feedback`), submit button "Recover password" (`btn btn-gray-800`, full width).

### `auth/password_reset_form.blade.php` — `route('password.reset', $token)` / `route('password.update')`
- **Layout/role:** `@extends('layouts.app')`, public, content-only.
- **Page structure:** "Back to log in" link (hardcoded `./sign-in.html`, dead link — note as a bug to not necessarily replicate literally, but documents actual current behavior), H1 "Reset password".
- **Form fields:** hidden `token`; Email (prefilled `$email ?? old('email')`, with `@error` styling); Password (icon-prefixed input-group, `@error` styling); Confirm Password (icon-prefixed input-group). Submit "Reset password" (`btn btn-gray-800`).

### `auth/register_step1.blade.php` — `route('register.step1')` / posts to `register.processStep1`
- **Layout/role:** `@extends('layouts.app')`, public, content-only. Custom inline gradient background on the `<section>`.
- **Page structure:** "Back to log in" link, card with rounded-4 shadow-lg, header shows a pill badge "Step 1 of 2" + "Account details" label, H1 "Create your account", subtitle, a 50%-filled progress bar.
- **Form — sectioned into 3 groups with small uppercase section headers:**
  1. **Referral & management**
     - Referral Code — text, `readonly`, prefilled from `?ref=` query string (`request('ref')`).
     - Manager Code (`leader_code`) — select, options `SIG-00{id} - {name}` for each `$leaders`, default "Select Manager".
     - Executive Code (`executive_code`) — select, same options list, default "Select Executive".
  2. **Personal details**
     - Country (`country`) — select, required, options `{code} - {name}` for each `$countries`.
     - Full Name (`name`) — text, required, placeholder "Jane Doe".
     - Email (`email`) — email, required, placeholder "jane@example.com".
     - WhatsApp Number — compound input-group: country-code select (`country_code`, ~150 hardcoded dial-code options e.g. "+1 (USA, Canada)" … "+263 (Zimbabwe)", required) + text input `whatsapp_number` (required, `pattern="^\d{9}$"`, placeholder `712345678`, helper text "Enter 9 digits without the leading zero.").
  3. **Payment & security**
     - Binance Pay ID (`binance_pay_id`) — text, required.
     - Password (`password`) — password, required.
     - Confirm Password (`password_confirmation`) — password, required.
  - Submit: "Continue to Step 2" (`btn btn-primary btn-lg w-100 rounded-3`).
  - Footer link: "Already have an account? Sign in" (dead `./sign-in.html` href).
- **Interactions:** JS guards so Manager Code and Executive Code selects cannot both hold the same user — on `change` of either, if it now equals the other's value, a SweetAlert2 warning ("Invalid Selection" / "Manager and Executive Manager cannot be the same user.") fires and resets the changed select to blank.

### `auth/register_step2.blade.php` — `route('register.step2', $id)` / posts to `register.processStep2`
- **Layout/role:** `@extends('layouts.app')`, public, content-only, same card chrome as step1 (no progress bar/badge here).
- **Page structure:** H1 "Step 2: Select a Package".
- **Form:** hidden `newUserID` = `$id`; Package select (`package`, required) with options `{$pack->name} USD` for each `$package`. Submit "Next" (`btn btn-primary mt-4`).

### `auth/register_step3.blade.php` (content = "Step 3: Get Your Upliner Details", route `register.step3`)
- **Layout/role:** `@extends('layouts.app')`, public, content-only.
- **Page structure:** H1 "Step 3: Get Your Upliner Details", "Upliner Activation" sub-heading, shows `Binance ID: {{ $parentData->binance_pay_id }}`, WhatsApp number with a WhatsApp icon, a "Call Now" button (`tel:` link, `btn btn-success`), and a "Next" button/link → `route('register.step3', ['id' => $user])` (self-referential — presumably re-polls status).

### `auth/register_step4.blade.php` (content = "Step 4: Account Status")
- **Layout/role:** `@extends('layouts.app')`, public, content-only. (File's own leading comment mislabels it as step3 — see quirk note above.)
- **Page structure:** H1 "Step 4: Account Status". Conditional message by `$user->status`: `pending` → "Your account is currently pending. Please wait for activation."; `active` → "Your account is active. You can now login!"; else → "Your account is inactive. Please contact support." Then a fixed line "We'll notify you when your account has been activated." An `alert-info` box showing "Your current status: **{{ $user->status }}**". "Go to Login" button (`btn btn-primary`) → `route('login')`.

---

## Admin role

### `admin.blade.php` (top-level, distinct from `admin/` folder)
A minimal standalone Blade file (not part of the `layouts.app` chrome) — no `@extends`. Just:
```
@if(auth()->user()->role == 'admin') <h1>Welcome, Admin!</h1> @else <h1>Access Denied</h1> @endif
```
Not wired to a route in `web.php` as read; appears to be an orphaned/legacy view. Skip in rebuild unless the app references it elsewhere.

### `admin/dashboard.blade.php` — `route('admin.dashboard')`
- **Layout/role:** `@extends('layouts.app')`, admin dashboard, includes `layouts.sidebar` + `layouts.topbar`.
- **Page structure (top to bottom):**
  - **Welcome/profile card** (`bg-yellow-100`): "Welcome, **{role}** Head", rank/ID line `{{ $myPackage->userpackage->rank }} SIG-00{id}`, raw-HTML injected rank summary bar (from `rank()` helper — see Shared Components → Helper widgets), raw-HTML ROC summary line (from `roc()` helper, shown only if `auth()->user()->roc_status == 'active'`), a Vacation toggle switch (`#vacationSwitch`, labeled "Vacation ON/OFF"), a referral-link box: label "Your Referral Link", readonly text input prefilled with `$refLink`, "Copy URL" button with clipboard icon, and an (empty until JS fills it) `#urlDisplay` div.
  - **Stat cards row** (`col-12 col-sm-6 col-xl-4`, each `card border-0 shadow` with an icon-shape + label/value pair, doubled for mobile (`d-sm-none`) vs desktop (`d-none d-sm-block`) layouts of the same content):
    - "My Token" → `{{ $myTokens }} USDT`.
    - "My Wallet" → `{{ $totalValue - walletBalance(user) }} USDT` (× 2, once with USDT suffix once without — duplicate card in source, worth normalizing to one in Vue).
    - "My Global Director Share Wallet" → `{{ $myGlobleDirectorShare->balance }}` — only if `$myGlobleDirectorShare != null`.
    - "My Earn" → `{{ walletBalance(user) }} USDT`.
    - "This Time Direct Share Pool value" → `{{ $poolAmount }} USDT`.
    - "This Time All Share Value" → `{{ $totalPoolshareValue }}`.
    - "This Time My Share Portion" → `{{ (int)($myshareValue) }} USDT` + small-text "This Time My Share Value: {{ (int)auth()->user()->global_director_share }}" — only if `auth()->user()->global_director_share_status == 1`.
  - **Mining Community Staking Token card** (full width, `card shadow-lg`, header `bg-primary text-white`): title "⛏️ Mining Community Staking Token - All Users Count({{ allUsers() }})" + a connection-status badge (`#connectionBadge`, starts "Connecting…", flips to green "Connected"/red "Disconnected"). Body: 4-up stat row (Mining Token, Total Token, Daily Mining, Status badge), a striped/animated Bootstrap progress bar (`#progressBar`, `#progressPercent`), a mining-rate badge ("N tokens/second"), and "Last updated: …" text.
  - **Activations table card**: header "Activations" + "See all" button (dead `#` link). Table columns: **User name**, **Whats app**, **Need Tokens** (badge showing `price - fee%`), **Action** (either a green "Wait for company activation" badge when `company_status == 0`, or an "Active" primary button `data-id="{id}"` otherwise). Paginated (`$activations->links()`).
  - A collapsed "theme-settings" panel (Volt template's stock GitHub-star/download promo widget) — vendor boilerplate, not app-specific; omit from rebuild.
- **Interactions:** `.active-package` buttons POST to `/active-package` via jQuery AJAX with `package_id`; success shows SweetAlert2 success then reloads page; error shows SweetAlert2 error. `#vacationSwitch` change POSTs to `route('user.toggleVacation')`; toggles the label text and fires a SweetAlert2 toast. A full client-side "mining dashboard" simulation: `loadData()` fetches `/mining/user/{userId}` on load, `startLocalMining()` ticks the on-screen mining counter every second locally (client-side simulated increment), `startServerSync()` re-fetches from the server every 5s to reconcile; a completion SweetAlert2 fires when `mining_token >= total_token`.

### `admin/pending-activation.blade.php`
- **Layout/role:** `@extends('layouts.app')`, admin, sidebar+topbar. (Not currently linked from `web.php` routes shown — route commented out — but view exists; keep for parity.)
- **Page structure:** Flash alerts (`session('success')`/`error`). H1 "Pending Activations". Card with table — columns **User name**, **Whats app**, **Need Tokens** (badge, price minus fee%), **Package Value**, **Action** (primary "Active" button, `data-id`). Paginated.
- **Interactions:** same `.active-package` → POST `/active-package` AJAX pattern as admin dashboard, SweetAlert2 success/error, reload on confirm.

### `admin/setup_google_authenticator.blade.php` — `route('setup.google.auth', $userId)`
- **Layout/role:** `@extends('layouts.app')`, admin/company (route is under the company middleware group), sidebar+topbar.
- **Page structure:** Centered text container. H2 "Setup Google Authenticator", instruction paragraph "Scan the QR code below with your Google Authenticator app.", QR code image (`$qrCodeImageBase64`, `img-fluid`), "Secret Key: {{ $secret }}" bold label + value, "Use this key if you cannot scan the QR code." A commented-out alternate branch for "already set up" state exists in source but is disabled (`{{-- --}}`) — not currently reachable.

---

## Agent role

### `agent/dashboard.blade.php` — `route('agent.dashboard')`
- **Layout/role:** `@extends('layouts.app')`, agent, sidebar+topbar.
- **Page structure:** Identical in structure/content to `admin/dashboard.blade.php`'s welcome card, stat cards (My Token, My Wallet, My Global Director Share Wallet conditional, My Earn, This Time Direct Share Pool value, This Time All Share Value, This Time My Share Portion conditional), Mining Community Staking Token card, and Activations table — byte-for-byte the same component set as admin's dashboard (this file and `admin/dashboard.blade.php` are effectively duplicated views; in the Vue rebuild these should become one shared `DashboardPage` component parameterized by role label). One structural difference: the mining card here carries class `mb-4` (admin's doesn't) and total/daily default text shows `0` instead of `0.00000000` in two of the stat placeholders — cosmetic only.
- **Interactions:** identical to admin dashboard (active-package AJAX with SweetAlert2, vacation toggle, full mining-dashboard client simulation script).

---

## Company role

### `company/dashboard.blade.php` — `route('company.dashboard')`
- **Layout/role:** `@extends('layouts.app')`, company, sidebar+topbar.
- **Page structure:**
  - Header card: "Welcome, **{role}** Head - All Users Count({{ allUsers() }})" (a large commented-out "Company Wallet" stat block exists in source but is disabled).
  - **All Tokens table card** (8-col width): header "All Tokens" + "See all" (dead link). Columns: **User name**, **Total Tokens**, **Active Tokens** (green badge), **Used Tokens** (red badge). Paginated.
  - **Generate Token card** (4-col width, `shadow-lg rounded-4`, gradient primary header "Generate Token"): a Select "Select User" populated from *all* `User::all()` (label = "Company Head" for id 1, else `SIG-00{id}`), and a "View Tokens" button that calls `redirectToTokens()` → navigates to `/view-tokens/{userId}`.
  - **Filter toolbar card** (`bg-light`): GET form with From Date / To Date date inputs, "Apply Filter" (success) and "Reset" (outline-dark, links back to `route('company.dashboard')`) buttons.
  - **KPI summary row** (3 cards): "Total Activations" (`number_format(sum active_count)`), "Total Value" (`number_format($grandTotal) USDT`), "Packages Activated" (`count`).
  - **Package Activation Summary card**: header + a badge showing grand total (only if `$grandTotal > 0`). Body: for each `$packageWiseCounts` entry — a sub-card with package-name pill badge, right-aligned total value + "{n} Activations" small text, and a rounded progress bar showing that package's % share of the grand total. Empty state: "No activations found".
  - **Grand Total block** (dark card, only if `$grandTotal > 0`): "Total Activation Revenue" label + big green total.
  - Theme-settings vendor promo panel (same as admin dashboard — omit).
- **Interactions:** `redirectToTokens()` client-side redirect using the select's value (alerts if none chosen).

### `company/direct_share.blade.php` — "Globle Direct Share" — `route('company.direct_share')` [note: literal on-page spelling "Globle" is a typo in the source — reproduce verbatim as it's user-facing copy, but flag to product]
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts. H1 "Globle Direct Share". Card containing:
  - Date-range search form (`start_date`, `end_date` — GET), "Filter" button, conditional "Clear" link (shown when either date is set).
  - A single wide summary tile: "Per Month Total Pool Value" / "Sales Through Get Pool Value" / "Company Included the Pool Value" (each `USDT {number}`), with a coin icon, subtitle "Total for selected Date Range".
  - "Insert Pool Amount" inline form: hidden `user_id=1`, number input `pool_amount` (step 0.01, required), "Save" button — POSTs to `route('package-pools.store')`.
  - **Pools table**: columns **User name**, **SIG ID**, **Package Value** (or `-` for user_id 1), **Pool Amount**, **Date**, **Actions**. Actions column: only for `$pool->user_id == 1` rows — an edit button (warning, pencil icon, `data-bs-target="#editPoolModal"`, carries `data-id`/`data-amount`) and a delete button (danger, trash icon, `onclick="deletePool(id)"`); all other rows show a plain "Auto" secondary badge instead of actions. Paginated, empty state "No pools data found".
  - **Edit Pool Amount modal** (`#editPoolModal`): single "Pool Amount" number field, "Update" submit button; JS populates the field + form action from the clicked row's `data-*` on `show.bs.modal`.
- **Interactions:** `deletePool(id)` shows a SweetAlert2 confirm ("Are you sure? This pool will be deleted permanently.") before submitting a hidden per-row delete form (`DELETE` via `@method('DELETE')`). Unused-on-this-page leftover script block also wires a `.toggle-status` handler posting to `company.updateRocStatus` (dead code carried over from the ROC page template — not present in this page's markup, so effectively inert here).

### `company/direct_share_log.blade.php` — "Globle Direct Share Log" — `route('company.direct_share_log')`
- **Layout/role:** `@extends('layouts.app')`, company. Structurally a sibling of `direct_share.blade.php`.
- **Page structure:** H1 "Globle Direct Share Log - {{ $startDate }} to {{ $endDate }}". Same date-range filter form pattern. Table columns: **User name**, **SIG ID**, **Binance Pay ID** (with an inline "copy" icon-button next to the value), **Amount**, **Date**. Paginated, empty state "No pools data found". No Actions column here (log is read-only).
- **Interactions:** `copyBinance(value)` copies the Binance Pay ID to clipboard and shows a plain browser `alert()` (not SweetAlert2, inconsistent with rest of app) confirming or reporting failure. (Leftover unused edit-pool-modal markup/script also present but not linked to any button on this page — dead code.)

### `company/leadership_bonus_log.blade.php` — "Leadership Bonus Log" — `route('company.leadership_bonus_log')`
- **Layout/role:** `@extends('layouts.app')`, company. Same skeleton as direct_share_log.
- **Page structure:** H1 "Leadership Bonus Log - {{ $startDate }} to {{ $endDate }}". Same date filter form. Table columns: **User name**, **SIG ID**, **Amount**, **Date**. Paginated, empty state "No pools data found". (Leftover unused edit-pool-modal + ROC-toggle script again present as dead code, same as the two views above — this looks like all three "log" pages were cloned from one template without trimming.)

### `company/mining.blade.php` — "Mining Token" — `route('mining.users')`
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts, H1 "Mining Token". Search card: text input "Enter Signet ID" (`#search_signet_id`) + "Search" button; results render into `#resultCard` (hidden until populated).
- **Result card (client-rendered via JS template string)**: gradient purple header with user name/ID and a status pill (Active/Inactive), body sections: Email; Packages (First Package / Current Package / Sale Count, 3-up mini tiles); Mining Data (Total Token / Mining Token / Daily Mining, 3-up colored tiles); footer "Edit Mining Data" warning button (`data-id`, `data-daily_mining`, `data-total_token`, `data-status`).
- **Update Daily Mining Token modal** (`#updateModal`): hidden `id`; Daily Mining Token text field, Total Issued Token text field, Status select (Active/Inactive). "Update" submit.
- **Interactions:** Enter-key or button click triggers `searchSignet()` → `GET /mining/search/{id}`; renders the result card or a SweetAlert2 "Not Found" error. Clicking "Edit Mining Data" fills and opens the modal. Modal submit → `POST /mining/update/{id}` (FormData/fetch); success → SweetAlert2 "Updated!" then page reload; failure → SweetAlert2 error.

### `company/new-activations.blade.php` — "New Packages/TopUp Activations" — `route('new.activations')`
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts, H1 "New Packages/TopUp Activations". Table columns: **User name**, **Whats app**, **Package** (name), **Action** ("Active" primary button, `data-id`). Paginated.
- **Interactions:** `.active-package` click → POST `/company/new_active-package` (jQuery AJAX) with `package_id`; SweetAlert2 success/error, reload page on confirm.

### `company/pending-activation.blade.php` — "Pending Activations" — `route('company.pending.activation')`
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Same as `new-activations.blade.php`: flash alerts, H1 "Pending Activations", table with **User name / Whats app / Package / Action** ("Active" button). Paginated.
- **Interactions:** `.active-package` → POST `/active-package` (note: different endpoint than new-activations' page, `/active-package` vs `/company/new_active-package` — same UI pattern, different backend route).

### `company/roc_income.blade.php` — "ROC Income" — `route('company.roc')`
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts, H1 "ROC Income". Filter form: a "Select Week" dropdown (`job_id`, options `{week_start} to {week_end}`), "Filter" submit, conditional "Clear" link.
- **Summary tiles** (2-up row): "Per Week Total" (+ derived "5% Week Total" beside it) in one tile; "Total Paying ROC" (`total_amount - balance_forward`) in a second tile with a down-trend icon. (A third "Previous/Balance Forward" tile exists in source but is commented out.)
- **Table**: columns **User name**, **SIGNET ID**, **Binance ID**, **WhatsApp**, **Earnings**, **Status** — Status cell is a toggle switch (`.toggle-status`, checked = "paid") with a text label next to it showing the current status capitalized. Paginated, empty state "No users found".
- **Interactions:** toggling `.toggle-status` POSTs JSON `{id, status}` to `route('company.updateRocStatus')`; on success updates the label text and fires a top-end SweetAlert2 toast ("Status changed to Paid/Pending"); on failure/error reverts the checkbox and fires an error toast.

### `company/user_parent_logs/index.blade.php` — "Fake Account (Deactive Users, 10+ hours old)" — `route('userparentlogs.index')`
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** dismissible success alert if flashed. H1 "Fake Account (Deactive Users, 10+ hours old)". Table columns: **#** (log id), **Activation Arrived User** (parent user's name + "SIG ID: SIG-00{id}" sub-line, or "User not found"), **New User** (same pattern for `$log->user`), **Status** (secondary badge showing `$log->user->status`), **Created At** (timestamp + relative "diffForHumans" sub-line), **Action** (right-aligned danger button "Delete mappings" with trash icon, `data-id`, `data-url`). Paginated, empty state "No logs found for deactive users older than 10 hours."
- **Interactions:** delete button → SweetAlert2 confirm ("Are you sure? This will delete all UserParent rows linked to this mapping log!") → on confirm, JS builds and submits a hidden POST form with `_method=DELETE` to the row's destroy URL.

### `company/users.blade.php` — "Find Users" — `route('company.users')` (shared with admin)
- **Layout/role:** `@extends('layouts.app')`, company + admin.
- **Page structure:** Flash alerts. Header row: H1 "Find Users" + "Leader Code Logs" outline-primary button (clock-history icon) → `route('leader.code.logs')`.
  - **Quick lookup card**: "Enter Signet ID" text input + "Search" button; results render client-side into `#resultCard`.
  - **Search/filter card**: GET form, text input "Search by name, ID, WhatsApp, or status, Country Code (eg: US, IN, etc.)" (`search`), "Search" submit, conditional "Clear" link.
  - **Users table**: columns **User name**, **SIGNET ID**, **WhatsApp**, **Status**, **Leader Status** (a toggle switch + label for role=company viewers; plain text for others), **Leader** (current leader's `SIG-00{id}` or "—", plus a small outline "edit" pencil-icon button that opens the Assign Leader modal, carrying `data-id/name/leader-code/executive-code`), **Executive** (same pattern for executive assignment). Paginated, empty state "No users found".
- **Result card (client-rendered)**: gradient header with name/ID and 3 status pills (Account status Active/Danger, "ROC: …" success/warning/danger, "Global Director Share: Active/Inactive"); body: Email; Packages section (First/Current/Sale Count/Direct Sale Count/Wallet Balance — 5 mini tiles); three full-width footer buttons: "Change Account Status" (warning), "Change ROC Status" (primary), "Change Global Director Share" (success).
- **Modals:**
  - **Update User Status** (`#updateModal`): Status select (Active/Inactive) → `POST /users/update/{id}`.
  - **Update User ROC Status** (`#updateModalRoc`): ROC Status select (Active/Inactive) → `POST /users/update-roc/{id}`.
  - **Update Global Director Share** (`#updateModalGlobalDirectorShare`): numeric "Global Director Share Value" input + Status select (Active=1/Inactive=0) → `POST /users/update-global-director-share/{id}`.
  - **Assign Leader** (`#updateModalLeaderCode`): title suffixed with the user's name; Leader select ("No Leader" + `SIG-00{id} - {name}` per `$leaders`) → `POST /users/update-leader-code/{id}`. Guard: if the chosen leader equals that row's already-assigned executive, a SweetAlert2 warning fires and clears the selection ("This person is already assigned as the Leader"/Executive, mirrored both directions).
  - **Assign Executive Code** (`#updateModalExecutiveCode`): same pattern → `POST /users/update-executive-code/{id}`.
- **Interactions:** all four update forms submit via `fetch` + FormData, SweetAlert2 "Updated!" then `location.reload()` on success, SweetAlert2 error otherwise. Row toggle `.toggle-leader-status` POSTs JSON to `/users/update-leader-status/{id}` with a top-end toast pattern identical to ROC Income's toggle. `searchSignet()` same client-rendered-card pattern as Mining page, hitting `GET /users/search/{id}`.

---

## Cross-role: Countries, Packages, Users find-index sub-pages

### `countries/index.blade.php` — "Countries List" — `route('countries.index')` (company)
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** H3 "Countries List", "Add New Country" primary button (opens modal). Session success alert. Table: `table-bordered table-striped`, columns **#**, **Country Code**, **Country Name**, **Actions** (outline-primary "Edit" button opening the modal pre-filled via inline `onclick="editCountry({{ $country }})"` passing the whole model as JSON; outline-danger "Delete" button inside a per-row form, `onclick="return confirm(...)"` — uses the native browser `confirm()`, not SweetAlert2, inconsistent with the rest of the app). Paginated (`$countries->links()`).
- **Add/Edit Country modal** (`#countryModal`): Country Code text input, Country Name text input, "Save" submit, "Close" button. Title and form action/method swap between "Add Country"/`countries.store` (POST) and "Edit Country"/`/countries/{id}` (PUT, via injected `@method('PUT')` HTML) depending on which trigger button was clicked.

### `packages/index.blade.php` — "Package List" — `route('packages.index')` (company)
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** H3 "Package List", "Add New Package" button → `route('packages.create')`. Session success alert. Table (`table-bordered`): columns **#**, **Name**, **Price**, **Commission**, **Status** (badge: green "Active" / gray "Deactive"), **Rank**, **Actions** (outline-primary "Edit" → `route('packages.edit', id)`; outline-danger "Delete" form with native `confirm()`).

### `packages/form.blade.php` — shared partial for create/edit
- **Fields:** Name (text, required), Price (number step 0.01, required), Commission (number step 0.01, required), Rank (text, required), Telegram Link (text, placeholder `https://t.me/yourchannel`, optional), Status (select: Active/Deactive, required). All values default from `old()` falling back to `$package->{field}` when editing.

### `packages/create.blade.php` — `route('packages.create')` / `route('packages.store')`
- H3 "Create New Package", includes `packages.form`, "Save" submit button.

### `packages/edit.blade.php` — `route('packages.edit', $package)` / `route('packages.update', $package)`
- H3 "Edit Package", includes `packages.form` with `$package` passed in, method PUT, "Update" submit button.

### `packages/buy-package.blade.php` — "Buy Package" — `route('buy.package')` / posts `route('buy.packages')`
- **Layout/role:** `@extends('layouts.app')`, any authenticated user, content-only, same centered-card chrome as auth pages.
- **Form:** single Package select (`package`, required, options `{name} USD`), "Next" submit button.

### `packages/buy-package-done.blade.php` — post-purchase "wait for upliner" screen — target of `buy.packages` POST
- **Layout/role:** `@extends('layouts.app')`, content-only, same card chrome.
- **Page structure:** "Upliner Activation" heading, Binance ID + WhatsApp number of `$parentData`, WhatsApp icon, "Call Now" button (`tel:` link), "Next" link → `route('buy.package.history')`.

### `packages/buy-packageHistory.blade.php` — "Buy Packages History" — `route('buy.package.history')`
- **Layout/role:** `@extends('layouts.app')`, any authenticated user, sidebar+topbar.
- **Page structure:** Flash alerts, H1 "Buy Packages History", "Buy Package" button → `route('buy.package')`. Table columns: **User name**, **Package** (name), **Earn** (green badge, `USDT`), **Buy Date** (info badge, raw timestamp). Paginated.

---

## Financial/report list pages (Leaders, Executives, Leader/Executive Code Logs, Earn, Salaries, Tokens, KYC, Geneology)

### `leaders/index.blade.php` — "Leaders Gain" — `route('leaders.gain')` (company)
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts, H1 "Leaders Gain". Filter form (GET, not date-range only): Leader select (`leader_id`, "All Leaders" + each `$leaders`, `@selected` to preserve current filter), From/To date inputs, "Search" button. Table columns: **Leader**, **SIGNET ID**, **Binance ID** (+ inline copy button, shown only if `binance_pay_id` set), **Total Package Value** (`$` formatted), **5% Gain** (`$` formatted, = total × 0.05). Paginated.
- **Interactions:** `copyBinanceId(id)` copies to clipboard, shows a SweetAlert2 top-end toast "Binance ID copied!". Select2 + SweetAlert2 CDN loaded but Select2 isn't actually applied to any element on this specific page's markup (leftover from a shared script partial).

### `executives/index.blade.php` — "Executives Gain" — `route('executives.gain')` (company)
- Identical structure/behavior to `leaders/index.blade.php`, with "Executive"/`executive_id` substituted for "Leader"/`leader_id` throughout (filter select label "Executive", options from `$executives`).

### `leader_code_logs/index.blade.php` — "Leader Code Logs" — `route('leader.code.logs')` (company, admin)
- **Layout/role:** `@extends('layouts.app')`.
- **Page structure:** H1 "Leader Code Logs" + "Back to Find Users" outline-secondary button (arrow-left icon) → `route('company.users')`. Filter form (GET): User text search (`search`, placeholder "Search by name or Signet ID"), From date, To date, "Search" submit. Table columns: **User** (name + `SIG-00{id}` in parens, or "N/A"), **Old Leader** (name or "None"), **New Leader** (name or "None"), **Changed By** (name or "N/A"), **Date** (`Y-m-d H:i`). Paginated, empty state "No leader changes found".

### `executive_code_logs/index.blade.php` — "Executive Code Logs" — `route('executive.code.logs')` (company, admin)
- Identical structure to leader_code_logs, with "Executive" substituted for "Leader" throughout (columns: User, Old Executive, New Executive, Changed By, Date).

### `earn/index.blade.php` — "Earn History" — `route('earn.history')` (any authenticated)
- **Layout/role:** `@extends('layouts.app')`.
- **Page structure:** Flash alerts, H1 "Earn History". Date-range filter (`date_from`/`date_to`, GET), "Filter"/"Clear". Table columns: **Amount** (green badge, `{amount} USDT` + optional ` -{description}` suffix), **Date** (info badge, raw timestamp). Paginated, empty state "No earnings found".

### `salaries/index.blade.php` — "Salaries" — `route('salaries.index')` (company)
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** Flash alerts, H1 "Salaries" + "Add Salary" primary button (plus icon) opening a modal. Date-range filter form (`date_from`/`date_to`), "Filter"/"Clear", plus a right-aligned "Total in range: {amount}" readout in the same form row. Table columns: **User**, **SIGNET ID**, **Amount**, **Date**, **Remarks** (or "-"). (A commented-out sixth "Actions"/delete column exists in source but is disabled.) Paginated, empty state "No salary records found".
- **Add Salary modal** (`#addSalaryModal`): User (Select2 AJAX-searchable dropdown hitting `route('salaries.searchUsers')`, placeholder "Search by name, Signet ID, or WhatsApp", `minimumInputLength: 1`), Amount (number, step 0.01, min 0, required), Date (date, required), Remarks (text, optional). "Save" submit.
- **Interactions:** modal reset + Select2 clear on `hidden.bs.modal`. Submit → `fetch POST route('salaries.store')`; SweetAlert2 "Saved!" then reload, or error alert. A `.deleteBtn` handler exists in script (SweetAlert2 confirm → `DELETE /salaries/{id}` → reload) but has no matching button currently rendered in the markup (dead code paired with the commented-out Actions column).

### `tokens/index.blade.php` — "Tokens for {name}" — `route('view.tokens', $userId)` (company)
- **Layout/role:** `@extends('layouts.app')`, company.
- **Page structure:** H2 "Tokens for {{ $user->name }}". Flash alerts. "Generate Tokens" form (half-width column): number of tokens (`token_count`, number, min 1, max 500, required), Google Auth Code (`google_auth_code`, number, min 1, required), "Generate Tokens" submit → `route('generate.tokens', $user->id)`. Table columns: **#**, **Token**, **Status** (badge green "Active"/red "Inactive"). Paginated.

### `tokens/share.blade.php` — "Share Token" — `route('token.share')` (admin, user, agent)
- **Layout/role:** `@extends('layouts.app')`, content-only card chrome (auth-page style, though wrapped inside the dashboard layout with sidebar/topbar included above it).
- **Page structure:** H1 "Share Token", "My Token : {{ $tokens }} USDT" line (note: markup has a malformed/unclosed `<p>...</a` tag in source — harmless in Blade/HTML but worth just fixing cleanly in the Vue version). Session error alert. Form → `route('token.shares')`: Tokens Value (number, `max=$tokens`, `min=1`, required, autofocus), User ID (number, required). "Send Tokens" submit (`btn btn-gray-800`).

### `tokens/share-log.blade.php` — "Tokens Share Log" — `route('token.share.log')` (admin, user, agent)
- **Layout/role:** `@extends('layouts.app')`.
- **Page structure:** Flash alerts, H1 "Tokens Share Log". Table columns: **Tokens** (green badge, `{amount} USDT`), **Receiver** (info badge, name or "Unknown User"), **Receiver Whats App** (info badge, number or "N/A"), **Date** (info badge, raw timestamp). Paginated.

### KYC pages (`kyc/create.blade.php`, `kyc/edit.blade.php`, `kyc/index.blade.php`, `kyc/show.blade.php`, `kyc/verified.blade.php`)

**`kyc/create.blade.php`** — "Submit Your KYC" — `route('kyc.create')` / `route('kyc.store')`
- Validation error summary box (`$errors->any()`) listing all messages.
- Form (`multipart/form-data`) fields: Full Name (text, required), Email (email, required), Contact Number 1 (text, required), Contact Number 2 (text, optional), Address (text, required), Telegrame User Name [sic — literal label spelling] (text, required), Document Type (select: NIC/Passport, required), Document Number (text, required); conditional file inputs shown/hidden by JS based on Document Type: NIC → "NIC Front Image" + "NIC Back Image" (both `type=file`); Passport → "Passport Image" (`type=file`). Submit "Submit KYC" (success button).

**`kyc/edit.blade.php`** — "Edit Your KYC" — `route('kyc.edit', $kyc)` / `route('kyc.update', $kyc)` (PUT)
- Same field set as create, all pre-filled from `$kyc` via `old(field, $kyc->field)`; file-upload labels append a "(View)" link to the existing uploaded file when present. Submit "Update KYC" (primary button).
- Both create/edit share the same `toggleFields()` JS pattern (show/hide NIC vs Passport file inputs based on the Document Type select).

**`kyc/index.blade.php`** — "KYC" (company: all records; others: own) — `route('kyc.index')`
- Flash alerts, H1 "KYC", "Verified KYC" success button (company only visually relevant, but rendered regardless) → `route('kyc.verified')`.
- Table (shown only if records exist): columns **Full Name**, **Email**, **WhatsApp 1**, **WhatsApp 2** (or "-"), **Address**, **Telegrame Username** (or "-"), **Document Type** (uppercased), **Document No**, **Documents** (icon links to uploaded files — id-card icon ×2 for NIC front/back, passport icon for passport, or "No document"), **Verified** (badge-success "Verified" / badge-warning "Pending"), **Actions**:
  - Non-company role: if not yet verified → "Edit" (primary) + "Delete" (danger, native `confirm()`) buttons; if verified → "N/A" badge (edit/delete locked once verified).
  - Company role: "Verify" (success submit) or "Unverify" (warning submit) form button depending on current state.
- Pagination shown only for company role. Empty state: non-company → "You have not submitted your KYC yet." + "Submit KYC" button; company → info alert "No KYC records found."

**`kyc/show.blade.php`** — "KYC" (card-grid variant, likely the "own KYC" view for non-company users) — `route('kyc.show')`
- Same header/flash pattern as index but renders KYC records as a 2-column card grid instead of a table: each record is a card with header (name), an info-alert body listing Email/WhatsApp 1/WhatsApp 2/Address/Telegram/Document-type+number with icons, a document-image preview strip (thumbnails, not just icon links, for NIC front/back or passport), a verified/pending badge, and the same role-based action buttons as `kyc/index.blade.php` (Edit/Delete for unverified own record, "Already Verified" text once verified; Verify/Unverify for company). 
- **Verified-record bonus panel** (only rendered when `$kyc->is_verified` is true): a "Telegrame Join Link" card containing a prominent danger-styled warning box about not sharing Telegram links publicly (long compliance copy — reproduce verbatim, it's a policy notice), followed by "Join Telegram" primary buttons (one per `$userPackages` entry, `https://t.me/{telegram_link}`, telegram icon).
- Empty state identical pattern to index.

**`kyc/verified.blade.php`** — "Verified KYC" — `route('kyc.verified')` (company)
- Structurally identical to `kyc/index.blade.php` (same table/columns/actions/pagination logic) — presumably the same list pre-filtered server-side to verified-only records.

### `geneology/index.blade.php` — "MY Genealogy" — `route('my.geneology')`
- **Layout/role:** `@extends('layouts.app')`, non-company roles.
- **Page structure:** H1 "MY Genealogy". A pure-CSS org-chart tree (custom `<style>` block in-page: connector lines via `::before`/`::after` borders, scrollable container up to 80vh) — root node = current user's name in a blue box; children = direct downline members rendered as colored link-boxes (`route('geneology.show', $childern->user->id)`), colored via a `getDynamicColor()` PHP helper keyed by position (orange `#FF5733` at fixed milestone indices/every-10th beyond 100, default blue `#3498db`), with a special yellow box override for any child whose `status === 'deactive'`. Children with `status === 'pending'` are skipped entirely (`@continue`).
- **Interactions:** none beyond navigation; SweetAlert2 script tag included but unused on this page.

### `geneology/show.blade.php` — individual downline member detail — `route('geneology.show', $userId)`
- **Layout/role:** `@extends('layouts.app')`.
- **Page structure:** Card with primary header showing the user's name. Body:
  - **User Information** table (bordered): Name, Signet ID (`SIG-00{id}`), Email, Mobile (or "N/A"), Registered At (`Y-m-d`).
  - **Referred By (Parent)** table (Name, Email) — only if `$parentData` present, else italic muted "No parent user found."
  - **User Packages** table (bordered/striped): **#**, **Package Name** (or "N/A"), **Earnings** (2-decimal), **Activated On** (`Y-m-d`) — only if `$userPackage->count()`, else italic muted "No packages assigned to this user."

---

## Shared components

### Layout shells
- **`layouts/app.blade.php`** — the dashboard/auth HTML shell. `<head>`: title "signetint", favicon set, SweetAlert2 CSS, Notyf CSS, `resources/css/volt.css` (Volt Bootstrap 5 theme) + `resources/css/custome.css` (project overrides), Font Awesome 5 via CDN. `<body>`: renders `@include('sweetalert::alert')` (server-flashed SweetAlert trigger), then `@yield('sidebar')`, `@yield('content')`, `@yield('footer')` in that order, then script includes (jQuery 3.6, Popper, Bootstrap JS, SweetAlert2 JS, `volt.js`, `custome.js`, Pusher JS), then `@yield('scripts')`.
- **`layouts/main.blade.php`** — the public marketing HTML shell (see Public Site section above for what it loads). Has its own header/nav and footer baked directly into the layout (not yielded sections) — only `@yield('content')` is used by `welcome.blade.php`.
- **`layouts/sidebar.blade.php`** — the app shell's left nav, `@include`d from every dashboard page's `@section('sidebar')`. Structure: a mobile top navbar (hamburger toggling `#sidebarMenu` on small screens) + the actual `<nav id="sidebarMenu">` sidebar (dark theme, `bg-gray-800 text-white`, `data-simplebar` scroll container). Logo links to `/dashboard`.
  - **Nav items and visibility (role-gated with `@if(Auth::user()->role == '...')`):**
    - **Dashboard** (home icon) → `/dashboard` — all roles, always shown, marked `active`.
    - *Company only:*
      - Google Auth Setup → `route('setup.google.auth', auth()->id())`
      - Countries → `route('countries.index')`
      - Packages → `route('packages.index')`
      - Mining Token → `route('mining.users')`
      - Users → `route('company.users')`
      - ROC Income → `route('company.roc')`
      - Global Director Share → `route('company.direct_share')`
      - GDS Log → `route('company.direct_share_log')`
      - Salaries → `route('salaries.index')`
      - Leaders Gain → `route('leaders.gain')`
      - Executives Gain → `route('executives.gain')`
      - Leader Code Logs → `route('leader.code.logs')`
      - LB Log → `route('company.leadership_bonus_log')`
      - Fake Accounts → `route('userparentlogs.index')`
    - *Admin only:*
      - Users → `route('company.users')`
      - Leader Code Logs → `route('leader.code.logs')`
    - *Admin, User, Agent:*
      - Token Share → `route('token.share')`
      - Token Share Log → `route('token.share.log')`
      - Top Up → `route('buy.package.history')`
      - Earn Log → `route('earn.history')`
    - *Company only (second block):*
      - My Activations → `route('admin.pending.activation')` if role is admin (dead branch — this block is inside an outer `@if role == 'company'`, so the inner admin check never actually fires; effectively always renders `route('company.pending.activation')` for company users) — label "My Activations".
      - Waiting Activations → `route('new.activations')`, with a live danger-pill badge showing `{{ newActivations() ?? 0 }}` (count of `company_status = 0` package rows).
    - *Everyone except company:*
      - Geneology → `route('my.geneology')`
    - *Everyone (role !== company shows `kyc.show`, company shows `kyc.index`):*
      - KYC → conditionally `route('kyc.show')` or `route('kyc.index')`.
    - Divider, then always: **Support** (bell/info icon, dead `javascript:void(0)` link, empty badge placeholder) and **Logout** (submits a hidden POST form to `route('logout')`).
  - All nav icons are hand-inlined heroicons-style SVGs except a couple that reuse one generic "box" icon path repeatedly (Packages/Mining/Users/ROC/etc. all share the identical box-icon SVG — a source shortcut, not a design intent; in the Vue rebuild consider distinct icons per item for clarity, or keep identical if pixel parity is required).
  - Note a rendering artifact at line ~300 of the source (garbled full-width Unicode digits inside an unused/duplicate SVG path fragment for the "Waiting Activations" icon) — appears to be a corrupted edit in the original repo; treat the icon as the same generic box/table SVG used elsewhere rather than reproducing the garbage characters.
- **`layouts/topbar.blade.php`** — `@include`d at the top of `@section('content')` on every dashboard page. A slim top navbar (`navbar-dashboard`) with no page title/breadcrumb — just a right-aligned user-avatar dropdown showing `{{ auth()->user()->name }}` and, in the dropdown menu, a single "Logout" item (danger-colored icon+text) that submits the same hidden logout form pattern as the sidebar.

### Flash messages / alerts
- Rendered inline per-page (no single shared partial observed) as `@if(session('success')) <div class="alert alert-success">...</div> @endif` and the equivalent for `error`/`status`, placed directly under the topbar include at the top of each page's content. A few pages (e.g. `company/user_parent_logs/index.blade.php`) use the dismissible variant (`alert-dismissible fade show` + a `.btn-close`). Validation errors on forms use Laravel's `$errors->any()` bag rendered as a red summary box (KYC create/edit) or per-field `@error()` `invalid-feedback` spans (password reset forms).
- **`@include('sweetalert::alert')`** in the layout head renders any server-side flashed SweetAlert (via the `realrashid/sweet-alert` package) as a auto-firing `Swal.fire()` on page load — used in addition to/instead of the inline Bootstrap alerts on some flows.
- Toast-style ephemeral notices (top-end, 1.5s auto-dismiss, pause-on-hover) are built ad hoc per page via `Swal.mixin({ toast: true, position: 'top-end', showConfirmButton: false, timer: 1500, timerProgressBar: true, ... })` — this exact pattern is repeated verbatim across ROC Income, Direct Share, Direct Share Log, Leadership Bonus Log, Users (leader-status toggle), and Leaders/Executives Gain (copy-to-clipboard) pages. Worth extracting into one shared Vue composable/util (`useToast()`), since it's copy-pasted 6+ times in the source with identical config.

### Tables (recurring pattern across nearly all list pages)
- `card border-0 shadow` (sometimes `shadow-sm`, `shadow mb-4`) wrapping a `table-responsive > table.table.align-items-center.table-flush`, header row `thead-light` with `border-bottom` on each `<th>`. Status/quantity values are almost always wrapped in `<span class="badge bg-*">` rather than plain text. Pagination via Laravel's default paginator (`{{ $collection->links() }}`) rendered centered (`d-flex justify-content-center`) below the table — Bootstrap-styled Laravel pagination component (prev/next + numbered pages).

### Badge/status color conventions observed (for exact 1:1 recreation)
| Context | Value | Class |
|---|---|---|
| Package/User "Active" | active | `bg-success` |
| Package/User "Inactive"/"Deactive" | inactive/deactive | `bg-danger` or `bg-secondary` (Package list uses `bg-secondary` for Deactive; user "Deactive" account status row in Fake Accounts uses plain `bg-secondary`) |
| ROC status | active | `bg-success` |
| ROC status | inactive | `bg-warning` |
| ROC status | other/unset | `bg-danger` |
| Global Director Share status | 1 | `bg-success` |
| Global Director Share status | 0 | `bg-danger` |
| KYC verified | true | `badge-success` "Verified" (note: plain `badge-success`, not Bootstrap's `bg-success` — a legacy Bootstrap 4-style class that likely relies on `custome.css` for color; verify visually against the live app since it won't be styled by stock Bootstrap 5) |
| KYC verified | false | `badge-warning` "Pending" (same legacy-class caveat) |
| Needed-tokens badge | any | `bg-success` |
| Amount/earn badges | any | `bg-success` (amount) / `bg-info` (dates, receiver info) |
| Mining status | active | `bg-success` |
| Mining status | inactive | `bg-danger` |
| Genealogy tree node | normal position | orange `#FF5733` (constant across all milestone indices — the "dynamic color" map currently always resolves to the same orange) |
| Genealogy tree node | status = deactive | yellow `#f1c40f`, black text |
| Genealogy tree node | root (self) | blue `#3498db`, white text |

### Modals (recurring Bootstrap pattern)
Every modal in the app follows the same shape: `modal fade` → `modal-dialog` → `modal-content` → `<form>` wrapping `modal-header` (title + `.btn-close`), `modal-body` (fields), `modal-footer` (secondary "Close"/Cancel button + primary submit). Triggered either declaratively (`data-bs-toggle="modal" data-bs-target="#id"`, often paired with an inline `onclick` to pre-populate title/fields — see Countries) or programmatically via `new bootstrap.Modal(...).show()` after a delegated click handler reads `data-*` attributes off the triggering button (Users page's 4 modals, Mining page's update modal, Direct Share's edit-pool modal).

### Helper-driven inline widgets (server-rendered HTML fragments used inside dashboards)
Two PHP helpers (`app/helpers.php`) inject raw HTML directly into Blade via `{!! ... !!}` on the Admin/Agent/User/Company dashboards — reproduce their *output shape* as real Vue components rather than raw HTML strings:
- **`rank($userId)`** → a flex row of labeled badges: "Team Sales" (`bg-info text-dark`), "Gratitude" (`bg-info text-dark`), "Current Rank" (`bg-success`), "Next Rank" (`bg-warning text-dark`), "Remaining Team" (`bg-danger`), "Remaining Gratitude" (`bg-danger`). Rank names ladder: Crystal → Jade → Emerald → Ruby → Diamond → Senior Diamond → Senior Executive Diamond → Crown Diamond, each requiring a team-sales threshold and a "gratitude/super" threshold.
- **`roc($userId)`** → (only rendered when the user's `roc_status == 'active'`) a line-break + flex row showing "Last Week Summary: {week_start} - {week_end}" and "Total sales per week: {amount}".
- **`allUsers()`** → integer count of active users with role `user`, used in headings like "All Users Count(N)" and the Mining card title.
- **`newActivations()`** → integer count of package rows pending company activation (`company_status = 0`), used only for the sidebar's "Waiting Activations" badge.
- **`walletBalance($userId)`** → a user's wallet balance, used throughout stat cards ("My Earn", wallet math).

### Mining "Community Staking Token" widget (recurring block: Admin, Agent, User dashboards)
A self-contained card+script block duplicated verbatim across the three role dashboards (admin/agent/user — company's dashboard does not include it): header with a live "Connecting…/Connected/Disconnected" badge, a 4-up stat row (Mining Token, Total Token, Daily Mining, Status badge), an animated striped progress bar with a live percentage label, a "N tokens/second" rate badge, and a "Last updated" timestamp — driven by polling `GET /mining/user/{userId}` on load and every 5s, with a client-side per-second local increment simulation between polls, and a SweetAlert2 "Mining Complete!" toast when the local counter reaches the total. This should become one shared `MiningWidget.vue` component in the rebuild, parameterized only by the current user id.

### "Active Package"/activation AJAX pattern (recurring across many pages)
Buttons with class `.active-package` and `data-id="{id}"` appear on: Admin dashboard, Admin pending-activation, Agent dashboard, User dashboard, Company new-activations, Company pending-activation, KYC index/show/verified (leftover script only, no matching button actually rendered on KYC pages), Buy-package history (leftover script only). The button POSTs `{_token, package_id}` to one of `/active-package` or `/company/new_active-package` depending on the page, shows a SweetAlert2 success/error dialog, and reloads the page on confirm. Consolidate into one composable/action in the rebuild, parameterized by target endpoint.
