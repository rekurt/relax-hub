export type CancellationPolicyDetails = { label: string; description: string }

const cancellationPolicyDetails = {
  flexible: {
    label: 'Гибкая',
    description: 'Бесплатная отмена за 24+ ч. Возврат 50% менее чем за 24 ч.',
  },
  moderate: {
    label: 'Умеренная',
    description: 'Бесплатная отмена за 72+ ч. Возврат 50% за 24-72 ч. Без возврата менее 24 ч.',
  },
  strict: {
    label: 'Строгая',
    description: 'Бесплатная отмена за 7+ дней. Возврат 50% за 3-7 дней. Без возврата менее 3 дней.',
  },
} as const satisfies Record<string, CancellationPolicyDetails>

type CancellationPolicyKey = keyof typeof cancellationPolicyDetails

function isCancellationPolicyKey(value: string): value is CancellationPolicyKey {
  return value in cancellationPolicyDetails
}

export const CANCELLATION_POLICY_DETAILS = cancellationPolicyDetails

const DEFAULT_CANCELLATION_POLICY = cancellationPolicyDetails.flexible

export function getCancellationPolicyDetails(policy?: string | null): CancellationPolicyDetails {
  if (policy && isCancellationPolicyKey(policy)) {
    return cancellationPolicyDetails[policy]
  }

  return DEFAULT_CANCELLATION_POLICY
}

export function getBookingModeLabel(mode?: string | null) {
  return mode === 'request' ? 'По запросу' : 'Мгновенное'
}

export function getBookingModeTrustCopy(mode?: string | null) {
  return mode === 'request' ? 'Подтверждение владельцем' : 'Мгновенное подтверждение'
}

export function getCancellationPolicyLabel(policy?: string | null) {
  return getCancellationPolicyDetails(policy).label
}

export function getCancellationPolicyDescription(policy?: string | null) {
  return getCancellationPolicyDetails(policy).description
}

export function getDepositSummary(percent?: number | null) {
  return percent && percent > 0 ? `${percent}% залога` : 'Без залога'
}
