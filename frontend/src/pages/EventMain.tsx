import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { EventSummary, Expense, EventMember, Tab } from '../types'
import { formatMoney } from '../utils/money'
import BottomNav from '../components/BottomNav'

interface Props {
  eventId: number
  initialTab: Tab
  onBack: () => void
  onAddExpense: () => void
  onOpenExpense: (expId: number) => void
  onTabChange: (tab: Tab) => void
  showToast: (msg: string) => void
}

export default function EventMain({ eventId, initialTab, onBack, onAddExpense, onOpenExpense, onTabChange, showToast }: Props) {
  const [tab, setTab] = useState<Tab>(initialTab)
  const [summary, setSummary] = useState<EventSummary | null>(null)
  const [loading, setLoading] = useState(true)

  const reload = useCallback(() => {
    api.getSummary(eventId)
      .then(setSummary)
      .catch(() => showToast('Не удалось загрузить данные'))
      .finally(() => setLoading(false))
  }, [eventId])

  useEffect(() => { reload() }, [reload])

  const handleTab = (t: Tab) => {
    setTab(t)
    onTabChange(t)
  }

  if (loading) return <div className="loader" style={{ height: '100dvh' }}>Загрузка...</div>
  if (!summary) return <div className="loader" style={{ height: '100dvh' }}>Ошибка загрузки</div>

  return (
    <div className="page">
      {/* Header */}
      <div className="page-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'space-between' }}>
          <h1 style={{ fontSize: 17 }}>{summary.title}</h1>
          {summary.status === 'archived' && (
            <span style={{ fontSize: 12, color: 'var(--tg-hint)', background: 'var(--tg-secondary-bg)', padding: '2px 8px', borderRadius: 8 }}>Архив</span>
          )}
        </div>
      </div>

      {/* Tab body */}
      <div className="page-body">
        {tab === 'expenses' && (
          <ExpensesTab
            eventId={eventId}
            summary={summary}
            onAddExpense={onAddExpense}
            onOpenExpense={onOpenExpense}
            onReload={reload}
          />
        )}
        {tab === 'payment' && (
          <PaymentTab
            summary={summary}
            onReload={reload}
            showToast={showToast}
          />
        )}
        {tab === 'members' && (
          <MembersTab
            eventId={eventId}
            summary={summary}
            showToast={showToast}
            onReload={reload}
          />
        )}
        {tab === 'settings' && (
          <SettingsTab
            eventId={eventId}
            summary={summary}
            onBack={onBack}
            showToast={showToast}
            onReload={reload}
          />
        )}
      </div>

      <BottomNav active={tab} onChange={handleTab} />
    </div>
  )
}

// ── Expenses Tab ──────────────────────────────────────────────────────────────

function ExpensesTab({ eventId, summary, onAddExpense, onOpenExpense, onReload }: {
  eventId: number
  summary: EventSummary
  onAddExpense: () => void
  onOpenExpense: (id: number) => void
  onReload: () => void
}) {
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.listExpenses(eventId)
      .then(setExpenses)
      .finally(() => setLoading(false))
  }, [eventId])

  const canAdd = summary.allowMembersAddExpenses || summary.currentUserRole === 'creator'

  return (
    <>
      {/* Summary cards */}
      <div className="summary-grid">
        <div className="summary-card">
          <div className="label">Общая сумма</div>
          <div className="value">{formatMoney(summary.totalAmountMinor, summary.currency)}</div>
        </div>
        <div className="summary-card">
          <div className="label">Твоя сумма</div>
          <div className="value" style={{ color: summary.currentUserAmountMinor > 0 && !summary.currentUserIsCollector ? 'var(--tg-destructive)' : 'inherit' }}>
            {formatMoney(summary.currentUserAmountMinor, summary.currency)}
          </div>
        </div>
        <div className="summary-card">
          <div className="label">Участников</div>
          <div className="value">{summary.membersCount}</div>
        </div>
        <div className="summary-card">
          <div className="label">Твой статус</div>
          <div className="value" style={{ fontSize: 16 }}>
            {summary.currentUserPaymentStatus === 'paid' ? '✅ Оплачено' : '❌ Не оплачено'}
          </div>
        </div>
      </div>

      {/* Expenses list */}
      {loading && <div className="loader">Загрузка...</div>}

      {!loading && expenses.length === 0 && (
        <div className="empty">
          <div className="empty-icon">🧾</div>
          <div className="empty-text">Пока нет расходов. Добавьте первый.</div>
        </div>
      )}

      {expenses.map((ex) => (
        <div className="list-item" key={ex.id} onClick={() => onOpenExpense(ex.id)}>
          <div className="list-item-body">
            <div className="list-item-title">{ex.title}</div>
            <div className="list-item-sub" style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
              <span>{ex.paidByEmoji} {ex.paidByName}</span>
              <span style={{ color: 'var(--tg-hint)' }}>·</span>
              <span>{ex.participants.map(p => p.emoji).join(' ')}</span>
            </div>
          </div>
          <div className="list-item-right">
            <div className="list-item-amount">{formatMoney(ex.amountMinor, summary.currency)}</div>
          </div>
        </div>
      ))}

      {canAdd && (
        <button className="btn btn-primary" style={{ marginTop: 8 }} onClick={onAddExpense}>
          ➕ Добавить расход
        </button>
      )}
    </>
  )
}

