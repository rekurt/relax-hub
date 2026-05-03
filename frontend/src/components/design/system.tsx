/* eslint-disable react-refresh/only-export-components */
import {
  Children,
  Fragment,
  createContext,
  isValidElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import {
  Checkbox,
  ColorPicker,
  Collapse,
  DatePicker,
  Descriptions,
  Drawer,
  Dropdown,
  Form,
  Input,
  InputNumber,
  Modal,
  Pagination,
  Popconfirm,
  Popover,
  QRCode,
  Radio,
  Segmented,
  Select,
  Slider,
  Steps,
  Switch,
  Table,
  Tabs,
  TimePicker,
  Tooltip,
  Upload,
} from './controls'
import type {
  AnchorHTMLAttributes,
  ButtonHTMLAttributes,
  CSSProperties,
  ImgHTMLAttributes,
  HTMLAttributes,
  MouseEvent,
  ReactNode,
} from 'react'
import DesignAvatar from './Avatar'

export {
  Checkbox,
  ColorPicker,
  Collapse,
  DatePicker,
  Descriptions,
  Drawer,
  Dropdown,
  Form,
  Input,
  InputNumber,
  Modal,
  Pagination,
  Popconfirm,
  Popover,
  QRCode,
  Radio,
  Segmented,
  Select,
  Slider,
  Steps,
  Switch,
  Table,
  Tabs,
  TimePicker,
  Tooltip,
  Upload,
}

export const ruRU = {
  locale: 'ru',
  global: {
    placeholder: 'Выберите',
  },
}

type SizeToken = 'small' | 'middle' | 'large'
type ToneType = 'secondary' | 'success' | 'warning' | 'danger'

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

function toGap(size: unknown): string {
  if (Array.isArray(size)) {
    return size.map((item) => (typeof item === 'number' ? `${item}px` : String(item))).join(' ')
  }
  if (typeof size === 'number') return `${size}px`
  if (size === 'large') return '24px'
  if (size === 'middle') return '16px'
  return '8px'
}

interface ConfigProviderProps {
  children?: ReactNode
  locale?: unknown
  theme?: unknown
}

export function ConfigProvider({ children }: ConfigProviderProps) {
  return <>{children}</>
}

type NoticeType = 'success' | 'error' | 'warning' | 'info' | 'loading'

interface Notice {
  id: number
  type: NoticeType
  content: ReactNode
}

interface ConfirmOptions {
  title?: ReactNode
  content?: ReactNode
  icon?: ReactNode
  okType?: string
  okText?: ReactNode
  cancelText?: ReactNode
  okButtonProps?: ButtonHTMLAttributes<HTMLButtonElement> & { danger?: boolean }
  cancelButtonProps?: ButtonHTMLAttributes<HTMLButtonElement>
  onOk?: () => unknown | Promise<unknown>
  onCancel?: () => void
}

interface AppApi {
  message: Record<NoticeType, (content: ReactNode) => () => void> & {
    open: (config: { type?: NoticeType; content?: ReactNode }) => () => void
    destroy: () => void
  }
  modal: {
    confirm: (config: ConfirmOptions) => { destroy: () => void; update: (next: ConfirmOptions) => void }
  }
  notification: {
    open: (config: { type?: NoticeType; message?: ReactNode; description?: ReactNode }) => () => void
  }
}

const noopDestroy = () => undefined

const fallbackAppApi: AppApi = {
  message: Object.assign(
    {
      success: () => noopDestroy,
      error: () => noopDestroy,
      warning: () => noopDestroy,
      info: () => noopDestroy,
      loading: () => noopDestroy,
    },
    {
      open: () => noopDestroy,
      destroy: noopDestroy,
    },
  ),
  modal: {
    confirm: (config) => {
      const accepted = typeof window === 'undefined' || window.confirm(String(config.title ?? config.content ?? 'Подтвердите действие'))
      if (accepted) void config.onOk?.()
      else config.onCancel?.()
      return { destroy: noopDestroy, update: noopDestroy }
    },
  },
  notification: {
    open: () => noopDestroy,
  },
}

const AppContext = createContext<AppApi>(fallbackAppApi)

interface AppProps {
  children?: ReactNode
}

function AppRoot({ children }: AppProps) {
  const [notices, setNotices] = useState<Notice[]>([])
  const [confirmDialog, setConfirmDialog] = useState<(ConfirmOptions & { id: number }) | null>(null)

  const removeNotice = useCallback((id: number) => {
    setNotices((current) => current.filter((notice) => notice.id !== id))
  }, [])

  const pushNotice = useCallback((type: NoticeType, content: ReactNode) => {
    const id = Date.now() + Math.random()
    setNotices((current) => [...current.slice(-3), { id, type, content }])
    if (type !== 'loading') {
      window.setTimeout(() => removeNotice(id), 3200)
    }
    return () => removeNotice(id)
  }, [removeNotice])

  const api = useMemo<AppApi>(() => {
    const openMessage = (type: NoticeType, content: ReactNode) => pushNotice(type, content)
    return {
      message: Object.assign(
        {
          success: (content: ReactNode) => openMessage('success', content),
          error: (content: ReactNode) => openMessage('error', content),
          warning: (content: ReactNode) => openMessage('warning', content),
          info: (content: ReactNode) => openMessage('info', content),
          loading: (content: ReactNode) => openMessage('loading', content),
        },
        {
          open: (config: { type?: NoticeType; content?: ReactNode }) => openMessage(config.type ?? 'info', config.content),
          destroy: () => setNotices([]),
        },
      ),
      modal: {
        confirm: (config: ConfirmOptions) => {
          const id = Date.now() + Math.random()
          setConfirmDialog({ ...config, id })
          return {
            destroy: () => setConfirmDialog((current) => (current?.id === id ? null : current)),
            update: (next: ConfirmOptions) => setConfirmDialog((current) => (current?.id === id ? { ...current, ...next } : current)),
          }
        },
      },
      notification: {
        open: (config) => openMessage(config.type ?? 'info', (
          <span>
            {config.message}
            {config.description && <span className="rh-toast__description">{config.description}</span>}
          </span>
        )),
      },
    }
  }, [pushNotice])

  const closeConfirm = useCallback(() => setConfirmDialog(null), [])
  const confirmOk = useCallback(async () => {
    const active = confirmDialog
    if (!active) return
    await active.onOk?.()
    closeConfirm()
  }, [closeConfirm, confirmDialog])
  const {
    type: cancelHtmlType,
    ...cancelButtonRest
  } = confirmDialog?.cancelButtonProps ?? {}
  const {
    type: okHtmlType,
    danger: okDanger,
    ...okButtonRest
  } = confirmDialog?.okButtonProps ?? {}

  return (
    <AppContext.Provider value={api}>
      {children}
      {notices.length > 0 && (
        <div className="rh-toast-stack" aria-live="polite" aria-atomic="true">
          {notices.map((notice) => (
            <div key={notice.id} className={cx('rh-toast', `rh-toast--${notice.type}`)}>
              <span className="rh-toast__dot" aria-hidden="true" />
              <span className="rh-toast__content">{notice.content}</span>
            </div>
          ))}
        </div>
      )}
      {confirmDialog && (
        <div className="rh-modal-root">
          <button className="rh-modal-mask" aria-label="Закрыть окно" type="button" onClick={() => {
            confirmDialog.onCancel?.()
            closeConfirm()
          }} />
          <div className="rh-confirm-modal ant-modal" role="dialog" aria-modal="true">
            <div className="rh-confirm-modal__content ant-modal-content">
              {confirmDialog.title && (
                <div className="rh-confirm-modal__header ant-modal-header">
                  <div className="rh-confirm-modal__title ant-modal-title">{confirmDialog.title}</div>
                </div>
              )}
              {confirmDialog.content && <div className="rh-confirm-modal__body ant-modal-body">{confirmDialog.content}</div>}
              <div className="rh-confirm-modal__footer ant-modal-footer">
                <Button {...cancelButtonRest} htmlType={cancelHtmlType} onClick={() => {
                  confirmDialog.onCancel?.()
                  closeConfirm()
                }}>
                  {confirmDialog.cancelText ?? 'Отмена'}
                </Button>
                <Button
                  {...okButtonRest}
                  type="primary"
                  htmlType={okHtmlType}
                  danger={confirmDialog.okType === 'danger' || okDanger}
                  onClick={() => void confirmOk()}
                >
                  {confirmDialog.okText ?? 'OK'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </AppContext.Provider>
  )
}

export const App = Object.assign(AppRoot, {
  useApp: () => useContext(AppContext),
})

function useMediaScreens() {
  const readScreens = useCallback(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return {} as Record<string, boolean>
    }
    return {
      xs: window.matchMedia('(max-width: 575px)').matches,
      sm: window.matchMedia('(min-width: 576px)').matches,
      md: window.matchMedia('(min-width: 768px)').matches,
      lg: window.matchMedia('(min-width: 992px)').matches,
      xl: window.matchMedia('(min-width: 1200px)').matches,
      xxl: window.matchMedia('(min-width: 1600px)').matches,
    }
  }, [])
  const [screens, setScreens] = useState<Record<string, boolean>>(readScreens)

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return undefined
    const queries = [
      window.matchMedia('(max-width: 575px)'),
      window.matchMedia('(min-width: 576px)'),
      window.matchMedia('(min-width: 768px)'),
      window.matchMedia('(min-width: 992px)'),
      window.matchMedia('(min-width: 1200px)'),
      window.matchMedia('(min-width: 1600px)'),
    ]
    const sync = () => setScreens(readScreens())
    queries.forEach((query) => query.addEventListener('change', sync))
    sync()
    return () => queries.forEach((query) => query.removeEventListener('change', sync))
  }, [readScreens])

  return screens
}

