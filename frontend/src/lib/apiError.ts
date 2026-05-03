import type { AxiosError } from 'axios'
import type { InternalHandlerAPIResponse } from '@/api/generated/model'

type ApiErrorEnvelope = AxiosError<InternalHandlerAPIResponse>

function asAxiosError(err: unknown): ApiErrorEnvelope | null {
  if (!err || typeof err !== 'object') return null
  const candidate = err as ApiErrorEnvelope
  if (candidate.isAxiosError) return candidate
  return null
}

export function getApiErrorCode(err: unknown): string | undefined {
  return asAxiosError(err)?.response?.data?.error?.code
}

export function getApiErrorMessage(err: unknown): string | undefined {
  return asAxiosError(err)?.response?.data?.error?.message
}

export function getApiErrorStatus(err: unknown): number | undefined {
  return asAxiosError(err)?.response?.status
}