// ── Payment Tab ───────────────────────────────────────────────────────────────

function PaymentTab({ summary, onReload, showToast }: {
  summary: EventSummary
  onReload: () => void
  showToast: (msg: string) => void
}) {
  const [loading, setLoading] = useState(false)

  const isPaid = summary.currentUserPaymentStatus === 'paid'
  const isCollector = summary.currentUserIsCollector

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => showToast(`${label} скопировано`))
    window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light')
  }

  const handlePayment = async () => {
    setLoading(true)
    try {
      if (isPaid) {
        await api.unmarkPaid(summary.eventId)
        showToast('Отметка оплаты снята')
      } else {
        await api.markPaid(summary.eventId)
        showToast('Оплата отмечена ✅')
        window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      }
      onReload()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    } finally {
      setLoading(false)
    }
  }

  const myMember = summary.members.find(m => m.isCollector === false && m.paymentStatus !== undefined)
  const amountStr = formatMoney(summary.currentUserAmountMinor, summary.currency)

  const progress = summary.requiredAmountMinor > 0
    ? Math.round((summary.paidAmountMinor / summary.requiredAmountMinor) * 100)
    : 100

  return (
    <>
      {/* My payment block */}
      <div className="card" style={{ marginBottom: 12 }}>
        <div className="card-label">{isCollector ? 'Твоя доля' : 'Ты должен'}</div>
        <div className="card-value" style={{ color: summary.currentUserAmountMinor > 0 && !isPaid && !isCollector ? 'var(--tg-destructive)' : 'inherit' }}>
          {amountStr}
        </div>

        {isCollector && (
          <div style={{ marginTop: 6, fontSize: 13, color: 'var(--tg-hint)' }}>
            Ты собираешь деньги — тебе не нужно никому платить
          </div>
        )}

        {!isCollector && summary.currentUserAmountMinor > 0 && (
          <>
            <div className="sep" />
            <div className="info-row">
              <span className="ir-label">Кому перевести</span>
              <span className="ir-value">{summary.collector.name || 'Сборщик'}</span>
            </div>
            <div className="info-row">
              <span className="ir-label">Телефон</span>
              <span className="ir-value">{summary.collector.phone || '—'}</span>
            </div>

            <div className="btn-row" style={{ marginTop: 12 }}>
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => handleCopy(summary.collector.phone, 'Номер')}
              >
                📋 Скопировать номер
              </button>
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => handleCopy(String(summary.currentUserAmountMinor / 100), 'Сумма')}
              >
                💰 Скопировать сумму
              </button>
            </div>
          </>
        )}
      </div>

      {/* Pay button — hidden for collector */}
      {!isCollector && (
        <button
          className={`btn ${isPaid ? 'btn-secondary' : 'btn-primary'}`}
          onClick={handlePayment}
          disabled={loading}
        >
          {isPaid ? '↩️ Отменить оплату' : '✅ Я оплатил'}
        </button>
      )}

      {/* Progress */}
      <div className="sep" style={{ margin: '16px 0 8px' }} />
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, color: 'var(--tg-hint)', marginBottom: 4 }}>
        <span>Собрано: {formatMoney(summary.paidAmountMinor, summary.currency)}</span>
        <span>{progress}%</span>
      </div>
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${progress}%` }} />
      </div>
      <div style={{ fontSize: 13, color: 'var(--tg-hint)', marginTop: 4, marginBottom: 12 }}>
        Оплатили: {summary.paidCount} из {summary.membersCount}
      </div>

      {/* Status table */}
      <p className="section-title">📋 Статусы оплат</p>
      {summary.members.map((m) => (
        <div className="payment-row" key={m.userId}>
          <span className="pr-icon">{m.paymentStatus === 'paid' ? '✅' : '❌'}</span>
          <span className="pr-emoji">{m.emoji}</span>
          <span className="pr-name">
            {m.name}
            {m.isCollector && <span style={{ fontSize: 11, color: 'var(--tg-hint)', marginLeft: 4 }}>сборщик</span>}
          </span>
          <span className="pr-amount">{m.isCollector ? '—' : formatMoney(m.amountMinor, summary.currency)}</span>
        </div>
      ))}
    </>
  )
}

// ── Members Tab ───────────────────────────────────────────────────────────────

function MembersTab({ eventId, summary, showToast, onReload }: {
  eventId: number
  summary: EventSummary
  showToast: (msg: string) => void
  onReload: () => void
}) {
  const [members, setMembers] = useState<EventMember[]>([])
  const isCreator = summary.currentUserRole === 'creator'

  useEffect(() => {
    api.listMembers(eventId).then(setMembers)
  }, [eventId])

  const [joinCodeStr, setJoinCodeStr] = useState('')

  useEffect(() => {
    api.getEvent(eventId).then(e => setJoinCodeStr(e.joinCode)).catch(() => {})
  }, [eventId])

  const copyInviteLink = () => {
    if (!joinCodeStr) { showToast('Загрузка кода...'); return }
    const link = `https://t.me/skinemsia_bot?startapp=${joinCodeStr}`
    navigator.clipboard.writeText(link).then(() => showToast('Ссылка скопирована'))
  }

  const handleRemove = async (userId: number, name: string) => {
    if (!confirm(`Удалить ${name} из события?`)) return
    try {
      await api.removeMember(eventId, userId)
      setMembers(members.filter(m => m.userId !== userId))
      showToast('Участник удалён')
      onReload()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    }
  }

  return (
    <>
      <p className="section-title">Участники ({members.length})</p>

      {members.map((m) => (
        <div
          key={m.userId}
          style={{
            display: 'flex', alignItems: 'center', gap: 10,
            padding: '12px 14px',
            background: 'var(--tg-secondary-bg)',
            borderRadius: 'var(--radius-sm)',
            marginBottom: 8,
          }}
        >
          <span style={{ fontSize: 22 }}>{m.emoji}</span>
          <div style={{ flex: 1 }}>
            <div style={{ fontWeight: 600 }}>{m.name}</div>
            <div style={{ fontSize: 12, color: 'var(--tg-hint)' }}>
              {m.role === 'creator' ? 'Создатель' : 'Участник'}
              {summary.members.find(s => s.userId === m.userId)?.isCollector && ' · Сборщик'}
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: 18 }}>
              {m.paymentStatus === 'paid' ? '✅' : '❌'}
            </span>
            {isCreator && m.role !== 'creator' && (
              <button
                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--tg-destructive)', fontSize: 18, padding: '4px' }}
                onClick={() => handleRemove(m.userId, m.name)}
              >
                ✕
              </button>
            )}
          </div>
        </div>
      ))}

      <div className="sep" />

      <button className="btn btn-secondary" onClick={copyInviteLink}>
        🔗 Пригласить участника
      </button>
    </>
  )
}

