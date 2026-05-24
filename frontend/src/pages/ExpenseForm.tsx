import { useState, useEffect } from 'react'
import { api } from '../api/client'
import type { EventMember } from '../types'
import { parseMoneyInput, formatMoney } from '../utils/money'

interface Props {
  eventId: number
  expenseId?: number
  onBack: () => void
  onSaved: () => void
  showToast: (msg: string) => void
}

export default function ExpenseForm({ eventId, expenseId, onBack, onSaved, showToast }: Props) {
  const isEdit = expenseId !== undefined

  const [members, setMembers] = useState<EventMember[]>([])
  const [currency, setCurrency] = useState('RUB')

  const [title, setTitle] = useState('')
  const [amountStr, setAmountStr] = useState('')
  const [paidByUserId, setPaidByUserId] = useState<number | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  // Load members and event currency
  useEffect(() => {
    Promise.all([
      api.listMembers(eventId),
      api.getEvent(eventId),
    ]).then(([ms, ev]) => {
      setMembers(ms)
      setCurrency(ev.currency)
      // default: current user as payer, all selected
      const allIds = new Set(ms.map(m => m.userId))
      setSelectedIds(allIds)

      // try to set current payer = current user (from Telegram)
      const tgId = window.Telegram?.WebApp?.initDataUnsafe?.user?.id
      const me = ms.find(m => m.userId === tgId || m.role === 'creator')
      if (me) setPaidByUserId(me.userId)
      else if (ms.length > 0) setPaidByUserId(ms[0].userId)
    }).finally(() => setLoading(false))
  }, [eventId])

  // Load existing expense for edit
  useEffect(() => {
    if (!expenseId) return
    api.getExpense(eventId, expenseId).then(ex => {
      setTitle(ex.title)
      setAmountStr(String(ex.amountMinor / 100))
      setPaidByUserId(ex.paidByUserId)
      setSelectedIds(new Set(ex.participants.map(p => p.userId)))
    })
  }, [expenseId])

  const toggleMember = (id: number) => {
    const next = new Set(selectedIds)
    if (next.has(id)) {
      if (next.size === 1) return // at least one
      next.delete(id)
    } else {
      next.add(id)
    }
    setSelectedIds(next)
  }

  const selectAll = () => setSelectedIds(new Set(members.map(m => m.userId)))
  const deselectAll = () => {
    // leave only payer
    if (paidByUserId) setSelectedIds(new Set([paidByUserId]))
  }

  const amountMinor = parseMoneyInput(amountStr)
  const canSave = title.trim().length > 0 && amountMinor > 0 && selectedIds.size > 0 && paidByUserId !== null

  const handleSave = async () => {
    if (!canSave || saving) return
    setSaving(true)
    try {
      const data = {
        title: title.trim(),
        amountMinor,
        paidByUserId: paidByUserId!,
        participantIds: Array.from(selectedIds),
      }
      if (isEdit) {
        await api.updateExpense(eventId, expenseId!, data)
        showToast('Расход обновлён')
      } else {
        await api.createExpense(eventId, data)
        showToast('Расход добавлен ✅')
      }
      onSaved()
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    } finally {
      setSaving(false)
    }
  }

  // Preview split
  const preview = () => {
    if (amountMinor <= 0 || selectedIds.size === 0) return null
    const ids = Array.from(selectedIds)
    const base = Math.floor(amountMinor / ids.length)
    const extra = amountMinor % ids.length
    return ids.map((id, i) => ({
      id,
      share: i < extra ? base + 1 : base,
    }))
  }
  const splits = preview()

  if (loading) return <div className="loader" style={{ height: '100dvh' }}>Загрузка...</div>

  return (
    <div className="page">
      <div className="page-header">
        <h1>{isEdit ? 'Изменить расход' : 'Новый расход'}</h1>
      </div>

      <div className="page-body page-body--no-nav">
        <div className="field">
          <label>Название *</label>
          <input
            className="input"
            placeholder="Например: Мясо, Бензин, Алкоголь..."
            value={title}
            onChange={e => setTitle(e.target.value)}
            maxLength={100}
            autoFocus
          />
        </div>

        <div className="field">
          <label>Сумма ({currency}) *</label>
          <input
            className="input"
            placeholder="0"
            type="number"
            inputMode="decimal"
            min="0"
            step="0.01"
            value={amountStr}
            onChange={e => setAmountStr(e.target.value)}
          />
        </div>

        <div className="field">
          <label>Кто оплатил</label>
          <select
            className="input"
            value={paidByUserId ?? ''}
            onChange={e => setPaidByUserId(Number(e.target.value))}
          >
            {members.map(m => (
              <option key={m.userId} value={m.userId}>
                {m.emoji} {m.name}
              </option>
            ))}
          </select>
        </div>

        <div className="field">
          <label>На кого делится</label>
          <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
            <button className="btn btn-secondary btn-sm" style={{ width: 'auto' }} onClick={selectAll}>
              Все
            </button>
            <button className="btn btn-secondary btn-sm" style={{ width: 'auto' }} onClick={deselectAll}>
              Снять
            </button>
          </div>

          {members.map(m => {
            const checked = selectedIds.has(m.userId)
            const split = splits?.find(s => s.id === m.userId)
            return (
              <div
                key={m.userId}
                className="checkbox-row"
                onClick={() => toggleMember(m.userId)}
              >
                <input type="checkbox" checked={checked} readOnly />
                <span className="member-emoji">{m.emoji}</span>
                <span className="member-name">{m.name}</span>
                {checked && split && amountMinor > 0 && (
                  <span style={{ fontSize: 13, color: 'var(--tg-hint)', flexShrink: 0 }}>
                    {formatMoney(split.share, currency)}
                  </span>
                )}
              </div>
            )
          })}
        </div>

        <button
          className="btn btn-primary"
          disabled={!canSave || saving}
          onClick={handleSave}
        >
          {saving ? 'Сохраняем...' : isEdit ? '💾 Сохранить' : '✅ Добавить расход'}
        </button>

        <button className="btn btn-secondary" style={{ marginTop: 8 }} onClick={onBack}>
          Отмена
        </button>
      </div>
    </div>
  )
}
