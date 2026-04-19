import { getRoleHomePath } from '@/stores/auth'

describe('getRoleHomePath', () => {
  it('returns /catalog for client role', () => {
    expect(getRoleHomePath('client')).toBe('/catalog')
  })

  it('returns /admin for admin role', () => {
    expect(getRoleHomePath('admin')).toBe('/admin')
  })

  it('returns /dashboard for owner role', () => {
    expect(getRoleHomePath('owner')).toBe('/dashboard')
  })

  it('returns /dashboard for representative role', () => {
    expect(getRoleHomePath('representative')).toBe('/dashboard')
  })

  it('returns / for undefined role', () => {
    expect(getRoleHomePath(undefined)).toBe('/')
  })

  it('returns / for unknown role', () => {
    expect(getRoleHomePath('unknown')).toBe('/')
  })
})
