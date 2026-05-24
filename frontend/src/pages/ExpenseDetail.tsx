import { useState, useEffect } from 'react'
import { api } from '../api/client'
import type { Expense, EventSummary } from '../types'
import { formatMoney } from '../utils/money'

interface Props {
  eventId: number
  expenseId: number
  onBack: () => void
  onEdit: () => void
  onDeleted: () => void
  showToast: (msg: string) => void
}

export default function ExpenseDetail({ eventId, expenseId, onBack, onEdit, onDeleted, showToast }: Props) {
  const [expense, setExpense] = useState<Expense | null>(null)
  const [summary, setSummary] = useState<EventSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.getExpense(eventId, expenseId),
      api.getSummary(eventId),
    ]).then(([ex, sum]) => {
      setExpense(ex)
      setSummary(sum)
    }).finally(() => setLoading(false))
  }, [eventId, expenseId])

  const handleDelete = async () => {
    if (!confirm('Удалить этот расход?')) return
    try {
      await api.deleteExpense(eventId, expenseId)
      showToast('Расход удалён')
      onDeleted()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    }
  }

  if (loading) return <div className="loader" style={{ height: '100dvh' }}>Загрузка...</div>
  if (!expense || !summary) return <div className="loader">Ошибка</div>

  const currency = summary.currency
  const canEdit = summary.currentUserRole === 'creator' || expense.createdByUserId === getCurrentUserId()
  const allIds = new Set(summary.members.map(m => m.userId))
  const participantIds = new Set(expense.participants.map(p => p.userId))
  const nonParticipants = summary.members.filter(m => !participantIds.has(m.userId))

  function getCurrentUserId() {
    return window.Telegram?.WebApp?.initDataUnsafe?.user?.id ?? 0
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1 style={{ fontSize: 17 }}>{expense.title}</h1>
      </div>

      <div className="page-body page-body--no-nav">
        {/* Main info */}
        <div className="card">
          <div className="info-row">
            <span className="ir-label">Сумма</span>
            <span className="ir-value" style={{ fontSize: 20, fontWeight: 700 }}>
              {formatMoney(expense.amountMinor, currency)}
            </span>
          </div>
          <div className="info-row">
            <span className="ir-label">Оплатил</span>
            <span className="ir-value">{expense.paidByEmoji} {expense.paidByName}</span>
          </div>
          <div className="info-row">
            <span className="ir-label">Участников</span>
            <span className="ir-value">{expense.participants.length}</span>
          </div>
        </div>

        {/* Participants with shares */}
        <p className="section-title">Делится между</p>
        {expense.participants.map(p => (
          <div key={p.userId} style={{
            display: 'flex', alignItems: 'center', gap: 10,
            padding: '11px 14px',
            background: 'var(--tg-secondary-bg)',
            borderRadius: 'var(--radius-sm)',
            marginBottom: 8,
          }}>
            <span style={{ fontSize: 20 }}>{p.emoji}</span>
            <span style={{ flex: 1, fontWeight: 500 }}>{p.name}</span>
            <span style={{ fontWeight: 600 }}>{formatMoney(p.shareMinor, currency)}</span>
          </div>
        ))}

        {/* Non-participants */}
        {nonParticipants.length > 0 && (
          <>
            <p className="section-title">Не участвуют</p>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 12 }}>
              {nonParticipants.map(m => (
                <span key={m.userId} style={{
                  display: 'flex', alignItems: 'center', gap: 4,
                  fontSize: 14, color: 'var(--tg-hint)',
                  background: 'var(--tg-secondary-bg)',
                  padding: '4px 10px', borderRadius: 20,
                }}>
                  {m.emoji} {m.name}
                </span>
              ))}
            </div>
          </>
        )}

        {/* Action buttons */}
        {(summary.currentUserRole === 'creator' || expense.createdByUserId) && (
          <>
            <div className="sep" />
            <div className="btn-row">
              <button className="btn btn-secondary" onClick={onEdit}>
                ✏️ Изменить
              </button>
              <button className="btn btn-danger" onClick={handleDelete}>
                🗑 Удалить
              </button>
            </div>
          </>
        )}

        <button className="btn btn-ghost" style={{ marginTop: 8 }} onClick={onBack}>
          ← Назад
        </button>
      </div>
    </div>
  )
}
