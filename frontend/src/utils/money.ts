const currencySymbols: Record<string, string> = {
  RUB: '₽',
  USD: '$',
  EUR: '€',
  GBP: '£',
  UAH: '₴',
  KZT: '₸',
}

export function formatMoney(minor: number, currency = 'RUB'): string {
  const amount = minor / 100
  const sym = currencySymbols[currency] ?? currency

  // show decimals only if needed
  const hasDecimals = minor % 100 !== 0
  const formatted = new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: hasDecimals ? 2 : 0,
    maximumFractionDigits: 2,
  }).format(amount)

  return `${formatted} ${sym}`
}

export function parseMoneyInput(value: string): number {
  // accept "100", "100.50", "100,50"
  const cleaned = value.replace(',', '.').replace(/[^\d.]/g, '')
  const num = parseFloat(cleaned)
  if (isNaN(num) || num <= 0) return 0
  return Math.round(num * 100)
}
