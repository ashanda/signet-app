// Package models defines Go structs matching the live database schema
// exactly (see docs/analysis/schema.md) — including columns that are
// string-typed despite numeric-looking names (earn_logs.user_id/amount,
// user_packages.earn, global_share_wallets_log.user_id/amount). Those are
// kept as strings here on purpose: the task requires reading/writing the
// SAME existing database, and changing a column's Go-side type doesn't
// change what MySQL actually stores.
package models

import (
	"time"
)

type Country struct {
	ID        uint64   `db:"id" json:"id"`
	Code      string   `db:"code" json:"code"`
	Name      string   `db:"name" json:"name"`
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

type User struct {
	ID                        uint64     `db:"id" json:"id"`
	CountryID                 NullInt64  `db:"country_id" json:"country_id"`
	Name                      string     `db:"name" json:"name"`
	Email                     string     `db:"email" json:"email"`
	RocStatus                 NullString `db:"roc_status" json:"roc_status"`
	GlobalDirectorShare       float64    `db:"global_director_share" json:"global_director_share"`
	GlobalDirectorShareStatus bool       `db:"global_director_share_status" json:"global_director_share_status"`
	EmailVerifiedAt           NullTime   `db:"email_verified_at" json:"email_verified_at"`
	Password                  string     `db:"password" json:"-"`
	RememberToken             NullString `db:"remember_token" json:"-"`
	CreatedAt                 NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt                 NullTime   `db:"updated_at" json:"updated_at"`
	Role                      string     `db:"role" json:"role"`
	WhatsappNumber            NullString `db:"whatsapp_number" json:"whatsapp_number"`
	BinancePayID              NullString `db:"binance_pay_id" json:"binance_pay_id"`
	Status                    string     `db:"status" json:"status"` // pending|active|inactive
	ReferredBy                NullInt64  `db:"referred_by" json:"referred_by"`
	LeaderCode                NullString `db:"leader_code" json:"leader_code"`
	LeaderStatus              string     `db:"leader_status" json:"leader_status"`
	ExecutiveCode             NullString `db:"executive_code" json:"executive_code"`
	GoogleAuthenticatorSecret NullString `db:"google_authenticator_secret" json:"-"`
	OnVacation                bool       `db:"on_vacation" json:"on_vacation"`
}

// SignetID formats a user id the same way the original app does throughout
// (MiningController, UserController, SalaryController): the literal prefix
// "SIG-00" concatenated directly with the raw id — NOT zero-padded to a
// fixed width (id 5 -> "SIG-005", id 123 -> "SIG-00123").
func SignetID(id uint64) string {
	return "SIG-00" + itoa(id)
}

func itoa(id uint64) string {
	if id == 0 {
		return "0"
	}
	digits := []byte{}
	for id > 0 {
		digits = append([]byte{byte('0' + id%10)}, digits...)
		id /= 10
	}
	return string(digits)
}

type ReferralCode struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	Code      string   `db:"code" json:"code"`
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

type Package struct {
	ID           uint64     `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	Price        int64      `db:"price" json:"price"`
	Commission   int64      `db:"commission" json:"commission"`
	Status       string     `db:"status" json:"status"`
	CreatedAt    NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt    NullTime   `db:"updated_at" json:"updated_at"`
	Rank         NullString `db:"rank" json:"rank"`
	TelegramLink NullString `db:"telegram_link" json:"telegram_link"`
}

type UserPackage struct {
	ID            uint64     `db:"id" json:"id"`
	UserID        uint64     `db:"user_id" json:"user_id"`
	Package       string     `db:"package" json:"package"` // package id, stored as string per schema
	RefID         NullString `db:"ref_id" json:"ref_id"`
	Status        string     `db:"status" json:"status"` // pending|active|deactivate
	CompanyStatus int        `db:"company_status" json:"company_status"`
	Earn          string     `db:"earn" json:"earn"` // string column, see package doc
	CreatedAt     NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime   `db:"updated_at" json:"updated_at"`
	ActivatedAt   NullTime   `db:"activated_at" json:"activated_at"`
	Sale          string     `db:"sale" json:"sale"` // first|other
}

type Token struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	Token     string   `db:"token" json:"token"`
	Status    string   `db:"status" json:"status"` // active|deactive
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

type Wallet struct {
	ID         uint64   `db:"id" json:"id"`
	UserID     uint64   `db:"user_id" json:"user_id"`
	Balance    float64  `db:"balance" json:"balance"`
	RocBalance float64  `db:"roc_balance" json:"roc_balance"`
	CreatedAt  NullTime `db:"created_at" json:"created_at"`
	UpdatedAt  NullTime `db:"updated_at" json:"updated_at"`
}

type UserParent struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	VirtualID uint64   `db:"virtual_id" json:"virtual_id"`
	ParentID  uint64   `db:"parent_id" json:"parent_id"`
	Node      string   `db:"node" json:"node"` // active|deactive|gratitude|correct
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

// EarnLog: user_id and amount are VARCHAR columns in the live schema (see
// schema.md Note 6) — kept as strings here, parsed at the point of use.
type EarnLog struct {
	ID          uint64     `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"user_id"`
	Amount      string     `db:"amount" json:"amount"`
	Description NullString `db:"description" json:"description"`
	CreatedAt   NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt   NullTime   `db:"updated_at" json:"updated_at"`
}

type TokenLog struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	SharedBy  uint64   `db:"shared_by" json:"shared_by"`
	Amount    float64  `db:"amount" json:"amount"`
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

type Kyc struct {
	ID               uint64     `db:"id" json:"id"`
	UserID           uint64     `db:"user_id" json:"user_id"`
	FullName         string     `db:"full_name" json:"full_name"`
	Email            string     `db:"email" json:"email"`
	ContactNumber1   string     `db:"contact_number1" json:"contact_number1"`
	ContactNumber2   NullString `db:"contact_number2" json:"contact_number2"`
	Address          string     `db:"address" json:"address"`
	TelegramUsername string     `db:"telegram_username" json:"telegram_username"`
	DocumentType     string     `db:"document_type" json:"document_type"` // nic|passport
	DocumentNumber   string     `db:"document_number" json:"document_number"`
	NicFront         NullString `db:"nic_front" json:"nic_front"`
	NicBack          NullString `db:"nic_back" json:"nic_back"`
	PassportImage    NullString `db:"passport_image" json:"passport_image"`
	IsVerified       bool       `db:"is_verified" json:"is_verified"`
	Comments         NullString `db:"comments" json:"comments"`
	CreatedAt        NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt        NullTime   `db:"updated_at" json:"updated_at"`
}

type ApiUser struct {
	ID            uint64     `db:"id" json:"id"`
	Username      string     `db:"username" json:"username"`
	Password      string     `db:"password" json:"-"`
	RememberToken NullString `db:"remember_token" json:"-"`
	CreatedAt     NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime   `db:"updated_at" json:"updated_at"`
}

type PersonalAccessToken struct {
	ID            uint64     `db:"id" json:"id"`
	TokenableType string     `db:"tokenable_type" json:"tokenable_type"`
	TokenableID   uint64     `db:"tokenable_id" json:"tokenable_id"`
	Name          string     `db:"name" json:"name"`
	Token         string     `db:"token" json:"-"` // sha256 hex digest
	Abilities     NullString `db:"abilities" json:"abilities"`
	LastUsedAt    NullTime   `db:"last_used_at" json:"last_used_at"`
	ExpiresAt     NullTime   `db:"expires_at" json:"expires_at"`
	CreatedAt     NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime   `db:"updated_at" json:"updated_at"`
}

type SuperParentLog struct {
	ID            uint64    `db:"id" json:"id"`
	CurrentUserID uint64    `db:"current_user_id" json:"current_user_id"`
	PackageID     uint64    `db:"package_id" json:"package_id"`
	GratitudeUser int64     `db:"gratitude_user" json:"gratitude_user"`
	UserPackage   NullInt64 `db:"user_package" json:"user_package"`
	CreatedAt     NullTime  `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime  `db:"updated_at" json:"updated_at"`
}

type WeeklyPackageSummary struct {
	ID             uint64    `db:"id" json:"id"`
	JobID          string    `db:"job_id" json:"job_id"`
	WeekStart      time.Time `db:"week_start" json:"week_start"`
	WeekEnd        time.Time `db:"week_end" json:"week_end"`
	PerWeekTotal   float64   `db:"per_week_total" json:"per_week_total"`
	BalanceForward float64   `db:"balance_forward" json:"balance_forward"`
	RocPotation    uint64    `db:"roc_potation" json:"roc_potation"`
	TotalAmount    float64   `db:"total_amount" json:"total_amount"`
	CreatedAt      NullTime  `db:"created_at" json:"created_at"`
	UpdatedAt      NullTime  `db:"updated_at" json:"updated_at"`
}

type RocIncomeLog struct {
	ID          uint64   `db:"id" json:"id"`
	JobID       string   `db:"job_id" json:"job_id"`
	UserID      uint64   `db:"user_id" json:"user_id"`
	Amount      float64  `db:"amount" json:"amount"`
	Description string   `db:"description" json:"description"`
	Status      string   `db:"status" json:"status"` // pending|paid|active (convention only, not a DB enum)
	EarnLogID   uint64   `db:"earn_log_id" json:"earn_log_id"`
	CreatedAt   NullTime `db:"created_at" json:"created_at"`
	UpdatedAt   NullTime `db:"updated_at" json:"updated_at"`
}

type UserParentMapsLog struct {
	ID            uint64     `db:"id" json:"id"`
	UserID        NullInt64  `db:"user_id" json:"user_id"`
	ParentID      NullInt64  `db:"parent_id" json:"parent_id"`
	CreatedRowIDs NullString `db:"created_row_ids" json:"created_row_ids"` // JSON array, decoded by caller
	Note          NullString `db:"note" json:"note"`
	CreatedAt     NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime   `db:"updated_at" json:"updated_at"`
}

type UserSecretKey struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	SecretKey string   `db:"secret_key" json:"-"` // bcrypt hash, used as an opaque external id
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

type UserMining struct {
	ID              uint64     `db:"id" json:"id"`
	UserID          uint64     `db:"user_id" json:"user_id"`
	TotalToken      int64      `db:"total_token" json:"total_token"`
	MiningToken     float64    `db:"mining_token" json:"mining_token"`
	DailyMining     int64      `db:"daily_mining" json:"daily_mining"`
	Status          string     `db:"status" json:"status"` // active|inactive
	CreatedAt       NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt       NullTime   `db:"updated_at" json:"updated_at"`
	WebhookSentAt   NullTime   `db:"webhook_sent_at" json:"webhook_sent_at"`
	WebhookStatus   string     `db:"webhook_status" json:"webhook_status"`
	WebhookAttempts int        `db:"webhook_attempts" json:"webhook_attempts"`
	NextRetryAt     NullTime   `db:"next_retry_at" json:"next_retry_at"`
	WebhookResponse NullString `db:"webhook_response" json:"webhook_response"`
}

type PackagePool struct {
	ID         uint64   `db:"id" json:"id"`
	UserID     uint64   `db:"user_id" json:"user_id"`
	PackageID  uint64   `db:"package_id" json:"package_id"`
	PoolAmount float64  `db:"pool_amount" json:"pool_amount"`
	CreatedAt  NullTime `db:"created_at" json:"created_at"`
	UpdatedAt  NullTime `db:"updated_at" json:"updated_at"`
}

type GlobalShareWallet struct {
	ID        uint64   `db:"id" json:"id"`
	UserID    uint64   `db:"user_id" json:"user_id"`
	Balance   float64  `db:"balance" json:"balance"`
	MaxOut    float64  `db:"max_out" json:"max_out"`
	CreatedAt NullTime `db:"created_at" json:"created_at"`
	UpdatedAt NullTime `db:"updated_at" json:"updated_at"`
}

// GlobalShareWalletLog: user_id and amount are VARCHAR columns (see
// schema.md Note 6), matching earn_logs' pattern.
type GlobalShareWalletLog struct {
	ID          uint64     `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"user_id"`
	Amount      string     `db:"amount" json:"amount"`
	Description NullString `db:"description" json:"description"`
	CreatedAt   NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt   NullTime   `db:"updated_at" json:"updated_at"`
}

type Salary struct {
	ID         uint64     `db:"id" json:"id"`
	UserID     uint64     `db:"user_id" json:"user_id"`
	Amount     float64    `db:"amount" json:"amount"`
	SalaryDate time.Time  `db:"salary_date" json:"salary_date"`
	Remarks    NullString `db:"remarks" json:"remarks"`
	CreatedAt  NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt  NullTime   `db:"updated_at" json:"updated_at"`
}

type GlobalDirectorShareDistribution struct {
	ID            uint64   `db:"id" json:"id"`
	Period        string   `db:"period" json:"period"`
	TotalPool     float64  `db:"total_pool" json:"total_pool"`
	TotalShares   float64  `db:"total_shares" json:"total_shares"`
	ValuePerShare float64  `db:"value_per_share" json:"value_per_share"`
	UserCount     uint64   `db:"user_count" json:"user_count"`
	CreatedAt     NullTime `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime `db:"updated_at" json:"updated_at"`
}

type LeaderCodeLog struct {
	ID            uint64     `db:"id" json:"id"`
	UserID        uint64     `db:"user_id" json:"user_id"`
	OldLeaderCode NullString `db:"old_leader_code" json:"old_leader_code"`
	NewLeaderCode NullString `db:"new_leader_code" json:"new_leader_code"`
	ChangedBy     NullInt64  `db:"changed_by" json:"changed_by"`
	CreatedAt     NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt     NullTime   `db:"updated_at" json:"updated_at"`
}

type ExecutiveCodeLog struct {
	ID               uint64     `db:"id" json:"id"`
	UserID           uint64     `db:"user_id" json:"user_id"`
	OldExecutiveCode NullString `db:"old_executive_code" json:"old_executive_code"`
	NewExecutiveCode NullString `db:"new_executive_code" json:"new_executive_code"`
	ChangedBy        NullInt64  `db:"changed_by" json:"changed_by"`
	CreatedAt        NullTime   `db:"created_at" json:"created_at"`
	UpdatedAt        NullTime   `db:"updated_at" json:"updated_at"`
}
