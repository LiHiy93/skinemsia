import { useState, useEffect } from 'react'
import { api } from '../api/client'

interface Props {
  initialCode?: string
  onBack: () => void
  onJoined: (eventId: number) => void
  showToast: (msg: string) => void
}

export default function EventJoin({ initialCode, onBack, onJoined, showToast }: Props) {
  const [code, setCode] = useState(initialCode ?? '')
  const [loading, setLoading] = useState(false)

  // auto-join if code was passed from start_param
  useEffect(() => {
    if (initialCode) {
      handleJoin(initialCode)
    }
  }, [])

  const handleJoin = async (joinCode = code) => {
    const c = joinCode.trim().toUpperCase()
    if (!c) return
    setLoading(true)
    try {
      const event = await api.joinEvent(c)
      showToast(`Добро пожаловать в «${event.title}»!`)
      onJoined(event.id)
    } catch (e: any) {
      showToast('Событие не найдено. Проверь код.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>Войти по коду</h1>
      </div>

      <div className="page-body page-body--no-nav">
        <div className="empty" style={{ padding: '32px 0 24px' }}>
          <div className="empty-icon">🔑</div>
        </div>

        <div className="field">
          <label>Код приглашения</label>
          <input
            className="input"
            placeholder="Например: AB3K7X2Q"
            value={code}
            onChange={(e) => setCode(e.target.value.toUpperCase())}
            maxLength={10}
            style={{ textTransform: 'uppercase', letterSpacing: 2, fontSize: 20, textAlign: 'center' }}
            autoFocus
          />
        </div>

        <button
          className="btn btn-primary"
          disabled={code.trim().length < 4 || loading}
          onClick={() => handleJoin()}
        >
          {loading ? 'Ищем событие...' : '➡️ Войти'}
        </button>

        <button className="btn btn-secondary" style={{ marginTop: 8 }} onClick={onBack}>
          Отмена
        </button>
      </div>
    </div>
  )
}
