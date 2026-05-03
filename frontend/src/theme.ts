import type { ThemeConfig } from 'antd'

export const appTheme: ThemeConfig = {
  token: {
    colorPrimary: '#0f766e',
    colorPrimaryHover: '#0a5f59',
    colorPrimaryActive: '#084c47',
    colorInfo: '#0f766e',
    colorSuccess: '#15803d',
    colorWarning: '#b45309',
    colorError: '#b42318',
    colorText: '#16212b',
    colorTextSecondary: '#5f6877',
    colorTextTertiary: 'rgba(22, 33, 43, 0.48)',
    colorTextDisabled: 'rgba(22, 33, 43, 0.42)',
    colorBgBase: '#f8f4ec',
    colorBgLayout: '#f8f4ec',
    colorBgContainer: '#fffdf8',
    colorBgElevated: '#fffefb',
    colorBorder: 'rgba(15, 23, 42, 0.12)',
    colorBorderSecondary: 'rgba(15, 23, 42, 0.08)',
    colorFillAlter: 'rgba(15, 118, 110, 0.06)',
    colorFillSecondary: 'rgba(15, 118, 110, 0.08)',
    colorFillTertiary: 'rgba(248, 244, 236, 0.72)',
    colorSplit: 'rgba(15, 23, 42, 0.08)',
    borderRadius: 16,
    borderRadiusSM: 12,
    borderRadiusLG: 24,
    borderRadiusXS: 8,
    controlHeight: 44,
    controlHeightSM: 36,
    controlHeightLG: 50,
    fontFamily:
      'Manrope, "Avenir Next", "Segoe UI", ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, sans-serif',
    fontSize: 14,
    fontSizeLG: 16,
    lineHeight: 1.55,
  },
  components: {
    Button: {
      borderRadius: 999,
      controlHeight: 44,
      controlHeightLG: 50,
      fontWeight: 700,
      primaryShadow: '0 14px 28px rgba(15, 118, 110, 0.20)',
    },
    Card: {
      borderRadiusLG: 28,
      paddingLG: 24,
      headerHeight: 64,
    },
    Drawer: {
      borderRadiusLG: 28,
    },
    Form: {
      itemMarginBottom: 18,
      labelColor: '#16212b',
    },
    Input: {
      borderRadius: 16,
      controlHeight: 50,
      activeBorderColor: 'rgba(15, 118, 110, 0.52)',
      hoverBorderColor: 'rgba(15, 118, 110, 0.28)',
    },
    InputNumber: {
      borderRadius: 16,
      controlHeight: 50,
      activeBorderColor: 'rgba(15, 118, 110, 0.52)',
      hoverBorderColor: 'rgba(15, 118, 110, 0.28)',
    },
    Modal: {
      borderRadiusLG: 28,
    },
    Select: {
      borderRadius: 16,
      controlHeight: 50,
      optionSelectedBg: 'rgba(15, 118, 110, 0.08)',
    },
    Segmented: {
      borderRadius: 20,
      trackBg: 'rgba(248, 244, 236, 0.94)',
      itemSelectedBg: 'rgba(255, 255, 255, 0.98)',
    },
    Table: {
      borderColor: 'rgba(15, 23, 42, 0.08)',
      headerBg: 'rgba(248, 244, 236, 0.96)',
      headerColor: '#16212b',
      rowHoverBg: 'rgba(15, 118, 110, 0.04)',
    },
    Tabs: {
      itemSelectedColor: '#0f766e',
      itemHoverColor: '#0a5f59',
      inkBarColor: '#0f766e',
    },
    Tag: {
      borderRadiusSM: 999,
    },
  },
}