export const Grid = {
  useBreakpoint: useMediaScreens,
}

interface LayoutProps extends HTMLAttributes<HTMLElement> {
  hasSider?: boolean
}

function LayoutRoot({ hasSider, className, children, ...props }: LayoutProps) {
  return (
    <section className={cx('rh-layout', 'ant-layout', hasSider && 'ant-layout-has-sider', className)} {...props}>
      {children}
    </section>
  )
}

function LayoutHeader({ className, children, ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <header className={cx('rh-layout__header', 'ant-layout-header', className)} {...props}>
      {children}
    </header>
  )
}

function LayoutContent({ className, children, ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <main className={cx('rh-layout__content', 'ant-layout-content', className)} {...props}>
      {children}
    </main>
  )
}

function LayoutFooter({ className, children, ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <footer className={cx('rh-layout__footer', 'ant-layout-footer', className)} {...props}>
      {children}
    </footer>
  )
}

function LayoutSider({ className, children, ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <aside className={cx('rh-layout__sider', 'ant-layout-sider', className)} {...props}>
      {children}
    </aside>
  )
}

export const Layout = Object.assign(LayoutRoot, {
  Header: LayoutHeader,
  Content: LayoutContent,
  Footer: LayoutFooter,
  Sider: LayoutSider,
})