// ── Settings Tab ──────────────────────────────────────────────────────────────

function SettingsTab({ eventId, summary, onBack, showToast, onReload }: {
  eventId: number
  summary: EventSummary
  onBack: () => void
  showToast: (msg: string) => void
  onReload: () => void
}) {
  const [title, setTitle] = useState(summary.title)
  const [collectorName, setCollectorName] = useState(summary.collector.name)
  const [collectorPhone, setCollectorPhone] = useState(summary.collector.phone)
  const [allowAll, setAllowAll] = useState(summary.allowMembersAddExpenses)
  const [loading, setLoading] = useState(false)

  const isCreator = summary.currentUserRole === 'creator'
  const joinCode = summary.eventId // we don't have joinCode in summary — fetch event
  const [joinCodeStr, setJoinCodeStr] = useState('')

  useEffect(() => {
    api.getEvent(eventId).then(e => setJoinCodeStr(e.joinCode)).catch(() => {})
  }, [eventId])

  const copyCode = () => {
    navigator.clipboard.writeText(joinCodeStr)
    showToast('Код скопирован: ' + joinCodeStr)
  }

  const copyLink = () => {
    const link = `https://t.me/skinemsia_bot?startapp=${joinCodeStr}`
    navigator.clipboard.writeText(link).then(() => showToast('Ссылка скопирована'))
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      await api.updateEvent(eventId, {
        title: title.trim(),
        collectorName: collectorName.trim(),
        collectorPhone: collectorPhone.trim(),
        allowMembersAddExpenses: allowAll,
      })
      showToast('Сохранено ✅')
      onReload()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    } finally {
      setLoading(false)
    }
  }

  const handleArchive = async () => {
    if (!confirm('Архивировать событие?')) return
    try {
      await api.archiveEvent(eventId)
      showToast('Событие архивировано')
      onBack()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Удалить событие? Это действие нельзя отменить.')) return
    try {
      await api.deleteEvent(eventId)
      showToast('Событие удалено')
      onBack()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    }
  }

  return (
    <>
      {/* Join code */}
      <div className="card" style={{ marginBottom: 16 }}>
        <div className="card-label">Код приглашения</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 6 }}>
          <span style={{ fontSize: 22, fontWeight: 700, letterSpacing: 3, fontFamily: 'monospace' }}>
            {joinCodeStr || '...'}
          </span>
          <button className="btn btn-secondary btn-sm" style={{ width: 'auto' }} onClick={copyCode}>
            📋 Код
          </button>
          <button className="btn btn-secondary btn-sm" style={{ width: 'auto' }} onClick={copyLink}>
            🔗 Ссылка
          </button>
        </div>
      </div>

      {isCreator ? (
        <>
          <div className="field">
            <label>Название события</label>
            <input className="input" value={title} onChange={e => setTitle(e.target.value)} maxLength={100} />
          </div>

          <div className="field">
            <label>Имя сборщика</label>
            <input className="input" value={collectorName} onChange={e => setCollectorName(e.target.value)} maxLength={60} />
          </div>

          <div className="field">
            <label>Телефон сборщика</label>
            <input className="input" type="tel" value={collectorPhone} onChange={e => setCollectorPhone(e.target.value)} maxLength={30} />
          </div>

          <div
            className="checkbox-row"
            style={{ marginBottom: 16 }}
            onClick={() => setAllowAll(!allowAll)}
          >
            <input type="checkbox" checked={allowAll} readOnly />
            <span className="member-name">Все участники могут добавлять расходы</span>
          </div>

          <button className="btn btn-primary" disabled={loading} onClick={handleSave}>
            {loading ? 'Сохраняем...' : '💾 Сохранить'}
          </button>

          <div className="sep" style={{ margin: '16px 0' }} />

          {summary.status === 'active' && (
            <button className="btn btn-secondary" style={{ marginBottom: 8 }} onClick={handleArchive}>
              📦 Архивировать событие
            </button>
          )}

          <button className="btn btn-danger" onClick={handleDelete}>
            🗑 Удалить событие
          </button>
        </>
      ) : (
        <div className="card">
          <p style={{ fontSize: 14, color: 'var(--tg-hint)' }}>
            Только создатель события может изменять настройки.
          </p>
        </div>
      )}
    </>
  )
}
