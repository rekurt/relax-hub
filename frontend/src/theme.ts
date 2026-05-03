import type { ThemeConfig } from 'antd'

const colors = {
  primary: '#0f766e',
  primaryStrong: '#0a5f59',
  primaryPressed: '#084c47',
  accent: '#d97706',
  success: '#15803d',
  warning: '#b45309',
  error: '#b42318',
  text: '#16212b',
  textSoft: '#5f6877',
  bg: '#f8f4ec',
  container: '#fffdf8',
  elevated: '#fffefb',
  border: 'rgba(15, 23, 42, 0.08)',
  borderStrong: 'rgba(15, 23, 42, 0.12)',
  primarySoft: 'rgba(15, 118, 110, 0.08)',
}

export const appTheme: ThemeConfig = {
  token: {
    colorPrimary: colors.primary,
    colorPrimaryHover: colors.primaryStrong,
    colorPrimaryActive: colors.primaryPressed,
    colorInfo: colors.primary,
    colorSuccess: colors.success,
    colorWarning: colors.warning,
    colorError: colors.error,
    colorText: colors.text,
    colorTextSecondary: colors.textSoft,
    colorTextTertiary: 'rgba(22, 33, 43, 0.48)',
    colorTextDisabled: 'rgba(22, 33, 43, 0.42)',
    colorBgBase: colors.bg,
    colorBgLayout: colors.bg,
    colorBgContainer: colors.container,
    colorBgElevated: colors.elevated,
    colorBorder: colors.borderStrong,
    colorBorderSecondary: colors.border,
    colorFillAlter: 'rgba(248, 244, 236, 0.78)',
    colorFillSecondary: colors.primarySoft,
    colorFillTertiary: 'rgba(248, 244, 236, 0.72)',
    colorSplit: colors.border,
    borderRadius: 16,
    borderRadiusSM: 12,
    borderRadiusLG: 28,
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
    Alert: {
      borderRadiusLG: 20,
      colorInfoBg: colors.primarySoft,
      colorInfoBorder: 'rgba(15, 118, 110, 0.16)',
      colorSuccessBg: 'rgba(21, 128, 61, 0.08)',
      colorSuccessBorder: 'rgba(21, 128, 61, 0.18)',
      colorWarningBg: 'rgba(217, 119, 6, 0.10)',
      colorWarningBorder: 'rgba(217, 119, 6, 0.18)',
      colorErrorBg: 'rgba(180, 35, 24, 0.08)',
      colorErrorBorder: 'rgba(180, 35, 24, 0.16)',
    },
    Badge: {
      colorBgContainer: colors.elevated,
      colorError: colors.error,
    },
    Collapse: {
      borderRadiusLG: 20,
      headerBg: 'rgba(248, 244, 236, 0.72)',
      contentBg: 'rgba(255, 253, 248, 0.72)',
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
    Popover: {
      borderRadiusLG: 20,
    },
    Progress: {
      defaultColor: colors.primary,
      remainingColor: 'rgba(15, 23, 42, 0.08)',
    },
    Rate: {
      starColor: '#d97706',
    },
    Result: {
      titleFontSize: 22,
      subtitleFontSize: 14,
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
    Tooltip: {
      borderRadius: 12,
      colorBgSpotlight: '#16212b',
    },
    Upload: {
      borderRadiusLG: 20,
    },
  },
}
