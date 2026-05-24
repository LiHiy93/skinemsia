import { useState } from 'react'
import { api } from '../api/client'

interface Props {
  onBack: () => void
  onCreated: (eventId: number) => void
  showToast: (msg: string) => void
}

const CURRENCIES = ['RUB', 'USD', 'EUR', 'GBP', 'UAH', 'KZT']

export default function EventCreate({ onBack, onCreated, showToast }: Props) {
  const [title, setTitle] = useState('')
  const [collectorName, setCollectorName] = useState('')
  const [collectorPhone, setCollectorPhone] = useState('')
  const [currency, setCurrency] = useState('RUB')
  const [loading, setLoading] = useState(false)

  // pre-fill collector name from Telegram user
  useState(() => {
    const u = window.Telegram?.WebApp?.initDataUnsafe?.user
    if (u) {
      setCollectorName(u.first_name + (u.last_name ? ' ' + u.last_name : ''))
    }
  })

  const canSubmit = title.trim().length > 0 && collectorPhone.trim().length > 0

  const handleCreate = async () => {
    if (!canSubmit || loading) return
    setLoading(true)
    try {
      const event = await api.createEvent({
        title: title.trim(),
        collectorName: collectorName.trim(),
        collectorPhone: collectorPhone.trim(),
        currency,
      })
      showToast('Событие создано!')
      onCreated(event.id)
    } catch (e: any) {
      showToast('Ошибка: ' + e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>Новое событие</h1>
      </div>

      <div className="page-body page-body--no-nav">
        <div className="field">
          <label>Название события *</label>
          <input
            className="input"
            placeholder="Например: Пикник на озере"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={100}
          />
        </div>

        <div className="field">
          <label>Имя сборщика денег</label>
          <input
            className="input"
            placeholder="Кто будет собирать деньги"
            value={collectorName}
            onChange={(e) => setCollectorName(e.target.value)}
            maxLength={60}
          />
        </div>

        <div className="field">
          <label>Номер телефона сборщика *</label>
          <input
            className="input"
            placeholder="+7 999 123-45-67"
            type="tel"
            value={collectorPhone}
            onChange={(e) => setCollectorPhone(e.target.value)}
            maxLength={30}
          />
        </div>

        <div className="field">
          <label>Валюта</label>
          <select
            className="input"
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
          >
            {CURRENCIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </div>

        <div className="card" style={{ marginTop: 8, marginBottom: 16 }}>
          <p style={{ fontSize: 13, color: 'var(--tg-hint)', lineHeight: 1.5 }}>
            После создания события ты получишь код приглашения. Поделись им с участниками.
          </p>
        </div>

        <button
          className="btn btn-primary"
          disabled={!canSubmit || loading}
          onClick={handleCreate}
        >
          {loading ? 'Создаём...' : '✅ Создать событие'}
        </button>

        <button className="btn btn-secondary" style={{ marginTop: 8 }} onClick={onBack}>
          Отмена
        </button>
      </div>
    </div>
  )
}