interface FlexProps extends HTMLAttributes<HTMLDivElement> {
  vertical?: boolean
  gap?: CSSProperties['gap']
  align?: CSSProperties['alignItems']
  justify?: CSSProperties['justifyContent']
  wrap?: CSSProperties['flexWrap'] | boolean
}

export function Flex({ vertical, gap, align, justify, wrap, className, style, children, ...props }: FlexProps) {
  return (
    <div
      className={cx('rh-flex', 'ant-flex', vertical && 'rh-flex--vertical', className)}
      style={{
        display: 'flex',
        flexDirection: vertical ? 'column' : undefined,
        gap,
        alignItems: align,
        justifyContent: justify,
        flexWrap: wrap === true ? 'wrap' : wrap || undefined,
        ...style,
      }}
      {...props}
    >
      {children}
    </div>
  )
}

interface CarouselProps extends HTMLAttributes<HTMLDivElement> {
  autoplay?: boolean
  autoplaySpeed?: number
  dots?: boolean
}

export function Carousel({ autoplay, autoplaySpeed = 5000, dots, className, children, ...props }: CarouselProps) {
  const slides = Children.toArray(children)
  const [active, setActive] = useState(0)

  useEffect(() => {
    if (!autoplay || slides.length <= 1) return undefined
    const timer = window.setInterval(() => {
      setActive((current) => (current + 1) % slides.length)
    }, autoplaySpeed)
    return () => window.clearInterval(timer)
  }, [autoplay, autoplaySpeed, slides.length])

  if (slides.length === 0) return null

  return (
    <div className={cx('rh-carousel', 'ant-carousel', className)} {...props}>
      <div className="rh-carousel__viewport">
        {slides.map((slide, index) => (
          <div key={index} className={cx('rh-carousel__slide', index === active && 'rh-carousel__slide--active')}>
            {slide}
          </div>
        ))}
      </div>
      {dots && slides.length > 1 && (
        <div className="rh-carousel__dots">
          {slides.map((_, index) => (
            <button
              key={index}
              type="button"
              className={cx('rh-carousel__dot', index === active && 'rh-carousel__dot--active')}
              aria-label={`Слайд ${index + 1}`}
              onClick={() => setActive(index)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

interface ListProps<T> extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  dataSource?: T[]
  renderItem?: (item: T, index: number) => ReactNode
  loading?: boolean
  locale?: { emptyText?: ReactNode }
  size?: SizeToken | string
  grid?: { gutter?: number; xs?: number; sm?: number; md?: number; lg?: number; xl?: number }
  pagination?: false | {
    current?: number
    pageSize?: number
    total?: number
    onChange?: (page: number, pageSize: number) => void
    showSizeChanger?: boolean
    showTotal?: (total: number) => ReactNode
  }
  children?: ReactNode
}

interface ListItemProps extends HTMLAttributes<HTMLDivElement> {
  actions?: ReactNode[]
  extra?: ReactNode
}

interface ListMetaProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  avatar?: ReactNode
  title?: ReactNode
  description?: ReactNode
}

function ListMeta({ avatar, title, description, className, ...props }: ListMetaProps) {
  return (
    <div className={cx('rh-list-meta', 'ant-list-item-meta', className)} {...props}>
      {avatar && <div className="rh-list-meta__avatar ant-list-item-meta-avatar">{avatar}</div>}
      <div className="rh-list-meta__content ant-list-item-meta-content">
        {title && <div className="rh-list-meta__title ant-list-item-meta-title">{title}</div>}
        {description && <div className="rh-list-meta__description ant-list-item-meta-description">{description}</div>}
      </div>
    </div>
  )
}

function ListItem({ actions, extra, className, children, ...props }: ListItemProps) {
  return (
    <div className={cx('rh-list-item', 'ant-list-item', className)} {...props}>
      <div className="rh-list-item__main">{children}</div>
      {extra && <div className="rh-list-item__extra">{extra}</div>}
      {actions && actions.length > 0 && (
        <div className="rh-list-item__actions ant-list-item-action">
          {actions.map((action, index) => (
            <span key={index} className="rh-list-item__action">
              {action}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

export const List = Object.assign(
  function ListRoot<T>({
    dataSource,
    renderItem,
    loading,
    locale,
    size,
    grid,
    pagination,
    className,
    children,
    ...props
  }: ListProps<T>) {
    const items = dataSource?.map((item, index) => renderItem?.(item, index) ?? null) ?? Children.toArray(children)
    const gridColumns = grid?.lg ?? grid?.md ?? grid?.sm ?? grid?.xs
    return (
      <div className={cx('rh-list', 'ant-list', size === 'small' && 'rh-list--small ant-list-sm', grid && 'rh-list--grid', className)} {...props}>
        {loading ? (
          <Spin />
        ) : items.length > 0 ? (
          <div
            className="rh-list__items"
            style={grid ? { gridTemplateColumns: `repeat(${gridColumns ?? 1}, minmax(0, 1fr))`, gap: grid.gutter ?? 16 } : undefined}
          >
            {items.map((item, index) => (
              <Fragment key={isValidElement(item) ? item.key ?? index : index}>
                {item}
              </Fragment>
            ))}
          </div>
        ) : (
          <div className="rh-list__empty">{locale?.emptyText ?? <Empty description="Нет данных" />}</div>
        )}
        {pagination && (
          <div className="rh-list__pagination">
            <Pagination {...pagination} />
          </div>
        )}
      </div>
    )
  },
  { Item: Object.assign(ListItem, { Meta: ListMeta }) },
)

interface ImageProps extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'placeholder'> {
  preview?: boolean | {
    visible?: boolean
    src?: string
    onVisibleChange?: (visible: boolean) => void
  }
  fallback?: string
  placeholder?: ReactNode
}

function PreviewGroup({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cx('rh-image-preview-group', 'ant-image-preview-group', className)} {...props}>
      {children}
    </div>
  )
}

export const Image = Object.assign(
  function ImageRoot({ preview = true, fallback, placeholder, className, style, src, alt, onError, ...props }: ImageProps) {
    const [internalPreview, setInternalPreview] = useState(false)
    const [resolvedSrc, setResolvedSrc] = useState(src)
    const previewConfig = typeof preview === 'object' ? preview : undefined
    const previewVisible = previewConfig?.visible ?? internalPreview
    const previewSrc = previewConfig?.src ?? resolvedSrc
    const canPreview = preview !== false && Boolean(previewSrc)
    const setPreviewVisible = (visible: boolean) => {
      if (previewConfig?.onVisibleChange) previewConfig.onVisibleChange(visible)
      else setInternalPreview(visible)
    }

    useEffect(() => {
      setResolvedSrc(src)
    }, [src])

    return (
      <>
        <span className={cx('rh-image', 'ant-image', className)} style={style}>
          {resolvedSrc ? (
            <img
              className="rh-image__img ant-image-img"
              src={resolvedSrc}
              alt={alt}
              onClick={() => canPreview && setPreviewVisible(true)}
              onError={(event) => {
                if (fallback) setResolvedSrc(fallback)
                onError?.(event)
              }}
              {...props}
            />
          ) : placeholder ? (
            <span className="rh-image__placeholder">{placeholder}</span>
          ) : null}
        </span>
        {previewVisible && previewSrc && (
          <div className="rh-image-preview ant-image-preview-root" role="dialog" aria-modal="true">
            <button type="button" className="rh-image-preview__mask" aria-label="Закрыть просмотр" onClick={() => setPreviewVisible(false)} />
            <img className="rh-image-preview__img" src={previewSrc} alt={alt} />
            <button type="button" className="rh-image-preview__close" aria-label="Закрыть просмотр" onClick={() => setPreviewVisible(false)}>
              ×
            </button>
          </div>
        )}
      </>
    )
  },
  { PreviewGroup },
)

type ButtonKind = 'primary' | 'default' | 'dashed' | 'link' | 'text'

interface ButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'type'> {
  type?: ButtonKind
  htmlType?: ButtonHTMLAttributes<HTMLButtonElement>['type']
  href?: string
  target?: AnchorHTMLAttributes<HTMLAnchorElement>['target']
  rel?: AnchorHTMLAttributes<HTMLAnchorElement>['rel']
  icon?: ReactNode
  loading?: boolean
  danger?: boolean
  block?: boolean
  size?: SizeToken
  shape?: 'default' | 'circle' | 'round'
  ghost?: boolean
}

export function Button({
  type = 'default',
  htmlType = 'button',
  href,
  target,
  rel,
  icon,
  loading = false,
  danger = false,
  block = false,
  size = 'middle',
  shape = 'default',
  ghost = false,
  className,
  children,
  disabled,
  onClick,
  ...props
}: ButtonProps) {
  const isPrimary = type === 'primary'
  const isText = type === 'text'
  const isLink = type === 'link'
  const isIconOnly = Boolean(icon) && children == null
  const classes = cx(
    'rh-system-button',
    'rh-btn',
    `rh-btn--${isPrimary && !ghost ? 'primary' : isText || isLink || ghost ? 'ghost' : 'default'}`,
    size === 'small' && 'rh-btn--sm',
    size === 'large' && 'rh-btn--lg',
    block && 'rh-btn--block',
    danger && 'rh-btn--danger',
    loading && 'rh-btn--loading',
    isText && 'rh-btn--text',
    isLink && 'rh-btn--link',
    shape !== 'default' && `rh-btn--${shape}`,
    'ant-btn',
    isPrimary && 'ant-btn-primary',
    ghost && 'ant-btn-background-ghost',
    !isPrimary && !isText && !isLink && 'ant-btn-default',
    isText && 'ant-btn-text',
    isLink && 'ant-btn-link',
    danger && 'ant-btn-dangerous',
    loading && 'ant-btn-loading',
    size === 'small' && 'ant-btn-sm',
    size === 'large' && 'ant-btn-lg',
    isIconOnly && 'ant-btn-icon-only',
    className,
  )
  const content = (
    <>
      {loading && <span className="rh-btn__loader" aria-hidden="true" />}
      {icon && <span className="ant-btn-icon rh-btn__icon">{icon}</span>}
      {children && <span>{children}</span>}
    </>
  )

  if (href) {
    return (
      <a
        className={classes}
        href={disabled ? undefined : href}
        target={target}
        rel={rel}
        aria-disabled={disabled || loading || undefined}
        onClick={(event) => {
          if (disabled || loading) event.preventDefault()
        }}
      >
        {content}
      </a>
    )
  }

  return (
    <button
      className={classes}
      type={htmlType}
      disabled={disabled || loading}
      onClick={(event) => {
        if (loading) return
        onClick?.(event)
      }}
      {...props}
    >
      {content}
    </button>
  )
}

interface CardProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: ReactNode
  extra?: ReactNode
  cover?: ReactNode
  actions?: ReactNode[]
  loading?: boolean
  hoverable?: boolean
  bordered?: boolean
  variant?: 'borderless' | 'outlined'
  size?: SizeToken
  styles?: {
    header?: CSSProperties
    body?: CSSProperties
  }
  bodyStyle?: CSSProperties
}

interface CardMetaProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  avatar?: ReactNode
  title?: ReactNode
  description?: ReactNode
}

function CardMeta({ avatar, title, description, className, ...props }: CardMetaProps) {
  return (
    <div className={cx('rh-card-meta', 'ant-card-meta', className)} {...props}>
      {avatar && <div className="rh-card-meta__avatar ant-card-meta-avatar">{avatar}</div>}
      <div className="rh-card-meta__detail ant-card-meta-detail">
        {title && <div className="rh-card-meta__title ant-card-meta-title">{title}</div>}
        {description && <div className="rh-card-meta__description ant-card-meta-description">{description}</div>}
      </div>
    </div>
  )
}

export const Card = Object.assign(
  function Card({
    title,
    extra,
    cover,
    actions,
    loading = false,
    hoverable = false,
    bordered = true,
    variant,
    size = 'middle',
    styles,
    bodyStyle,
    className,
    children,
    ...props
  }: CardProps) {
    return (
      <div
        className={cx(
          'rh-system-card',
          'rh-card',
          'ant-card',
          size === 'small' && 'ant-card-small rh-card--small',
          hoverable && 'rh-card--hoverable ant-card-hoverable',
          (!bordered || variant === 'borderless') && 'rh-card--borderless ant-card-borderless',
          className,
        )}
        {...props}
      >
        {cover && <div className="rh-card__cover ant-card-cover">{cover}</div>}
        {(title || extra) && (
          <div className="rh-card__head ant-card-head" style={styles?.header}>
            <div className="rh-card__head-title ant-card-head-title">{title}</div>
            {extra && <div className="rh-card__extra ant-card-extra">{extra}</div>}
          </div>
        )}
        <div className="rh-card__body ant-card-body" style={{ ...bodyStyle, ...styles?.body }}>
          {loading ? <Skeleton active paragraph={{ rows: 2 }} /> : children}
        </div>
        {actions && actions.length > 0 && (
          <ul className="rh-card__actions ant-card-actions">
            {actions.map((action, index) => (
              <li key={String(index)}>{action}</li>
            ))}
          </ul>
        )}
      </div>
    )
  },
  { Meta: CardMeta },
)

type ColorTone = 'default' | 'success' | 'processing' | 'error' | 'warning' | string

interface TagProps extends HTMLAttributes<HTMLSpanElement> {
  color?: ColorTone
  icon?: ReactNode
  closable?: boolean
  closeIcon?: ReactNode
  onClose?: (event: MouseEvent<HTMLElement>) => void
}

function tagTone(color?: ColorTone) {
  if (!color || color === 'default') return 'default'
  if (['success', 'green'].includes(color)) return 'green'
  if (['error', 'red', 'volcano', 'magenta'].includes(color)) return 'red'
  if (['warning', 'orange', 'gold', 'yellow'].includes(color)) return 'gold'
  if (['processing', 'blue', 'cyan', 'geekblue', 'purple'].includes(color)) return 'primary'
  return 'ghost'
}

export function Tag({ color, icon, closable, closeIcon, onClose, className, children, ...props }: TagProps) {
  const tone = tagTone(color)
  return (
    <span
      className={cx(
        'rh-tag',
        tone !== 'default' && `rh-tag--${tone}`,
        'ant-tag',
        color && `ant-tag-${color}`,
        className,
      )}
      {...props}
    >
      {icon}
      {children}
      {closable && (
        <button
          type="button"
          className="rh-tag__close"
          aria-label="close"
          onClick={(event) => {
            event.stopPropagation()
            onClose?.(event)
          }}
        >
          {closeIcon ?? '×'}
        </button>
      )}
    </span>
  )
}

interface TypographyBaseProps extends HTMLAttributes<HTMLElement> {
  type?: ToneType
  strong?: boolean
  disabled?: boolean
  code?: boolean
  ellipsis?: boolean | { rows?: number }
  copyable?: boolean | { text?: string; tooltips?: boolean }
}

interface TitleProps extends TypographyBaseProps {
  level?: 1 | 2 | 3 | 4 | 5
}

function typographyClass(type?: ToneType, strong?: boolean, disabled?: boolean, className?: string) {
  return cx(
    'rh-typography',
    'ant-typography',
    type === 'secondary' && 'rh-typography--secondary ant-typography-secondary',
    type === 'success' && 'rh-typography--success',
    type === 'warning' && 'rh-typography--warning',
    type === 'danger' && 'rh-typography--danger',
    strong && 'rh-typography--strong',
    disabled && 'rh-typography--disabled',
    className,
  )
}

function Title({ level = 1, type, strong, disabled, className, children, ...props }: TitleProps) {
  const Component = `h${level}` as 'h1'
  return (
    <Component className={typographyClass(type, strong, disabled, cx('rh-title', `rh-title--h${level}`, className))} {...props}>
      {children}
    </Component>
  )
}

function Text({ type, strong, disabled, code, ellipsis, copyable, className, children, ...props }: TypographyBaseProps) {
  const Component = code ? 'code' : 'span'
  return (
    <Component
      className={typographyClass(type, strong, disabled, cx(ellipsis && 'rh-typography--ellipsis', copyable && 'rh-typography--copyable', className))}
      {...props}
    >
      {children}
    </Component>
  )
}

function Paragraph({ type, strong, disabled, ellipsis, copyable, className, children, ...props }: TypographyBaseProps) {
  return (
    <p
      className={typographyClass(type, strong, disabled, cx('ant-typography-paragraph', ellipsis && 'rh-typography--ellipsis', copyable && 'rh-typography--copyable', className))}
      {...props}
    >
      {children}
    </p>
  )
}

export const Typography = Object.assign(
  function TypographyRoot({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
    return (
      <div className={cx('rh-typography-block', 'ant-typography', className)} {...props}>
        {children}
      </div>
    )
  },
  { Title, Text, Paragraph },
)

interface SpaceProps extends HTMLAttributes<HTMLDivElement> {
  direction?: 'horizontal' | 'vertical' | string
  orientation?: 'horizontal' | 'vertical' | string
  size?: unknown
  align?: CSSProperties['alignItems'] | string
  wrap?: boolean
}

function SpaceRoot({ direction, orientation, size = 'small', align, wrap, className, style, children, ...props }: SpaceProps) {
  const resolvedDirection = direction ?? orientation ?? 'horizontal'
  return (
    <div
      className={cx('rh-space', 'ant-space', resolvedDirection === 'vertical' && 'rh-space--vertical ant-space-vertical', wrap && 'rh-space--wrap', className)}
      style={{ gap: toGap(size), alignItems: align, flexWrap: wrap ? 'wrap' : undefined, ...style }}
      {...props}
    >
      {children}
    </div>
  )
}

function Compact({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cx('rh-space-compact', 'ant-space-compact', className)} {...props}>
      {children}
    </div>
  )
}

export const Space = Object.assign(SpaceRoot, { Compact })

interface RowProps extends HTMLAttributes<HTMLDivElement> {
  gutter?: number | [number, number]
  align?: CSSProperties['alignItems']
  justify?: CSSProperties['justifyContent']
}

export function Row({ gutter = 0, align, justify, className, style, children, ...props }: RowProps) {
  const gap = Array.isArray(gutter) ? `${gutter[1]}px ${gutter[0]}px` : `${gutter}px`
  return (
    <div className={cx('rh-row', 'ant-row', className)} style={{ gap, alignItems: align, justifyContent: justify, ...style }} {...props}>
      {children}
    </div>
  )
}

interface ColProps extends HTMLAttributes<HTMLDivElement> {
  span?: number
  xs?: number
  sm?: number
  md?: number
  lg?: number
  xl?: number
  flex?: string | number
}

export function Col({ span, xs, sm, md, lg, xl, flex, className, style, children, ...props }: ColProps) {
  const basis = span ?? xs ?? sm ?? md ?? lg ?? xl ?? 24
  const width = `calc(${Math.min(24, basis) / 24 * 100}% - 0.01px)`
  return (
    <div
      className={cx(
        'rh-col',
        'ant-col',
        span != null && `ant-col-${span}`,
        xs != null && `ant-col-xs-${xs}`,
        sm != null && `ant-col-sm-${sm}`,
        md != null && `ant-col-md-${md}`,
        lg != null && `ant-col-lg-${lg}`,
        xl != null && `ant-col-xl-${xl}`,
        className,
      )}
      style={{ flex: flex ?? `1 1 ${width}`, maxWidth: flex ? undefined : width, ...style }}
      {...props}
    >
      {children}
    </div>
  )
}

interface SpinProps extends HTMLAttributes<HTMLDivElement> {
  spinning?: boolean
  size?: SizeToken
  tip?: ReactNode
}

export function Spin({ spinning = true, size = 'middle', tip, className, children, ...props }: SpinProps) {
  if (children) {
    return (
      <div className={cx('rh-spin-container', spinning && 'rh-spin-container--loading')} {...props}>
        {spinning && <Spin size={size} tip={tip} />}
        <div className={spinning ? 'rh-spin-blur' : undefined}>{children}</div>
      </div>
    )
  }

  if (!spinning) return null

  return (
    <div className={cx('rh-spin', 'rh-spin-spinning', 'ant-spin', 'ant-spin-spinning', size === 'large' && 'rh-spin--lg', size === 'small' && 'rh-spin--sm', className)} {...props}>
      <span className="rh-spin__dot ant-spin-dot">
        <i className="ant-spin-dot-item" />
        <i className="ant-spin-dot-item" />
        <i className="ant-spin-dot-item" />
        <i className="ant-spin-dot-item" />
      </span>
      {tip && <span className="rh-spin__tip">{tip}</span>}
    </div>
  )
}

interface EmptyProps extends HTMLAttributes<HTMLDivElement> {
  description?: ReactNode
  image?: ReactNode
}

const PRESENTED_IMAGE_SIMPLE = (
  <svg viewBox="0 0 120 96" width="120" height="96" aria-hidden="true" className="rh-empty__art">
    <rect x="18" y="24" width="84" height="52" rx="16" />
    <path d="M30 42h60M38 56h44" />
    <circle cx="88" cy="30" r="10" />
  </svg>
)

export const Empty = Object.assign(
  function Empty({ description = 'Нет данных', image, className, children, ...props }: EmptyProps) {
    return (
      <div className={cx('rh-empty', 'ant-empty', 'ant-empty-normal', className)} {...props}>
        <div className="rh-empty__image ant-empty-image">{image ?? PRESENTED_IMAGE_SIMPLE}</div>
        <div className="rh-empty__description ant-empty-description">{description}</div>
        {children && <div className="rh-empty__footer ant-empty-footer">{children}</div>}
      </div>
    )
  },
  { PRESENTED_IMAGE_SIMPLE },
)

interface AlertProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  type?: 'success' | 'info' | 'warning' | 'error'
  message?: ReactNode
  title?: ReactNode
  description?: ReactNode
  showIcon?: boolean
  icon?: ReactNode
  action?: ReactNode
  closable?: boolean
  onClose?: () => void
}

export function Alert({ type = 'info', message, title, description, showIcon, icon, action, closable, onClose, className, children, ...props }: AlertProps) {
  return (
    <div className={cx('rh-alert', `rh-alert--${type}`, 'ant-alert', `ant-alert-${type}`, className)} role={type === 'error' ? 'alert' : 'status'} {...props}>
      {showIcon && <span className="rh-alert__icon" aria-hidden="true">{icon}</span>}
      <div className="rh-alert__content">
        {(message || title) && <div className="rh-alert__message ant-alert-message">{message ?? title}</div>}
        {(description || children) && <div className="rh-alert__description ant-alert-description">{description ?? children}</div>}
      </div>
      {action && <div className="rh-alert__action">{action}</div>}
      {closable && (
        <button type="button" className="rh-alert__close" aria-label="close" onClick={onClose}>
          ×
        </button>
      )}
    </div>
  )
}

interface ResultProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  status?: string
  icon?: ReactNode
  title?: ReactNode
  subTitle?: ReactNode
  extra?: ReactNode
}

export function Result({ status = 'info', icon, title, subTitle, extra, className, children, ...props }: ResultProps) {
  return (
    <div className={cx('rh-result', `rh-result--${status}`, 'ant-result', className)} {...props}>
      <div className="rh-result__icon" aria-hidden="true">{icon}</div>
      {title && <div className="rh-result__title ant-result-title">{title}</div>}
      {subTitle && <div className="rh-result__subtitle ant-result-subtitle">{subTitle}</div>}
      {children}
      {extra && <div className="rh-result__extra">{extra}</div>}
    </div>
  )
}

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  count?: ReactNode
  dot?: boolean
  showZero?: boolean
  status?: 'success' | 'processing' | 'default' | 'error' | 'warning'
  text?: ReactNode
  offset?: [number, number]
  size?: SizeToken
}

export function Badge({ count, dot, showZero = false, status, text, offset, size, className, children, ...props }: BadgeProps) {
  if (status) {
    return (
      <span className={cx('rh-badge-status', `rh-badge-status--${status}`, 'ant-badge-status', className)} {...props}>
        <span className={cx('rh-badge-status__dot', 'ant-badge-status-dot', `ant-badge-status-${status}`)} />
        {text && <span className="rh-badge-status__text ant-badge-status-text">{text}</span>}
      </span>
    )
  }

  return (
    <span className={cx('rh-badge', 'ant-badge', className)} {...props}>
      {children}
      {(dot || (count != null && (showZero || count !== 0))) && (
        <sup
          className={cx('rh-badge__count', 'ant-badge-count', dot && 'rh-badge__dot', size === 'small' && 'rh-badge__count--sm')}
          style={offset ? { transform: `translate(${offset[0]}px, ${offset[1]}px)` } : undefined}
        >
          {dot ? '' : count}
        </sup>
      )}
    </span>
  )
}

interface RateProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value?: number
  defaultValue?: number
  count?: number
  allowHalf?: boolean
  disabled?: boolean
  onChange?: (value: number) => void
  tooltips?: string[]
}

export function Rate({ value, defaultValue = 0, count = 5, allowHalf: _allowHalf, tooltips: _tooltips, disabled, onChange, className, style, ...props }: RateProps) {
  const current = value ?? defaultValue
  return (
    <div className={cx('rh-rate', 'ant-rate', disabled && 'ant-rate-disabled', className)} style={style} role="radiogroup" aria-label={`${current} из ${count}`} {...props}>
      {Array.from({ length: count }, (_, index) => {
        const filled = current >= index + 1
        const half = !filled && current > index
        return (
          <button
            key={index}
            type="button"
            role="radio"
            aria-checked={current === index + 1}
            disabled={disabled}
            className={cx('rh-rate__star', filled && 'ant-rate-star-full', half && 'ant-rate-star-half')}
            onClick={() => onChange?.(index + 1)}
            aria-label={`${index + 1}`}
          >
            ★
          </button>
        )
      })}
    </div>
  )
}

interface ProgressProps extends HTMLAttributes<HTMLDivElement> {
  type?: string
  percent?: number
  showInfo?: boolean
  size?: SizeToken | number
  status?: string
  strokeColor?: string
  format?: (percent?: number) => ReactNode
}

export function Progress({ percent = 0, showInfo = true, size, status, strokeColor, format, className, ...props }: ProgressProps) {
  const safePercent = Math.max(0, Math.min(100, percent))
  return (
    <div className={cx('rh-progress', 'ant-progress', size === 'small' && 'rh-progress--sm', status && `rh-progress--${status}`, className)} {...props}>
      <div className="rh-progress__rail ant-progress-outer">
        <div className="rh-progress__bar ant-progress-bg" style={{ width: `${safePercent}%`, background: strokeColor }} />
      </div>
      {showInfo && <span className="rh-progress__text ant-progress-text">{format ? format(safePercent) : `${safePercent}%`}</span>}
    </div>
  )
}

interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  active?: boolean
  paragraph?: { rows?: number } | boolean
  avatar?: boolean
}

export function Skeleton({ active, paragraph = true, avatar, className, ...props }: SkeletonProps) {
  const rows = typeof paragraph === 'object' ? paragraph.rows ?? 3 : paragraph ? 3 : 0
  return (
    <div className={cx('rh-skeleton', active && 'rh-skeleton--active', 'ant-skeleton', className)} {...props}>
      {avatar && <span className="rh-skeleton__avatar ant-skeleton-avatar" />}
      <div className="rh-skeleton-content ant-skeleton-content">
        <div className="rh-skeleton__title ant-skeleton-title" />
        {rows > 0 && (
          <ul className="rh-skeleton__paragraph ant-skeleton-paragraph">
            {Array.from({ length: rows }, (_, index) => <li key={index} />)}
          </ul>
        )}
      </div>
    </div>
  )
}

interface StatisticProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title' | 'prefix'> {
  title?: ReactNode
  value?: ReactNode
  prefix?: ReactNode
  suffix?: ReactNode
  precision?: number
  valueStyle?: CSSProperties
  styles?: {
    content?: CSSProperties
    value?: CSSProperties
  }
  formatter?: (value?: ReactNode) => ReactNode
}

export function Statistic({ title, value, prefix, suffix, precision, valueStyle, styles, formatter, className, ...props }: StatisticProps) {
  const rawValue = typeof value === 'number' && precision != null ? value.toFixed(precision) : value
  const displayValue = formatter ? formatter(rawValue) : rawValue
  return (
    <div className={cx('rh-statistic', 'ant-statistic', className)} {...props}>
      {title && <div className="rh-statistic__title ant-statistic-title">{title}</div>}
      <div className="rh-statistic__content ant-statistic-content" style={{ ...valueStyle, ...styles?.content }}>
        {prefix && <span className="ant-statistic-content-prefix">{prefix}</span>}
        <span className="ant-statistic-content-value" style={styles?.value}>{displayValue}</span>
        {suffix && <span className="ant-statistic-content-suffix">{suffix}</span>}
      </div>
    </div>
  )
}

interface DividerProps extends HTMLAttributes<HTMLDivElement> {
  orientation?: 'left' | 'right' | 'center'
  plain?: boolean
}

export function Divider({ className, children, ...props }: DividerProps) {
  return (
    <div className={cx('rh-divider', Boolean(children) && 'rh-divider--with-text', 'ant-divider', className)} role="separator" {...props}>
      {children && <span className="rh-divider__text">{children}</span>}
    </div>
  )
}

interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string
  icon?: ReactNode
  size?: number | SizeToken
  children?: ReactNode
}

export function Avatar({ src, icon, size = 36, children, className, ...props }: AvatarProps) {
  const numericSize = typeof size === 'number' ? size : size === 'large' ? 44 : size === 'small' ? 28 : 36
  if (src) {
    return (
      <img
        className={cx('rh-avatar', 'ant-avatar', className)}
        src={src}
        alt=""
        style={{ width: numericSize, height: numericSize }}
      />
    )
  }

  return (
    <div className={cx('rh-avatar', 'ant-avatar', className)} style={{ width: numericSize, height: numericSize }} {...props}>
      {icon ?? children ?? <DesignAvatar name="Relax Hub" size={numericSize} />}
    </div>
  )
}
