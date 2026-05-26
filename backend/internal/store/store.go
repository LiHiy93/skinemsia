package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"skinemsia/internal/models"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")
var ErrConflict = errors.New("conflict")

var emojiList = []string{
	"🟢", "🔵", "🟣", "🟡", "🔴", "🟠", "⚫", "⚪", "🟤", "🟦", "🟥", "🟨", "🟩", "🟪",
}

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (s *Store) UpsertUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*models.User, error) {
	display := firstName
	if lastName != "" {
		display += " " + lastName
	}
	if display == "" {
		display = username
	}
	if display == "" {
		display = fmt.Sprintf("user%d", telegramID)
	}

	var u models.User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (telegram_user_id, username, first_name, last_name, display_name, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (telegram_user_id) DO UPDATE
			SET username     = EXCLUDED.username,
			    first_name   = EXCLUDED.first_name,
			    last_name    = EXCLUDED.last_name,
			    display_name = EXCLUDED.display_name,
			    updated_at   = NOW()
		RETURNING id, telegram_user_id, username, first_name, last_name, display_name, created_at, updated_at
	`, telegramID, username, firstName, lastName, display).Scan(
		&u.ID, &u.TelegramUserID, &u.Username, &u.FirstName, &u.LastName, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	return &u, err
}

func (s *Store) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, telegram_user_id, username, first_name, last_name, display_name, created_at, updated_at
		FROM users WHERE telegram_user_id = $1
	`, telegramID).Scan(
		&u.ID, &u.TelegramUserID, &u.Username, &u.FirstName, &u.LastName, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

// ── Events ────────────────────────────────────────────────────────────────────

type CreateEventInput struct {
	Title          string
	CollectorName  string
	CollectorPhone string
	Currency       string
}

func (s *Store) CreateEvent(ctx context.Context, creatorID int64, in CreateEventInput) (*models.Event, error) {
	if in.Currency == "" {
		in.Currency = "RUB"
	}
	code := generateCode()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var e models.Event
	err = tx.QueryRow(ctx, `
		INSERT INTO events
			(title, creator_user_id, collector_user_id, collector_name, collector_phone, currency, join_code)
		VALUES ($1, $2, $2, $3, $4, $5, $6)
		RETURNING id, title, creator_user_id, collector_user_id, collector_name, collector_phone,
		          currency, join_code, status, allow_members_add_expenses, created_at, updated_at, archived_at
	`, in.Title, creatorID, in.CollectorName, in.CollectorPhone, in.Currency, code).Scan(
		&e.ID, &e.Title, &e.CreatorUserID, &e.CollectorUserID, &e.CollectorName, &e.CollectorPhone,
		&e.Currency, &e.JoinCode, &e.Status, &e.AllowMembersAddExpenses, &e.CreatedAt, &e.UpdatedAt, &e.ArchivedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO event_members (event_id, user_id, role, emoji, payment_status)
		VALUES ($1, $2, 'creator', $3, 'unpaid')
	`, e.ID, creatorID, emojiList[0])
	if err != nil {
		return nil, err
	}

	return &e, tx.Commit(ctx)
}

func (s *Store) GetEvent(ctx context.Context, eventID, userID int64) (*models.Event, error) {
	var e models.Event
	err := s.pool.QueryRow(ctx, `
		SELECT e.id, e.title, e.creator_user_id, e.collector_user_id, e.collector_name, e.collector_phone,
		       e.currency, e.join_code, e.status, e.allow_members_add_expenses,
		       e.created_at, e.updated_at, e.archived_at
		FROM events e
		JOIN event_members em ON em.event_id = e.id AND em.user_id = $2
		WHERE e.id = $1 AND e.status != 'deleted'
	`, eventID, userID).Scan(
		&e.ID, &e.Title, &e.CreatorUserID, &e.CollectorUserID, &e.CollectorName, &e.CollectorPhone,
		&e.Currency, &e.JoinCode, &e.Status, &e.AllowMembersAddExpenses,
		&e.CreatedAt, &e.UpdatedAt, &e.ArchivedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &e, err
}

func (s *Store) ListUserEvents(ctx context.Context, userID int64) ([]*models.Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.title, e.creator_user_id, e.collector_user_id, e.collector_name, e.collector_phone,
		       e.currency, e.join_code, e.status, e.allow_members_add_expenses,
		       e.created_at, e.updated_at, e.archived_at
		FROM events e
		JOIN event_members em ON em.event_id = e.id AND em.user_id = $1
		WHERE e.status != 'deleted'
		ORDER BY e.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.CreatorUserID, &e.CollectorUserID, &e.CollectorName, &e.CollectorPhone,
			&e.Currency, &e.JoinCode, &e.Status, &e.AllowMembersAddExpenses,
			&e.CreatedAt, &e.UpdatedAt, &e.ArchivedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	return list, rows.Err()
}

type UpdateEventInput struct {
	Title                   *string
	CollectorName           *string
	CollectorPhone          *string
	Currency                *string
	AllowMembersAddExpenses *bool
}

func (s *Store) UpdateEvent(ctx context.Context, eventID, userID int64, in UpdateEventInput) (*models.Event, error) {
	e, err := s.GetEvent(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if e.CreatorUserID != userID {
		return nil, ErrForbidden
	}

	if in.Title != nil {
		e.Title = *in.Title
	}
	if in.CollectorName != nil {
		e.CollectorName = *in.CollectorName
	}
	if in.CollectorPhone != nil {
		e.CollectorPhone = *in.CollectorPhone
	}
	if in.Currency != nil {
		e.Currency = *in.Currency
	}
	if in.AllowMembersAddExpenses != nil {
		e.AllowMembersAddExpenses = *in.AllowMembersAddExpenses
	}

	var updated models.Event
	err = s.pool.QueryRow(ctx, `
		UPDATE events
		SET title = $1, collector_name = $2, collector_phone = $3, currency = $4,
		    allow_members_add_expenses = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, title, creator_user_id, collector_user_id, collector_name, collector_phone,
		          currency, join_code, status, allow_members_add_expenses, created_at, updated_at, archived_at
	`, e.Title, e.CollectorName, e.CollectorPhone, e.Currency, e.AllowMembersAddExpenses, eventID).Scan(
		&updated.ID, &updated.Title, &updated.CreatorUserID, &updated.CollectorUserID,
		&updated.CollectorName, &updated.CollectorPhone, &updated.Currency, &updated.JoinCode,
		&updated.Status, &updated.AllowMembersAddExpenses, &updated.CreatedAt, &updated.UpdatedAt, &updated.ArchivedAt,
	)
	return &updated, err
}

func (s *Store) ArchiveEvent(ctx context.Context, eventID, userID int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE events SET status = 'archived', archived_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND creator_user_id = $2 AND status = 'active'
	`, eventID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}

func (s *Store) DeleteEvent(ctx context.Context, eventID, userID int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE events SET status = 'deleted', deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND creator_user_id = $2 AND status != 'deleted'
	`, eventID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}

func (s *Store) GetEventByCode(ctx context.Context, code string) (*models.Event, error) {
	var e models.Event
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, creator_user_id, collector_user_id, collector_name, collector_phone,
		       currency, join_code, status, allow_members_add_expenses, created_at, updated_at, archived_at
		FROM events
		WHERE join_code = $1 AND status = 'active'
	`, code).Scan(
		&e.ID, &e.Title, &e.CreatorUserID, &e.CollectorUserID, &e.CollectorName, &e.CollectorPhone,
		&e.Currency, &e.JoinCode, &e.Status, &e.AllowMembersAddExpenses,
		&e.CreatedAt, &e.UpdatedAt, &e.ArchivedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &e, err
}

func (s *Store) JoinEvent(ctx context.Context, code string, userID int64) (*models.Event, error) {
	e, err := s.GetEventByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// already a member?
	var cnt int
	_ = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM event_members WHERE event_id=$1 AND user_id=$2`, e.ID, userID,
	).Scan(&cnt)
	if cnt > 0 {
		return e, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// pick next emoji
	var memberCount int
	_ = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM event_members WHERE event_id=$1`, e.ID,
	).Scan(&memberCount)
	emoji := emojiList[memberCount%len(emojiList)]

	_, err = tx.Exec(ctx, `
		INSERT INTO event_members (event_id, user_id, role, emoji, payment_status)
		VALUES ($1, $2, 'member', $3, 'unpaid')
	`, e.ID, userID, emoji)
	if err != nil {
		return nil, err
	}

	// Auto-add new member to all existing expenses, recalculating shares
	expRows, err := tx.Query(ctx, `SELECT id, amount_minor FROM expenses WHERE event_id=$1`, e.ID)
	if err != nil {
		return nil, err
	}
	type expRow struct{ id, amount int64 }
	var exps []expRow
	for expRows.Next() {
		var r expRow
		if err := expRows.Scan(&r.id, &r.amount); err != nil {
			expRows.Close()
			return nil, err
		}
		exps = append(exps, r)
	}
	expRows.Close()

	for _, ex := range exps {
		partRows, err := tx.Query(ctx, `SELECT user_id FROM expense_participants WHERE expense_id=$1`, ex.id)
		if err != nil {
			return nil, err
		}
		var pids []int64
		for partRows.Next() {
			var uid int64
			if err := partRows.Scan(&uid); err != nil {
				partRows.Close()
				return nil, err
			}
			pids = append(pids, uid)
		}
		partRows.Close()

		pids = append(pids, userID)
		shares := calculateShares(ex.amount, pids)

		if _, err := tx.Exec(ctx, `DELETE FROM expense_participants WHERE expense_id=$1`, ex.id); err != nil {
			return nil, err
		}
		for _, sh := range shares {
			if _, err := tx.Exec(ctx, `
				INSERT INTO expense_participants (expense_id, user_id, share_minor) VALUES ($1,$2,$3)
			`, ex.id, sh.userID, sh.shareMinor); err != nil {
				return nil, err
			}
		}
	}

	return e, tx.Commit(ctx)
}

// ── Members ───────────────────────────────────────────────────────────────────

func (s *Store) GetMembers(ctx context.Context, eventID int64) ([]*models.EventMember, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT em.event_id, em.user_id, u.display_name, u.username,
		       em.role, em.emoji, em.payment_status, em.joined_at
		FROM event_members em
		JOIN users u ON u.id = em.user_id
		WHERE em.event_id = $1
		ORDER BY em.joined_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.EventMember
	for rows.Next() {
		var m models.EventMember
		if err := rows.Scan(
			&m.EventID, &m.UserID, &m.Name, &m.Username,
			&m.Role, &m.Emoji, &m.PaymentStatus, &m.JoinedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (s *Store) RemoveMember(ctx context.Context, eventID, targetUserID, requesterID int64) error {
	// only creator can remove
	var creatorID int64
	err := s.pool.QueryRow(ctx,
		`SELECT creator_user_id FROM events WHERE id=$1 AND status!='deleted'`, eventID,
	).Scan(&creatorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if creatorID != requesterID {
		return ErrForbidden
	}
	if targetUserID == creatorID {
		return errors.New("cannot remove event creator")
	}

	// check if target participates in expenses
	var expCnt int
	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM expense_participants ep
		JOIN expenses ex ON ex.id = ep.expense_id
		WHERE ex.event_id = $1 AND ep.user_id = $2
	`, eventID, targetUserID).Scan(&expCnt)
	if expCnt > 0 {
		return errors.New("cannot remove member who participates in expenses")
	}

	_, err = s.pool.Exec(ctx,
		`DELETE FROM event_members WHERE event_id=$1 AND user_id=$2`, eventID, targetUserID,
	)
	return err
}

func (s *Store) UpdateMemberEmoji(ctx context.Context, eventID, targetUserID, requesterID int64, emoji string) error {
	var creatorID int64
	err := s.pool.QueryRow(ctx,
		`SELECT creator_user_id FROM events WHERE id=$1 AND status!='deleted'`, eventID,
	).Scan(&creatorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if creatorID != requesterID {
		return ErrForbidden
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE event_members SET emoji=$1 WHERE event_id=$2 AND user_id=$3`, emoji, eventID, targetUserID,
	)
	return err
}

// ── Expenses ──────────────────────────────────────────────────────────────────

type CreateExpenseInput struct {
	Title          string
	AmountMinor    int64
	PaidByUserID   int64
	ParticipantIDs []int64
}

func (s *Store) ListExpenses(ctx context.Context, eventID int64) ([]*models.Expense, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ex.id, ex.event_id, ex.title, ex.amount_minor,
		       ex.paid_by_user_id, pb.display_name, pbm.emoji,
		       ex.created_by_user_id, ex.created_at, ex.updated_at
		FROM expenses ex
		JOIN users pb ON pb.id = ex.paid_by_user_id
		LEFT JOIN event_members pbm ON pbm.event_id = ex.event_id AND pbm.user_id = ex.paid_by_user_id
		WHERE ex.event_id = $1
		ORDER BY ex.created_at DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(
			&e.ID, &e.EventID, &e.Title, &e.AmountMinor,
			&e.PaidByUserID, &e.PaidByName, &e.PaidByEmoji,
			&e.CreatedByUserID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// load participants for each expense
	for _, ex := range list {
		parts, err := s.getParticipants(ctx, ex.ID)
		if err != nil {
			return nil, err
		}
		ex.Participants = parts
	}
	return list, nil
}

func (s *Store) GetExpense(ctx context.Context, expenseID, eventID int64) (*models.Expense, error) {
	var e models.Expense
	err := s.pool.QueryRow(ctx, `
		SELECT ex.id, ex.event_id, ex.title, ex.amount_minor,
		       ex.paid_by_user_id, pb.display_name, COALESCE(pbm.emoji,''),
		       ex.created_by_user_id, ex.created_at, ex.updated_at
		FROM expenses ex
		JOIN users pb ON pb.id = ex.paid_by_user_id
		LEFT JOIN event_members pbm ON pbm.event_id = ex.event_id AND pbm.user_id = ex.paid_by_user_id
		WHERE ex.id = $1 AND ex.event_id = $2
	`, expenseID, eventID).Scan(
		&e.ID, &e.EventID, &e.Title, &e.AmountMinor,
		&e.PaidByUserID, &e.PaidByName, &e.PaidByEmoji,
		&e.CreatedByUserID, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	parts, err := s.getParticipants(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Participants = parts
	return &e, nil
}

func (s *Store) CreateExpense(ctx context.Context, eventID, userID int64, in CreateExpenseInput) (*models.Expense, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var expID int64
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO expenses (event_id, title, amount_minor, paid_by_user_id, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, eventID, in.Title, in.AmountMinor, in.PaidByUserID, userID).Scan(&expID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	shares := calculateShares(in.AmountMinor, in.ParticipantIDs)
	for _, sh := range shares {
		if _, err := tx.Exec(ctx, `
			INSERT INTO expense_participants (expense_id, user_id, share_minor)
			VALUES ($1, $2, $3)
		`, expID, sh.userID, sh.shareMinor); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetExpense(ctx, expID, eventID)
}

func (s *Store) UpdateExpense(ctx context.Context, expenseID, eventID, userID int64, in CreateExpenseInput) (*models.Expense, error) {
	// check expense exists in this event
	var creatorID int64
	err := s.pool.QueryRow(ctx,
		`SELECT created_by_user_id FROM expenses WHERE id=$1 AND event_id=$2`, expenseID, eventID,
	).Scan(&creatorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// check event creator, expense creator, or allowMembersAddExpenses
	var eventCreatorID int64
	var allowMembers bool
	_ = s.pool.QueryRow(ctx, `SELECT creator_user_id, allow_members_add_expenses FROM events WHERE id=$1`, eventID).Scan(&eventCreatorID, &allowMembers)
	if userID != creatorID && userID != eventCreatorID && !allowMembers {
		return nil, ErrForbidden
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE expenses SET title=$1, amount_minor=$2, paid_by_user_id=$3, updated_at=NOW()
		WHERE id=$4
	`, in.Title, in.AmountMinor, in.PaidByUserID, expenseID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM expense_participants WHERE expense_id=$1`, expenseID)
	if err != nil {
		return nil, err
	}

	for _, sh := range calculateShares(in.AmountMinor, in.ParticipantIDs) {
		if _, err := tx.Exec(ctx, `
			INSERT INTO expense_participants (expense_id, user_id, share_minor) VALUES ($1,$2,$3)
		`, expenseID, sh.userID, sh.shareMinor); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetExpense(ctx, expenseID, eventID)
}

func (s *Store) DeleteExpense(ctx context.Context, expenseID, eventID, userID int64) error {
	var creatorID, eventCreatorID int64
	_ = s.pool.QueryRow(ctx, `SELECT created_by_user_id FROM expenses WHERE id=$1 AND event_id=$2`, expenseID, eventID).Scan(&creatorID)
	_ = s.pool.QueryRow(ctx, `SELECT creator_user_id FROM events WHERE id=$1`, eventID).Scan(&eventCreatorID)
	if userID != creatorID && userID != eventCreatorID {
		return ErrForbidden
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM expenses WHERE id=$1 AND event_id=$2`, expenseID, eventID)
	return err
}

// ── Payment ───────────────────────────────────────────────────────────────────

func (s *Store) SetPaymentStatus(ctx context.Context, eventID, userID int64, status models.PaymentStatus) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE event_members SET payment_status=$1
		WHERE event_id=$2 AND user_id=$3
	`, string(status), eventID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Summary ───────────────────────────────────────────────────────────────────

func (s *Store) GetSummary(ctx context.Context, eventID, currentUserID int64) (*models.EventSummary, error) {
	e, err := s.GetEvent(ctx, eventID, currentUserID)
	if err != nil {
		return nil, err
	}

	// total amount
	var total int64
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_minor),0) FROM expenses WHERE event_id=$1`, eventID,
	).Scan(&total)

	// per-member amounts — join expenses first to scope to this event only
	rows, err := s.pool.Query(ctx, `
		SELECT em.user_id, u.display_name, em.emoji, em.role, em.payment_status,
		       COALESCE(SUM(ep.share_minor),0) AS amount
		FROM event_members em
		JOIN users u ON u.id = em.user_id
		LEFT JOIN expenses ex ON ex.event_id = $1
		LEFT JOIN expense_participants ep ON ep.expense_id = ex.id AND ep.user_id = em.user_id
		WHERE em.event_id = $1
		GROUP BY em.user_id, u.display_name, em.emoji, em.role, em.payment_status, em.joined_at
		ORDER BY em.joined_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collectorUserID := int64(0)
	if e.CollectorUserID != nil {
		collectorUserID = *e.CollectorUserID
	}

	var members []models.MemberSummary
	var currentUserAmount int64
	var currentUserPayStatus models.PaymentStatus = models.PaymentUnpaid
	var currentUserRole models.MemberRole = models.RoleMember
	var paidAmount, requiredAmount int64
	var paidCount int

	for rows.Next() {
		var m models.MemberSummary
		var role string
		var payStatus string
		if err := rows.Scan(&m.UserID, &m.Name, &m.Emoji, &role, &payStatus, &m.AmountMinor); err != nil {
			return nil, err
		}
		m.Role = models.MemberRole(role)
		m.PaymentStatus = models.PaymentStatus(payStatus)
		m.IsCollector = m.UserID == collectorUserID

		members = append(members, m)

		if m.UserID == currentUserID {
			currentUserAmount = m.AmountMinor
			currentUserPayStatus = m.PaymentStatus
			currentUserRole = m.Role
		}

		// required = non-collector members total; paid = paid non-collectors
		if !m.IsCollector {
			requiredAmount += m.AmountMinor
			if m.PaymentStatus == models.PaymentPaid {
				paidAmount += m.AmountMinor
				paidCount++
			}
		} else {
			// collector counts as "paid"
			paidCount++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// collector info
	collectorInfo := models.CollectorInfo{
		Name:  e.CollectorName,
		Phone: e.CollectorPhone,
	}
	if collectorUserID != 0 {
		collectorInfo.UserID = collectorUserID
	}

	return &models.EventSummary{
		EventID:                  eventID,
		Title:                    e.Title,
		Currency:                 e.Currency,
		Status:                   e.Status,
		TotalAmountMinor:         total,
		CurrentUserAmountMinor:   currentUserAmount,
		CurrentUserPaymentStatus: currentUserPayStatus,
		CurrentUserRole:          currentUserRole,
		CurrentUserIsCollector:   currentUserID == collectorUserID,
		AllowMembersAddExpenses:  e.AllowMembersAddExpenses,
		Collector:                collectorInfo,
		Members:                  members,
		PaidAmountMinor:          paidAmount,
		RequiredAmountMinor:      requiredAmount,
		PaidCount:                paidCount,
		MembersCount:             len(members),
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *Store) getParticipants(ctx context.Context, expenseID int64) ([]models.ExpenseParticipant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ep.user_id, u.display_name, COALESCE(em.emoji,''), ep.share_minor
		FROM expense_participants ep
		JOIN users u ON u.id = ep.user_id
		LEFT JOIN event_members em ON em.user_id = ep.user_id
			AND em.event_id = (SELECT event_id FROM expenses WHERE id = $1)
		WHERE ep.expense_id = $1
		ORDER BY ep.user_id
	`, expenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.ExpenseParticipant
	for rows.Next() {
		var p models.ExpenseParticipant
		if err := rows.Scan(&p.UserID, &p.Name, &p.Emoji, &p.ShareMinor); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}

type share struct {
	userID     int64
	shareMinor int64
}

func calculateShares(amountMinor int64, userIDs []int64) []share {
	n := int64(len(userIDs))
	if n == 0 {
		return nil
	}
	base := amountMinor / n
	extra := amountMinor % n

	result := make([]share, len(userIDs))
	for i, uid := range userIDs {
		s := base
		if int64(i) < extra {
			s++
		}
		result[i] = share{userID: uid, shareMinor: s}
	}
	return result
}

func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}
