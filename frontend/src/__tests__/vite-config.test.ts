// @vitest-environment node
import { describe, expect, it } from 'vitest'
import viteConfig from '../../vite.config'

describe('vite dev proxy config', () => {
  it('uses BANI_BACKEND_URL when running frontend outside backend localhost', () => {
    process.env.BANI_BACKEND_URL = 'http://app:8080'
    process.env.BANI_SERVER_PORT = '9999'

    const config = typeof viteConfig === 'function'
      ? viteConfig({ mode: 'development', command: 'serve' })
      : viteConfig

    expect(config.server?.proxy?.['/api']).toMatchObject({
      target: 'http://app:8080',
      changeOrigin: true,
    })
    expect(config.server?.proxy?.['/ws']).toMatchObject({
      target: 'ws://app:8080',
    })

    delete process.env.BANI_BACKEND_URL
    delete process.env.BANI_SERVER_PORT
  })
})
