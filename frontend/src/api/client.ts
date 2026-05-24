import type { Event, EventSummary, EventMember, Expense } from '../types'

const BASE = import.meta.env.VITE_API_URL ?? ''

function getInitData(): string {
  if (window.Telegram?.WebApp?.initData) {
    return window.Telegram.WebApp.initData
  }
  return ''
}

function getDevHeader(): Record<string, string> {
  // Only used when running outside Telegram in development
  const devId = import.meta.env.VITE_DEV_USER_ID
  if (devId) return { 'X-Dev-User-ID': devId }
  return {}
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const initData = getInitData()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(initData ? { Authorization: `Bearer ${initData}` } : getDevHeader()),
    ...(options.headers as Record<string, string> ?? {}),
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

// ── Events ────────────────────────────────────────────────────────────────────

export const api = {
  listEvents: () =>
    request<Event[]>('/api/events'),

  createEvent: (data: {
    title: string
    collectorName: string
    collectorPhone: string
    currency: string
  }) => request<Event>('/api/events', { method: 'POST', body: JSON.stringify(data) }),

  getEvent: (id: number) =>
    request<Event>(`/api/events/${id}`),

  updateEvent: (id: number, data: Partial<{
    title: string
    collectorName: string
    collectorPhone: string
    currency: string
    allowMembersAddExpenses: boolean
  }>) => request<Event>(`/api/events/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),

  archiveEvent: (id: number) =>
    request<{ status: string }>(`/api/events/${id}/archive`, { method: 'POST' }),

  deleteEvent: (id: number) =>
    request<{ status: string }>(`/api/events/${id}`, { method: 'DELETE' }),

  previewEvent: (code: string) =>
    request<{ id: number; title: string; code: string }>(`/api/events/preview/${code}`),

  joinEvent: (code: string) =>
    request<Event>('/api/events/join', { method: 'POST', body: JSON.stringify({ code }) }),

  getSummary: (id: number) =>
    request<EventSummary>(`/api/events/${id}/summary`),

  // ── Members ─────────────────────────────────────────────────────────────────

  listMembers: (eventId: number) =>
    request<EventMember[]>(`/api/events/${eventId}/members`),

  removeMember: (eventId: number, userId: number) =>
    request<{ status: string }>(`/api/events/${eventId}/members/${userId}`, { method: 'DELETE' }),

  updateMemberEmoji: (eventId: number, userId: number, emoji: string) =>
    request<{ status: string }>(`/api/events/${eventId}/members/${userId}/emoji`, {
      method: 'PATCH',
      body: JSON.stringify({ emoji }),
    }),

  // ── Expenses ─────────────────────────────────────────────────────────────────

  listExpenses: (eventId: number) =>
    request<Expense[]>(`/api/events/${eventId}/expenses`),

  getExpense: (eventId: number, expenseId: number) =>
    request<Expense>(`/api/events/${eventId}/expenses/${expenseId}`),

  createExpense: (eventId: number, data: {
    title: string
    amountMinor: number
    paidByUserId: number
    participantIds: number[]
  }) => request<Expense>(`/api/events/${eventId}/expenses`, { method: 'POST', body: JSON.stringify(data) }),

  updateExpense: (eventId: number, expenseId: number, data: {
    title: string
    amountMinor: number
    paidByUserId: number
    participantIds: number[]
  }) => request<Expense>(`/api/events/${eventId}/expenses/${expenseId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),

  deleteExpense: (eventId: number, expenseId: number) =>
    request<{ status: string }>(`/api/events/${eventId}/expenses/${expenseId}`, { method: 'DELETE' }),

  // ── Payment ──────────────────────────────────────────────────────────────────

  markPaid: (eventId: number) =>
    request<{ paymentStatus: string }>(`/api/events/${eventId}/payment/paid`, { method: 'POST' }),

  unmarkPaid: (eventId: number) =>
    request<{ paymentStatus: string }>(`/api/events/${eventId}/payment/paid`, { method: 'DELETE' }),
}

// Telegram WebApp global type
declare global {
  interface Window {
    Telegram?: {
      WebApp: {
        ready: () => void
        expand: () => void
        close: () => void
        initData: string
        initDataUnsafe: {
          user?: {
            id: number
            first_name: string
            last_name?: string
            username?: string
          }
          start_param?: string
        }
        themeParams: Record<string, string>
        colorScheme: 'light' | 'dark'
        BackButton: {
          show: () => void
          hide: () => void
          onClick: (fn: () => void) => void
          offClick: (fn: () => void) => void
        }
        HapticFeedback: {
          impactOccurred: (style: string) => void
          notificationOccurred: (type: string) => void
        }
      }
    }
  }
}
