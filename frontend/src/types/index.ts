export type EventStatus = 'active' | 'archived' | 'deleted'
export type MemberRole = 'creator' | 'member'
export type PaymentStatus = 'unpaid' | 'paid'

export interface User {
  id: number
  telegramUserId: number
  username: string
  firstName: string
  lastName: string
  displayName: string
}

export interface Event {
  id: number
  title: string
  creatorUserId: number
  collectorUserId: number | null
  collectorName: string
  collectorPhone: string
  currency: string
  joinCode: string
  status: EventStatus
  allowMembersAddExpenses: boolean
  createdAt: string
  updatedAt: string
  archivedAt: string | null
}

export interface EventMember {
  eventId: number
  userId: number
  name: string
  username: string
  role: MemberRole
  emoji: string
  paymentStatus: PaymentStatus
  joinedAt: string
}

export interface ExpenseParticipant {
  userId: number
  name: string
  emoji: string
  shareMinor: number
}

export interface Expense {
  id: number
  eventId: number
  title: string
  amountMinor: number
  paidByUserId: number
  paidByName: string
  paidByEmoji: string
  createdByUserId: number
  createdAt: string
  updatedAt: string
  participants: ExpenseParticipant[]
}

export interface CollectorInfo {
  userId: number
  name: string
  phone: string
}

export interface MemberSummary {
  userId: number
  name: string
  emoji: string
  amountMinor: number
  paymentStatus: PaymentStatus
  isCollector: boolean
}

export interface EventSummary {
  eventId: number
  title: string
  currency: string
  status: EventStatus
  totalAmountMinor: number
  currentUserAmountMinor: number
  currentUserPaymentStatus: PaymentStatus
  currentUserRole: MemberRole
  allowMembersAddExpenses: boolean
  collector: CollectorInfo
  members: MemberSummary[]
  paidAmountMinor: number
  requiredAmountMinor: number
  paidCount: number
  membersCount: number
}

// Navigation state
export type Tab = 'expenses' | 'payment' | 'members' | 'settings'

export type Screen =
  | { type: 'list' }
  | { type: 'create' }
  | { type: 'join'; code?: string }
  | { type: 'event'; eventId: number; tab: Tab }
  | { type: 'expenseForm'; eventId: number; expenseId?: number }
  | { type: 'expenseDetail'; eventId: number; expenseId: number }
