package models

import "time"

type User struct {
	ID             int64     `json:"id"`
	TelegramUserID int64     `json:"telegramUserId"`
	Username       string    `json:"username"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	DisplayName    string    `json:"displayName"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type EventStatus string

const (
	EventActive   EventStatus = "active"
	EventArchived EventStatus = "archived"
	EventDeleted  EventStatus = "deleted"
)

type Event struct {
	ID                      int64       `json:"id"`
	Title                   string      `json:"title"`
	CreatorUserID           int64       `json:"creatorUserId"`
	CollectorUserID         *int64      `json:"collectorUserId"`
	CollectorName           string      `json:"collectorName"`
	CollectorPhone          string      `json:"collectorPhone"`
	Currency                string      `json:"currency"`
	JoinCode                string      `json:"joinCode"`
	Status                  EventStatus `json:"status"`
	AllowMembersAddExpenses bool        `json:"allowMembersAddExpenses"`
	CreatedAt               time.Time   `json:"createdAt"`
	UpdatedAt               time.Time   `json:"updatedAt"`
	ArchivedAt              *time.Time  `json:"archivedAt"`
}

type MemberRole string

const (
	RoleCreator MemberRole = "creator"
	RoleMember  MemberRole = "member"
)

type PaymentStatus string

const (
	PaymentUnpaid PaymentStatus = "unpaid"
	PaymentPaid   PaymentStatus = "paid"
)

type EventMember struct {
	EventID       int64         `json:"eventId"`
	UserID        int64         `json:"userId"`
	Name          string        `json:"name"`
	Username      string        `json:"username"`
	Role          MemberRole    `json:"role"`
	Emoji         string        `json:"emoji"`
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	JoinedAt      time.Time     `json:"joinedAt"`
}

type Expense struct {
	ID              int64                `json:"id"`
	EventID         int64                `json:"eventId"`
	Title           string               `json:"title"`
	AmountMinor     int64                `json:"amountMinor"`
	PaidByUserID    int64                `json:"paidByUserId"`
	PaidByName      string               `json:"paidByName"`
	PaidByEmoji     string               `json:"paidByEmoji"`
	CreatedByUserID int64                `json:"createdByUserId"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	Participants    []ExpenseParticipant `json:"participants"`
}

type ExpenseParticipant struct {
	UserID     int64  `json:"userId"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
	ShareMinor int64  `json:"shareMinor"`
}

type MemberSummary struct {
	UserID        int64         `json:"userId"`
	Name          string        `json:"name"`
	Emoji         string        `json:"emoji"`
	Role          MemberRole    `json:"role"`
	AmountMinor   int64         `json:"amountMinor"`
	PaymentStatus PaymentStatus `json:"paymentStatus"`
	IsCollector   bool          `json:"isCollector"`
}

type CollectorInfo struct {
	UserID int64  `json:"userId"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
}

type EventSummary struct {
	EventID                 int64          `json:"eventId"`
	Title                   string         `json:"title"`
	Currency                string         `json:"currency"`
	Status                  EventStatus    `json:"status"`
	TotalAmountMinor        int64          `json:"totalAmountMinor"`
	CurrentUserAmountMinor  int64          `json:"currentUserAmountMinor"`
	CurrentUserPaymentStatus PaymentStatus `json:"currentUserPaymentStatus"`
	CurrentUserRole         MemberRole     `json:"currentUserRole"`
	AllowMembersAddExpenses bool           `json:"allowMembersAddExpenses"`
	Collector               CollectorInfo  `json:"collector"`
	Members                 []MemberSummary `json:"members"`
	PaidAmountMinor         int64          `json:"paidAmountMinor"`
	RequiredAmountMinor     int64          `json:"requiredAmountMinor"`
	PaidCount               int            `json:"paidCount"`
	MembersCount            int            `json:"membersCount"`
}
