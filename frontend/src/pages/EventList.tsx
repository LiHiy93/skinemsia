import { useState, useEffect } from 'react'
import { api } from '../api/client'
import type { Event, Tab } from '../types'
import { formatMoney } from '../utils/money'

interface Props {
  onCreateEvent: () => void
  onJoinEvent: () => void
  onOpenEvent: (eventId: number, tab?: Tab) => void
}

export default function EventList({ onCreateEvent, onJoinEvent, onOpenEvent }: Props) {
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.listEvents()
      .then(setEvents)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const activeEvents = events.filter((e) => e.status === 'active')
  const archivedEvents = events.filter((e) => e.status === 'archived')

  return (
    <div className="page">
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>🤝 Скинемся</h1>
        <button className="btn btn-ghost" style={{ width: 'auto', padding: '6px 0' }} onClick={onJoinEvent}>
          Войти по коду
        </button>
      </div>

      <div className="page-body page-body--no-nav">
        {loading && <div className="loader">Загрузка...</div>}

        {!loading && error && (
          <div className="card" style={{ color: 'var(--tg-destructive)' }}>
            Ошибка: {error}
          </div>
        )}

        {!loading && !error && activeEvents.length === 0 && archivedEvents.length === 0 && (
          <div className="empty">
            <div className="empty-icon">🎉</div>
            <div className="empty-text">
              Создай первое событие или войди по коду приглашения от друга
            </div>
          </div>
        )}

        {activeEvents.length > 0 && (
          <>
            {activeEvents.map((e) => (
              <EventCard key={e.id} event={e} onOpen={() => onOpenEvent(e.id)} />
            ))}
          </>
        )}

        {archivedEvents.length > 0 && (
          <>
            <p className="section-title">Архив</p>
            {archivedEvents.map((e) => (
              <EventCard key={e.id} event={e} onOpen={() => onOpenEvent(e.id)} />
            ))}
          </>
        )}

        <button className="btn btn-primary" style={{ marginTop: 16 }} onClick={onCreateEvent}>
          ➕ Создать событие
        </button>
      </div>
    </div>
  )
}

function EventCard({ event, onOpen }: { event: Event; onOpen: () => void }) {
  const isArchived = event.status === 'archived'
  return (
    <div className="list-item" onClick={onOpen}>
      <div className="list-item-emoji">{isArchived ? '📦' : '🎉'}</div>
      <div className="list-item-body">
        <div className="list-item-title">{event.title}</div>
        <div className="list-item-sub">
          Код: {event.joinCode}
          {isArchived && ' · Архив'}
        </div>
      </div>
      <div style={{ color: 'var(--tg-hint)', fontSize: 20 }}>›</div>
    </div>
  )
}
