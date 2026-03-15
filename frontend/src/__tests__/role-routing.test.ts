import { getRoleHomePath } from '@/stores/auth'

describe('getRoleHomePath', () => {
  it('returns /client for client role', () => {
    expect(getRoleHomePath('client')).toBe('/client')
  })

  it('returns /admin for admin role', () => {
    expect(getRoleHomePath('admin')).toBe('/admin')
  })

  it('returns / for owner role', () => {
    expect(getRoleHomePath('owner')).toBe('/')
  })

  it('returns / for representative role', () => {
    expect(getRoleHomePath('representative')).toBe('/')
  })

  it('returns / for undefined role', () => {
    expect(getRoleHomePath(undefined)).toBe('/')
  })

  it('returns / for unknown role', () => {
    expect(getRoleHomePath('unknown')).toBe('/')
  })
})
