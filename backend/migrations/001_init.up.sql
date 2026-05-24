CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE NOT NULL,
    username        TEXT NOT NULL DEFAULT '',
    first_name      TEXT NOT NULL DEFAULT '',
    last_name       TEXT NOT NULL DEFAULT '',
    display_name    TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
    id                          BIGSERIAL PRIMARY KEY,
    title                       TEXT NOT NULL,
    creator_user_id             BIGINT NOT NULL REFERENCES users(id),
    collector_user_id           BIGINT REFERENCES users(id),
    collector_name              TEXT NOT NULL DEFAULT '',
    collector_phone             TEXT NOT NULL DEFAULT '',
    currency                    TEXT NOT NULL DEFAULT 'RUB',
    join_code                   TEXT UNIQUE NOT NULL,
    status                      TEXT NOT NULL DEFAULT 'active'
                                    CHECK (status IN ('active','archived','deleted')),
    allow_members_add_expenses  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at                 TIMESTAMPTZ,
    deleted_at                  TIMESTAMPTZ
);

CREATE TABLE event_members (
    event_id        BIGINT NOT NULL REFERENCES events(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    role            TEXT NOT NULL DEFAULT 'member'
                        CHECK (role IN ('creator','member')),
    emoji           TEXT NOT NULL DEFAULT '🟢',
    payment_status  TEXT NOT NULL DEFAULT 'unpaid'
                        CHECK (payment_status IN ('unpaid','paid')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, user_id)
);

CREATE TABLE expenses (
    id                  BIGSERIAL PRIMARY KEY,
    event_id            BIGINT NOT NULL REFERENCES events(id),
    title               TEXT NOT NULL,
    amount_minor        BIGINT NOT NULL CHECK (amount_minor > 0),
    paid_by_user_id     BIGINT NOT NULL REFERENCES users(id),
    created_by_user_id  BIGINT NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE expense_participants (
    expense_id  BIGINT NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    share_minor BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (expense_id, user_id)
);

CREATE INDEX idx_events_join_code  ON events(join_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_creator    ON events(creator_user_id);
CREATE INDEX idx_em_user_id        ON event_members(user_id);
CREATE INDEX idx_expenses_event    ON expenses(event_id);
CREATE INDEX idx_ep_user           ON expense_participants(user_id);
