import { useState, useEffect, useCallback } from 'react'
import type { Screen, Tab } from './types'
import EventList from './pages/EventList'
import EventCreate from './pages/EventCreate'
import EventJoin from './pages/EventJoin'
import EventMain from './pages/EventMain'
import ExpenseForm from './pages/ExpenseForm'
import ExpenseDetail from './pages/ExpenseDetail'
import Toast from './components/Toast'

export default function App() {
  const [screen, setScreen] = useState<Screen>({ type: 'list' })
  const [toast, setToast] = useState('')

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(''), 2000)
  }, [])

  const go = useCallback((s: Screen) => setScreen(s), [])

  // Handle Telegram start_param (join code)
  useEffect(() => {
    const param = window.Telegram?.WebApp?.initDataUnsafe?.start_param
    if (param) {
      setScreen({ type: 'join', code: param })
    }
  }, [])

  // Telegram Back Button
  useEffect(() => {
    const tg = window.Telegram?.WebApp
    if (!tg) return

    const isRoot = screen.type === 'list'
    if (isRoot) {
      tg.BackButton.hide()
    } else {
      tg.BackButton.show()
    }

    const handleBack = () => {
      switch (screen.type) {
        case 'event':
          go({ type: 'list' })
          break
        case 'create':
        case 'join':
          go({ type: 'list' })
          break
        case 'expenseForm':
          go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })
          break
        case 'expenseDetail':
          go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })
          break
        default:
          go({ type: 'list' })
      }
    }

    tg.BackButton.onClick(handleBack)
    return () => tg.BackButton.offClick(handleBack)
  }, [screen, go])

  const goToEvent = (eventId: number, tab: Tab = 'expenses') =>
    go({ type: 'event', eventId, tab })

  return (
    <>
      {screen.type === 'list' && (
        <EventList
          onCreateEvent={() => go({ type: 'create' })}
          onJoinEvent={() => go({ type: 'join' })}
          onOpenEvent={goToEvent}
        />
      )}

      {screen.type === 'create' && (
        <EventCreate
          onBack={() => go({ type: 'list' })}
          onCreated={(eventId) => goToEvent(eventId)}
          showToast={showToast}
        />
      )}

      {screen.type === 'join' && (
        <EventJoin
          initialCode={screen.code}
          onBack={() => go({ type: 'list' })}
          onJoined={(eventId) => goToEvent(eventId)}
          showToast={showToast}
        />
      )}

      {screen.type === 'event' && (
        <EventMain
          eventId={screen.eventId}
          initialTab={screen.tab}
          onBack={() => go({ type: 'list' })}
          onAddExpense={() => go({ type: 'expenseForm', eventId: screen.eventId })}
          onOpenExpense={(expId) => go({ type: 'expenseDetail', eventId: screen.eventId, expenseId: expId })}
          onTabChange={(tab) => go({ type: 'event', eventId: screen.eventId, tab })}
          showToast={showToast}
        />
      )}

      {screen.type === 'expenseForm' && (
        <ExpenseForm
          eventId={screen.eventId}
          expenseId={screen.expenseId}
          onBack={() => go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })}
          onSaved={() => go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })}
          showToast={showToast}
        />
      )}

      {screen.type === 'expenseDetail' && (
        <ExpenseDetail
          eventId={screen.eventId}
          expenseId={screen.expenseId}
          onBack={() => go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })}
          onEdit={() => go({ type: 'expenseForm', eventId: screen.eventId, expenseId: screen.expenseId })}
          onDeleted={() => go({ type: 'event', eventId: screen.eventId, tab: 'expenses' })}
          showToast={showToast}
        />
      )}

      {toast && <Toast message={toast} />}
    </>
  )
}
