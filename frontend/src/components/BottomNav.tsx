import type { Tab } from '../types'

interface Props {
  active: Tab
  onChange: (tab: Tab) => void
}

const tabs: { id: Tab; icon: string; label: string }[] = [
  { id: 'expenses', icon: '🧾', label: 'Расходы' },
  { id: 'payment', icon: '💸', label: 'Оплата' },
  { id: 'members', icon: '👥', label: 'Участники' },
  { id: 'settings', icon: '⚙️', label: 'Ещё' },
]

export default function BottomNav({ active, onChange }: Props) {
  return (
    <nav className="bottom-nav">
      {tabs.map((t) => (
        <button
          key={t.id}
          className={`nav-item ${active === t.id ? 'active' : ''}`}
          onClick={() => onChange(t.id)}
        >
          <span className="nav-icon">{t.icon}</span>
          <span className="nav-label">{t.label}</span>
        </button>
      ))}
    </nav>
  )
}
