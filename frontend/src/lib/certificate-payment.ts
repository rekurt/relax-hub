export type CertificatePaymentMethod = 'card' | 'sbp' | 'apple_pay' | 'google_pay'

export const CERTIFICATE_PAYMENT_LABELS: Record<CertificatePaymentMethod, string> = {
  card: 'Банковская карта',
  sbp: 'СБП',
  apple_pay: 'Apple Pay',
  google_pay: 'Google Pay',
}
