/* eslint-disable @typescript-eslint/no-explicit-any, react-refresh/only-export-components */
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import {
  Children,
  cloneElement,
  createContext,
  Fragment,
  forwardRef,
  isValidElement,
  useContext,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from 'react'
import type {
  ButtonHTMLAttributes,
  ChangeEvent,
  CSSProperties,
  FormEvent,
  HTMLAttributes,
  InputHTMLAttributes,
  Key,
  KeyboardEvent as ReactKeyboardEvent,
  MouseEvent as ReactMouseEvent,
  Ref,
  ReactElement,
  ReactNode,
  TextareaHTMLAttributes,
} from 'react'

type FieldName = string | number | readonly (string | number)[]
type FieldValues = Record<string, any>
type SizeToken = 'small' | 'middle' | 'large'

export interface ThemeConfig {
  token?: Record<string, unknown>
  components?: Record<string, unknown>
}

export interface FormRule {
  required?: boolean
  message?: ReactNode
  type?: 'email' | 'url' | string
  whitespace?: boolean
  min?: number
  max?: number
  validator?: (_rule: FormRule, value: any, callback?: (error?: string) => void) => Promise<void> | void
}

type FormRuleFactory = (form: FormInstance<any>) => FormRule

export interface FormInstance<T = FieldValues> {
  setFieldsValue: (values: Partial<T>) => void
  setFieldValue: (name: FieldName, value: any) => void
  getFieldsValue: () => any
  getFieldValue: (name: FieldName) => any
  resetFields: () => void
  validateFields: (names?: FieldName[]) => Promise<any>
  submit: () => void
}

interface InternalFormInstance<T = FieldValues> extends FormInstance<T> {
  __setCallbacks: (callbacks: FormCallbacks<T>) => void
  __setInitialValues: (values?: Partial<T>) => void
  __registerField: (name: FieldName, config: FieldConfig) => () => void
  __subscribe: (listener: () => void) => () => void
  __getError: (name: FieldName) => ReactNode | undefined
}

interface FieldConfig {
  rules?: Array<FormRule | FormRuleFactory>
  initialValue?: any
}

interface FormCallbacks<T = FieldValues> {
  onFinish?: (values: T) => void
  onFinishFailed?: (info: { values: T; errorFields: Array<{ name: FieldName; errors: ReactNode[] }> }) => void
  onValuesChange?: (changedValues: Partial<T>, values: T) => void
}

export interface UploadFile {
  uid?: string
  name?: string
  status?: string
  url?: string
  originFileObj?: File
  [key: string]: any
}

export interface UploadProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  name?: string
  accept?: string
  multiple?: boolean
  disabled?: boolean
  showUploadList?: boolean
  fileList?: UploadFile[]
  beforeUpload?: (file: File, fileList: File[]) => boolean | symbol | Promise<boolean | symbol>
  onChange?: (info: { file: UploadFile; fileList: UploadFile[] }) => void
  children?: ReactNode
}

export interface TextAreaRef {
  resizableTextArea?: {
    textArea: HTMLTextAreaElement
  }
  focus?: () => void
  blur?: () => void
}

export interface MenuItem {
  key?: Key
  label?: ReactNode
  icon?: ReactNode
  danger?: boolean
  disabled?: boolean
  type?: 'divider' | 'group'
  children?: MenuItem[]
  onClick?: () => void
}

export interface MenuProps {
  items?: MenuItem[]
  onClick?: (info: { key: string; item: MenuItem; domEvent: ReactMouseEvent<HTMLElement> }) => void
  selectedKeys?: Key[]
  selectable?: boolean
}

export interface ColumnType<T> {
  key?: Key
  title?: ReactNode
  dataIndex?: FieldName
  render?: (value: any, record: T, index: number) => ReactNode
  width?: number | string
  align?: 'left' | 'center' | 'right'
  className?: string
  sorter?: boolean | ((a: T, b: T) => number)
  fixed?: 'left' | 'right' | boolean
  children?: ColumnsType<T>
  responsive?: string[]
  ellipsis?: boolean | Record<string, unknown>
  [key: string]: any
}

export type ColumnsType<T = any> = Array<ColumnType<T>>

function cx(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

function namePath(name: FieldName): Array<string | number> {
  return Array.isArray(name) ? [...name] : [name]
}

function nameKey(name: FieldName): string {
  return namePath(name).join('.')
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value) && !(value instanceof Date) && !dayjs.isDayjs(value)
}

function cloneValue<T>(value: T): T {
  if (Array.isArray(value)) return value.map((item) => cloneValue(item)) as T
  if (isPlainObject(value)) {
    return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, cloneValue(item)])) as T
  }
  return value
}

function mergeValues(base: FieldValues, next: FieldValues): FieldValues {
  const result = { ...base }
  Object.entries(next).forEach(([key, value]) => {
    if (isPlainObject(value) && isPlainObject(result[key])) {
      result[key] = mergeValues(result[key] as FieldValues, value as FieldValues)
    } else {
      result[key] = cloneValue(value)
    }
  })
  return result
}

function getValueAt(values: FieldValues, name: FieldName): any {
  return namePath(name).reduce<any>((current, part) => (current == null ? undefined : current[part]), values)
}

function setValueAt(values: FieldValues, name: FieldName, value: any): FieldValues {
  const path = namePath(name)
  if (path.length === 0) return values
  const result = Array.isArray(values) ? [...values] : { ...values }
  let cursor: any = result
  path.forEach((part, index) => {
    if (index === path.length - 1) {
      cursor[part] = value
      return
    }
    const next = cursor[part]
    cursor[part] = Array.isArray(next) ? [...next] : isPlainObject(next) ? { ...next } : {}
    cursor = cursor[part]
  })
  return result
}

function isEmptyValue(value: any) {
  if (value == null) return true
  if (typeof value === 'string') return value.length === 0
  if (Array.isArray(value)) return value.length === 0
  return value === false
}

function textFromNode(node: ReactNode): string {
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(textFromNode).join(' ')
  return ''
}

function breakpointMatches(name: string) {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  const queries: Record<string, string> = {
    xs: '(max-width: 575px)',
    sm: '(min-width: 576px)',
    md: '(min-width: 768px)',
    lg: '(min-width: 992px)',
    xl: '(min-width: 1200px)',
    xxl: '(min-width: 1600px)',
  }
  const query = queries[name]
  return query ? window.matchMedia(query).matches : false
}

function responsiveVisible(responsive?: string[]) {
  if (!responsive || responsive.length === 0) return true
  return responsive.some((name) => breakpointMatches(name))
}

function createFormStore<T extends FieldValues = FieldValues>(): InternalFormInstance<T> {
  let values: FieldValues = {}
  let initialValues: FieldValues = {}
  let callbacks: FormCallbacks<T> = {}
  const fields = new Map<string, { name: FieldName; config: FieldConfig }>()
  const errors = new Map<string, ReactNode>()
  const listeners = new Set<() => void>()

  const notify = () => listeners.forEach((listener) => listener())

  const validateField = async (name: FieldName, rules: Array<FormRule | FormRuleFactory> = []) => {
    const value = getValueAt(values, name)
    for (const rawRule of rules) {
      const rule = typeof rawRule === 'function' ? rawRule(store) : rawRule
      if (rule.required && isEmptyValue(value)) {
        return rule.message ?? 'Заполните поле'
      }
      if (rule.whitespace && typeof value === 'string' && value.trim().length === 0) {
        return rule.message ?? 'Заполните поле'
      }
      if (rule.type === 'email' && value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(value))) {
        return rule.message ?? 'Введите корректный email'
      }
      if (rule.type === 'url' && value) {
        try {
          new URL(String(value))
        } catch {
          return rule.message ?? 'Введите корректный URL'
        }
      }
      if (typeof rule.min === 'number' && rule.type === 'number' && value != null && Number(value) < rule.min) {
        return rule.message ?? `Минимум ${rule.min}`
      }
      if (typeof rule.min === 'number' && rule.type !== 'number' && value != null && String(value).length < rule.min) {
        return rule.message ?? `Минимум ${rule.min}`
      }
      if (typeof rule.max === 'number' && rule.type === 'number' && value != null && Number(value) > rule.max) {
        return rule.message ?? `Максимум ${rule.max}`
      }
      if (typeof rule.max === 'number' && rule.type !== 'number' && value != null && String(value).length > rule.max) {
        return rule.message ?? `Максимум ${rule.max}`
      }
      if (rule.validator) {
        try {
          await rule.validator(rule, value)
        } catch (error) {
          return error instanceof Error ? error.message : rule.message ?? String(error)
        }
      }
    }
    return undefined
  }

  const store: InternalFormInstance<T> = {
    setFieldsValue(next) {
      values = mergeValues(values, next as FieldValues)
      callbacks.onValuesChange?.(next, cloneValue(values) as T)
      notify()
    },
    setFieldValue(name, value) {
      values = setValueAt(values, name, value)
      callbacks.onValuesChange?.(setValueAt({}, name, value) as Partial<T>, cloneValue(values) as T)
      notify()
    },
    getFieldsValue() {
      return cloneValue(values) as T
    },
    getFieldValue(name) {
      return getValueAt(values, name)
    },
    resetFields() {
      values = cloneValue(initialValues)
      errors.clear()
      notify()
    },
    async validateFields(names) {
      const entries = [...fields.values()].filter((field) => !names || names.some((name) => nameKey(name) === nameKey(field.name)))
      const errorFields: Array<{ name: FieldName; errors: ReactNode[] }> = []
      for (const field of entries) {
        const error = await validateField(field.name, field.config.rules)
        if (error) {
          errors.set(nameKey(field.name), error)
          errorFields.push({ name: field.name, errors: [error] })
        } else {
          errors.delete(nameKey(field.name))
        }
      }
      notify()
      if (errorFields.length > 0) {
        throw { values: cloneValue(values) as T, errorFields }
      }
      return cloneValue(values) as T
    },
    submit() {
      void store.validateFields()
        .then((validValues) => callbacks.onFinish?.(validValues))
        .catch((info: { values: T; errorFields: Array<{ name: FieldName; errors: ReactNode[] }> }) => callbacks.onFinishFailed?.(info))
    },
    __setCallbacks(nextCallbacks) {
      callbacks = nextCallbacks
    },
    __setInitialValues(nextInitialValues) {
      initialValues = cloneValue((nextInitialValues ?? {}) as FieldValues)
      values = mergeValues(cloneValue(initialValues), values)
      notify()
    },
    __registerField(name, config) {
      const key = nameKey(name)
      fields.set(key, { name, config })
      if (config.initialValue !== undefined && getValueAt(values, name) === undefined) {
        values = setValueAt(values, name, config.initialValue)
      }
      notify()
      return () => {
        fields.delete(key)
        errors.delete(key)
      }
    },
    __subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    __getError(name) {
      return errors.get(nameKey(name))
    },
  }

  return store
}

const FormContext = createContext<InternalFormInstance | null>(null)

interface FormProps<T extends FieldValues = FieldValues> extends Omit<HTMLAttributes<HTMLFormElement>, 'onSubmit' | 'onChange'> {
  form?: FormInstance<any>
  layout?: 'horizontal' | 'vertical' | 'inline'
  initialValues?: Partial<T>
  onFinish?: (values: T) => void
  onFinishFailed?: (info: { values: T; errorFields: Array<{ name: FieldName; errors: ReactNode[] }> }) => void
  onValuesChange?: (changedValues: Partial<T>, values: T) => void
  [key: string]: any
}

interface FormItemProps extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  name?: FieldName
  label?: ReactNode
  rules?: Array<FormRule | FormRuleFactory>
  valuePropName?: string
  initialValue?: any
  noStyle?: boolean
  extra?: ReactNode
  help?: ReactNode
  normalize?: (value: any, previousValue: any, allValues: FieldValues) => any
  getValueFromEvent?: (...args: any[]) => any
  dependencies?: FieldName[]
  shouldUpdate?: boolean | ((previousValues: FieldValues, nextValues: FieldValues) => boolean)
  children?: ReactNode | ((form: FormInstance) => ReactNode)
  tooltip?: ReactNode
  [key: string]: any
}

function useForm<T extends FieldValues = FieldValues>(form?: FormInstance<T>): [FormInstance<T>] {
  const [internalForm] = useState<InternalFormInstance<T>>(() => (form as InternalFormInstance<T>) ?? createFormStore<T>())
  return [internalForm]
}

function useWatch(name: FieldName, form?: FormInstance<any>) {
  const contextForm = useContext(FormContext)
  const targetForm = (form as InternalFormInstance | undefined) ?? contextForm
  const [value, setValue] = useState(() => targetForm?.getFieldValue(name))

  useEffect(() => {
    if (!targetForm) return undefined
    const sync = () => setValue(targetForm.getFieldValue(name))
    sync()
    return targetForm.__subscribe(sync)
  }, [name, targetForm])

  return value
}

function FormRoot<T extends FieldValues = FieldValues>({
  form,
  layout = 'horizontal',
  initialValues,
  onFinish,
  onFinishFailed,
  onValuesChange,
  className,
  children,
  ...props
}: FormProps<T>) {
  const [internalForm] = useForm<T>(form)

  useEffect(() => {
    ;(internalForm as InternalFormInstance<T>).__setCallbacks({ onFinish, onFinishFailed, onValuesChange })
  }, [internalForm, onFinish, onFinishFailed, onValuesChange])

  useEffect(() => {
    ;(internalForm as InternalFormInstance<T>).__setInitialValues(initialValues)
  }, [initialValues, internalForm])

  return (
    <FormContext.Provider value={internalForm as InternalFormInstance}>
      <form
        className={cx('rh-form', 'ant-form', layout === 'vertical' && 'ant-form-vertical', layout === 'inline' && 'ant-form-inline', className)}
        onSubmit={(event: FormEvent<HTMLFormElement>) => {
          event.preventDefault()
          internalForm.submit()
        }}
        {...props}
      >
        {children}
      </form>
    </FormContext.Provider>
  )
}

function extractValue(eventOrValue: any, valuePropName: string) {
  if (eventOrValue?.target) {
    return valuePropName === 'checked' ? eventOrValue.target.checked : eventOrValue.target.value
  }
  return eventOrValue
}

function FormItem({
  name,
  label,
  rules,
  valuePropName = 'value',
  initialValue,
  noStyle,
  extra,
  help,
  normalize,
  getValueFromEvent,
  children,
  className,
  ...props
}: FormItemProps) {
  const form = useContext(FormContext)
  const [, forceRender] = useState(0)

  useEffect(() => {
    if (!form || name === undefined) return undefined
    return form.__registerField(name, { rules, initialValue })
  }, [form, initialValue, name, rules])

  useEffect(() => {
    if (!form) return undefined
    return form.__subscribe(() => forceRender((value) => value + 1))
  }, [form])

  const error = form && name !== undefined ? form.__getError(name) : undefined
  const currentValue = form && name !== undefined ? form.getFieldValue(name) : undefined
  const childContent = typeof children === 'function' ? children(form ?? createFormStore()) : children
  const fieldId = name !== undefined ? `rh-form-${nameKey(name).replace(/[^a-zA-Z0-9_-]/g, '-')}` : undefined

  let control = childContent
  const childArray = Children.toArray(childContent)
  if (form && name !== undefined && childArray.length === 1 && isValidElement(childArray[0])) {
    const child = childArray[0] as ReactElement<Record<string, any>>
    const previousOnChange = child.props.onChange
    const controlValue = valuePropName === 'checked' ? Boolean(currentValue) : currentValue ?? ''
    control = cloneElement(child, {
      id: child.props.id ?? fieldId,
      [valuePropName]: controlValue,
      onChange: (...args: any[]) => {
        const rawValue = getValueFromEvent ? getValueFromEvent(...args) : extractValue(args[0], valuePropName)
        const nextValue = normalize ? normalize(rawValue, currentValue, form.getFieldsValue()) : rawValue
        form.setFieldValue(name, nextValue)
        previousOnChange?.(...args)
      },
    })
  }

  if (noStyle) return <>{control}</>

  return (
    <div className={cx('rh-form-item', 'ant-form-item', Boolean(error) && 'ant-form-item-has-error', className)} {...props}>
      {label && (
        <div className="rh-form-item__label ant-form-item-label">
          <label htmlFor={fieldId}>{label}</label>
        </div>
      )}
      <div className="rh-form-item__control ant-form-item-control">
        <div className="ant-form-item-control-input">
          <div className="ant-form-item-control-input-content">{control}</div>
        </div>
        {(help || error) && <div className="ant-form-item-explain-error">{help ?? error}</div>}
        {extra && <div className="rh-form-item__extra ant-form-item-extra">{extra}</div>}
      </div>
    </div>
  )
}

export const Form = Object.assign(FormRoot, {
  Item: FormItem,
  useForm,
  useWatch,
})

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size' | 'prefix'> {
  prefix?: ReactNode
  suffix?: ReactNode
  addonAfter?: ReactNode
  addonBefore?: ReactNode
  allowClear?: boolean
  size?: SizeToken
  status?: string
  showCount?: boolean
  onPressEnter?: (event: ReactKeyboardEvent<HTMLInputElement>) => void
  onClear?: () => void
}

function InputInner({
  prefix,
  suffix,
  addonAfter,
  addonBefore,
  allowClear,
  size: _size,
  status: _status,
  showCount: _showCount,
  onPressEnter,
  onKeyDown,
  onClear,
  className,
  value,
  onChange,
  ...props
}: InputProps, ref: Ref<HTMLInputElement>) {
  const input = (
    <input
      ref={ref}
      className={cx(
        'rh-input',
        'ant-input',
        'min-h-[50px] w-full rounded-rh-md border border-[rgba(15,23,42,0.14)] bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(251,247,240,0.96))] px-3.5 py-3 font-sans text-[15px] font-medium text-rh-text shadow-rh-control outline-none transition placeholder:text-[rgba(95,104,119,0.72)] hover:border-[rgba(15,118,110,0.28)] focus:border-[rgba(15,118,110,0.52)] focus:shadow-rh-control-focus disabled:cursor-not-allowed disabled:bg-[rgba(244,239,231,0.82)] disabled:text-[rgba(22,33,43,0.42)]',
        className,
      )}
      value={value}
      onChange={onChange}
      onKeyDown={(event) => {
        if (event.key === 'Enter') onPressEnter?.(event)
        onKeyDown?.(event)
      }}
      {...props}
    />
  )
  const clear = allowClear && value ? (
    <button
      type="button"
      className="rh-input-clear"
      aria-label="Очистить"
      onClick={() => {
        onChange?.({ target: { value: '' } } as ChangeEvent<HTMLInputElement>)
        onClear?.()
      }}
    >
      ×
    </button>
  ) : null

  const wrapped = prefix || suffix || clear ? (
    <span className="rh-input-affix ant-input-affix-wrapper flex min-h-[50px] w-full items-center gap-2 rounded-rh-md border border-[rgba(15,23,42,0.14)] bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(251,247,240,0.96))] px-3.5 py-0 shadow-rh-control transition focus-within:border-[rgba(15,118,110,0.52)] focus-within:shadow-rh-control-focus">
      {prefix && <span className="ant-input-prefix">{prefix}</span>}
      {input}
      {clear}
      {suffix && <span className="ant-input-suffix">{suffix}</span>}
    </span>
  ) : input

  if (addonBefore || addonAfter) {
    return (
      <span className="rh-compact-control ant-space-compact inline-flex w-full items-stretch">
        {addonBefore && <span className="rh-input-addon">{addonBefore}</span>}
        {wrapped}
        {addonAfter && <span className="rh-input-addon">{addonAfter}</span>}
      </span>
    )
  }

  return wrapped
}

const BaseInput = forwardRef<HTMLInputElement, InputProps>(InputInner)

interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  showCount?: boolean
  size?: SizeToken
  onPressEnter?: (event: ReactKeyboardEvent<HTMLTextAreaElement>) => void
}

const TextArea = forwardRef<TextAreaRef, TextAreaProps>(function TextArea({ className, showCount: _showCount, size: _size, onPressEnter, onKeyDown, ...props }, ref) {
  const textAreaRef = useRef<HTMLTextAreaElement>(null)
  useImperativeHandle(ref, () => ({
    resizableTextArea: textAreaRef.current ? { textArea: textAreaRef.current } : undefined,
    focus: () => textAreaRef.current?.focus(),
    blur: () => textAreaRef.current?.blur(),
  }))
  return (
    <textarea
      ref={textAreaRef}
      className={cx(
        'rh-input',
        'ant-input',
        'min-h-[96px] w-full rounded-rh-md border border-[rgba(15,23,42,0.14)] bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(251,247,240,0.96))] px-3.5 py-3 font-sans text-[15px] font-medium text-rh-text shadow-rh-control outline-none transition placeholder:text-[rgba(95,104,119,0.72)] hover:border-[rgba(15,118,110,0.28)] focus:border-[rgba(15,118,110,0.52)] focus:shadow-rh-control-focus disabled:cursor-not-allowed disabled:bg-[rgba(244,239,231,0.82)] disabled:text-[rgba(22,33,43,0.42)]',
        className,
      )}
      onKeyDown={(event) => {
        if (event.key === 'Enter') onPressEnter?.(event)
        onKeyDown?.(event)
      }}
      {...props}
    />
  )
})

function Password(props: InputProps) {
  return <BaseInput {...props} type="password" suffix={props.suffix ?? <span className="ant-input-password-icon">••</span>} />
}

interface SearchProps extends InputProps {
  enterButton?: ReactNode
  loading?: boolean
  onSearch?: (value: string) => void
}

function Search({ enterButton, loading, onSearch, onKeyDown, className, style, ...props }: SearchProps) {
  const [value, setValue] = useState(String(props.value ?? props.defaultValue ?? ''))
  const searchValue = props.value !== undefined ? String(props.value) : value
  return (
    <span className={cx('rh-input-search ant-input-search', className)} style={style}>
      <BaseInput
        {...props}
        value={searchValue}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          if (props.value === undefined) setValue(event.target.value)
          props.onChange?.(event)
        }}
        onKeyDown={(event: ReactKeyboardEvent<HTMLInputElement>) => {
          if (event.key === 'Enter') onSearch?.(searchValue)
          onKeyDown?.(event)
        }}
      />
      <button className="rh-input-search__button ant-btn" type="button" disabled={loading} onClick={() => onSearch?.(searchValue)}>
        {enterButton || 'Найти'}
      </button>
    </span>
  )
}

export const Input = Object.assign(BaseInput, { TextArea, Password, Search })

interface InputNumberProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'defaultValue' | 'onChange' | 'size' | 'prefix'> {
  value?: number | null
  defaultValue?: number
  min?: number
  max?: number
  step?: number
  size?: SizeToken
  prefix?: ReactNode
  addonAfter?: ReactNode
  precision?: number
  onChange?: (value: number | null) => void
  [key: string]: any
}

export function InputNumber({ value, defaultValue, onChange, prefix, addonAfter, precision: _precision, size: _size, className, ...props }: InputNumberProps) {
  const displayValue = value ?? defaultValue ?? ''
  return (
    <span className={cx('rh-input-number ant-input-number', className)}>
      {prefix && <span className="ant-input-number-prefix">{prefix}</span>}
      <span className="ant-input-number-input-wrap">
        <input
          className="ant-input-number-input"
          type="number"
          value={displayValue}
          onChange={(event) => {
            const next = event.target.value === '' ? null : Number(event.target.value)
            onChange?.(Number.isNaN(next) ? null : next)
          }}
          {...props}
        />
      </span>
      {addonAfter && <span className="rh-input-addon">{addonAfter}</span>}
    </span>
  )
}

interface SelectOption {
  label?: ReactNode
  value: any
  disabled?: boolean
}

interface SelectProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value?: any
  defaultValue?: any
  options?: SelectOption[]
  placeholder?: ReactNode
  disabled?: boolean
  allowClear?: boolean
  mode?: 'multiple' | 'tags'
  showSearch?: boolean
  dropdownClassName?: string
  popupClassName?: string
  classNames?: Record<string, any>
  loading?: boolean
  size?: SizeToken | string
  variant?: string
  suffixIcon?: ReactNode
  popupMatchSelectWidth?: boolean
  style?: CSSProperties
  onChange?: (value: any, option?: SelectOption | SelectOption[]) => void
  children?: ReactNode
  [key: string]: any
}

interface OptionProps {
  value: any
  children?: ReactNode
  disabled?: boolean
}

function Option(_props: OptionProps) {
  return null
}

function optionsFromChildren(children: ReactNode): SelectOption[] {
  return Children.toArray(children)
    .filter(isValidElement)
    .map((child) => {
      const element = child as ReactElement<OptionProps>
      return { value: element.props.value, label: element.props.children, disabled: element.props.disabled }
    })
}

function SelectRoot({
  value,
  defaultValue,
  options,
  placeholder,
  disabled,
  allowClear,
  mode,
  className,
  classNames: _classNames,
  dropdownClassName: _dropdownClassName,
  popupClassName: _popupClassName,
  popupMatchSelectWidth: _popupMatchSelectWidth,
  loading: _loading,
  size: _size,
  variant: _variant,
  suffixIcon,
  showSearch: _showSearch,
  id,
  style,
  onChange,
  children,
  ...props
}: SelectProps) {
  const rootRef = useRef<HTMLSpanElement>(null)
  const [open, setOpen] = useState(false)
  const [internalValue, setInternalValue] = useState(defaultValue ?? (mode ? [] : ''))
  const resolvedOptions = options ?? optionsFromChildren(children)
  const optionByDomValue = new Map(resolvedOptions.map((option) => [String(option.value), option]))
  const selectedValue = value !== undefined ? value : internalValue
  const selectedOptions = mode && Array.isArray(selectedValue)
    ? selectedValue.map((item) => resolvedOptions.find((option) => option.value === item)).filter(Boolean) as SelectOption[]
    : []
  const selectedOption = !mode ? optionByDomValue.get(String(selectedValue ?? '')) : undefined
  const hasSelection = mode ? selectedOptions.length > 0 : selectedValue !== '' && selectedValue != null
  const displayLabel = mode
    ? selectedOptions.map((option) => textFromNode(option.label ?? option.value)).join(', ')
    : textFromNode(selectedOption?.label ?? (hasSelection ? selectedValue : placeholder) ?? '')
  const hasEmptyOption = resolvedOptions.some((option) => String(option.value) === '')
  const dropdownOptions = allowClear && !mode && !hasEmptyOption
    ? [{ value: '', label: placeholder ?? 'Очистить' }, ...resolvedOptions]
    : resolvedOptions
  const popupRootClassName = typeof _classNames?.popup?.root === 'string' ? _classNames.popup.root : undefined
  const dropdownId = id ? `${id}-dropdown` : undefined

  useEffect(() => {
    if (!open) return

    const handlePointerDown = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false)
    }

    document.addEventListener('pointerdown', handlePointerDown)
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('pointerdown', handlePointerDown)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [open])

  const commitValue = (nextValue: any, option: SelectOption) => {
    if (mode) {
      const current = Array.isArray(selectedValue) ? selectedValue : []
      const exists = current.some((item) => String(item) === String(nextValue))
      const next = exists
        ? current.filter((item) => String(item) !== String(nextValue))
        : [...current, nextValue]
      if (value === undefined) setInternalValue(next)
      onChange?.(next, next.map((item) => resolvedOptions.find((candidate) => candidate.value === item) ?? { value: item }))
      return
    }

    if (value === undefined) setInternalValue(nextValue)
    onChange?.(nextValue, option)
    setOpen(false)
  }

  return (
    <span
      ref={rootRef}
      className={cx(
        'rh-select',
        'ant-select',
        'relative inline-flex w-full min-w-0 items-stretch',
        disabled && 'ant-select-disabled cursor-not-allowed opacity-60',
        mode ? 'ant-select-multiple' : 'ant-select-single',
        className,
      )}
      style={style}
      {...props}
    >
      <button
        type="button"
        id={id}
        className="rh-select__trigger ant-select-selector"
        disabled={disabled}
        role="combobox"
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-controls={dropdownId}
        onClick={() => {
          if (!disabled) setOpen((current) => !current)
        }}
        onKeyDown={(event) => {
          if (event.key === 'ArrowDown' || event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            if (!disabled) setOpen(true)
          }
        }}
      >
        <span className={cx('rh-select__value min-w-0 flex-1 truncate font-sans text-[15px] font-medium', !hasSelection && 'ant-select-selection-placeholder text-[rgba(95,104,119,0.72)]', hasSelection && 'ant-select-selection-item text-rh-text')}>
          {displayLabel}
        </span>
      </button>
      {open && !disabled && (
        <span
          id={dropdownId}
          className={cx('rh-select__dropdown ant-select-dropdown', _dropdownClassName, _popupClassName, popupRootClassName)}
          role="listbox"
          aria-multiselectable={Boolean(mode)}
        >
          {dropdownOptions.map((option) => {
            const selected = mode && Array.isArray(selectedValue)
              ? selectedValue.some((item) => String(item) === String(option.value))
              : String(selectedValue ?? '') === String(option.value)
            return (
              <button
                key={String(option.value)}
                type="button"
                className={cx(
                  'rh-select__option ant-select-item ant-select-item-option',
                  selected && 'rh-select__option--selected ant-select-item-option-selected',
                  option.disabled && 'ant-select-item-option-disabled',
                )}
                role="option"
                aria-selected={selected}
                disabled={option.disabled}
                onClick={() => commitValue(option.value, option)}
              >
                <span className="ant-select-item-option-state" aria-hidden="true">
                  {selected ? '✓' : ''}
                </span>
                <span className="ant-select-item-option-content">
                  {option.label ?? option.value}
                </span>
              </button>
            )
          })}
        </span>
      )}
      <span className="ant-select-arrow" aria-hidden="true">{suffixIcon ?? '⌄'}</span>
    </span>
  )
}

export const Select = Object.assign(SelectRoot, { Option })

interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'type'> {
  checked?: boolean
  defaultChecked?: boolean
  indeterminate?: boolean
  onChange?: (event: ChangeEvent<HTMLInputElement>) => void
  children?: ReactNode
}

function CheckboxRoot({ checked, defaultChecked, indeterminate, onChange, className, children, ...props }: CheckboxProps) {
  return (
    <label className={cx('rh-checkbox ant-checkbox-wrapper', checked && 'ant-checkbox-wrapper-checked', className)}>
      <span className={cx('ant-checkbox', checked && 'ant-checkbox-checked', indeterminate && 'ant-checkbox-indeterminate')}>
        <input className="ant-checkbox-input" type="checkbox" checked={checked} defaultChecked={defaultChecked} onChange={onChange} {...props} />
        <span className="ant-checkbox-inner" />
      </span>
      {children && <span>{children}</span>}
    </label>
  )
}

interface CheckboxGroupProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value?: any[]
  defaultValue?: any[]
  options?: SelectOption[]
  onChange?: (checkedValue: any[]) => void
}

function CheckboxGroup({ value, defaultValue = [], options, onChange, className, children, ...props }: CheckboxGroupProps) {
  const [internal, setInternal] = useState<any[]>(defaultValue)
  const selected = value ?? internal
  const setSelected = (next: any[]) => {
    setInternal(next)
    onChange?.(next)
  }
  const rendered = options
    ? options.map((option) => (
      <CheckboxRoot
        key={String(option.value)}
        checked={selected.includes(option.value)}
        disabled={option.disabled}
        onChange={(event) => {
          setSelected(event.target.checked ? [...selected, option.value] : selected.filter((item) => item !== option.value))
        }}
      >
        {option.label}
      </CheckboxRoot>
    ))
    : children
  return <div className={cx('rh-checkbox-group ant-checkbox-group', className)} {...props}>{rendered}</div>
}

export const Checkbox = Object.assign(CheckboxRoot, { Group: CheckboxGroup })

interface RadioContextValue {
  value: any
  onChange?: (value: any) => void
  name?: string
}

const RadioContext = createContext<RadioContextValue | null>(null)

interface RadioProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'type'> {
  value?: any
  checked?: boolean
  onChange?: (event: ChangeEvent<HTMLInputElement>) => void
  children?: ReactNode
}

function RadioRoot({ value, checked, onChange, className, children, ...props }: RadioProps) {
  const group = useContext(RadioContext)
  const resolvedChecked = checked ?? (group ? group.value === value : undefined)
  return (
    <label className={cx('rh-radio ant-radio-wrapper', resolvedChecked && 'ant-radio-wrapper-checked', className)}>
      <span className={cx('ant-radio', resolvedChecked && 'ant-radio-checked')}>
        <input
          className="ant-radio-input"
          type="radio"
          name={group?.name}
          value={String(value ?? '')}
          checked={resolvedChecked}
          onChange={(event) => {
            group?.onChange?.(value)
            onChange?.(event)
          }}
          {...props}
        />
        <span className="ant-radio-inner" />
      </span>
      {children && <span>{children}</span>}
    </label>
  )
}

interface RadioGroupProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value?: any
  defaultValue?: any
  options?: SelectOption[]
  name?: string
  buttonStyle?: string
  optionType?: string
  onChange?: (event: { target: { value: any } }) => void
}

function RadioGroup({ value, defaultValue, options, name, onChange, className, children, ...props }: RadioGroupProps) {
  const [internalValue, setInternalValue] = useState(defaultValue)
  const selected = value ?? internalValue
  const changeValue = (next: any) => {
    setInternalValue(next)
    onChange?.({ target: { value: next } })
  }
  return (
    <RadioContext.Provider value={{ value: selected, onChange: changeValue, name }}>
      <div className={cx('rh-radio-group ant-radio-group', className)} {...props}>
        {options ? options.map((option) => <RadioRoot key={String(option.value)} value={option.value} disabled={option.disabled}>{option.label}</RadioRoot>) : children}
      </div>
    </RadioContext.Provider>
  )
}

function RadioButton(props: RadioProps) {
  return <RadioRoot {...props} className={cx('rh-radio-button ant-radio-button-wrapper', props.className)} />
}

export const Radio = Object.assign(RadioRoot, { Group: RadioGroup, Button: RadioButton })

interface SwitchProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onChange'> {
  checked?: boolean
  defaultChecked?: boolean
  onChange?: (checked: boolean) => void
  checkedChildren?: ReactNode
  unCheckedChildren?: ReactNode
  loading?: boolean
  size?: SizeToken | string
}

export function Switch({ checked, defaultChecked = false, onChange, checkedChildren, unCheckedChildren, loading, size: _size, className, ...props }: SwitchProps) {
  const [internal, setInternal] = useState(defaultChecked)
  const active = checked ?? internal
  return (
    <button
      type="button"
      role="switch"
      aria-checked={active}
      className={cx('rh-switch ant-switch', active && 'ant-switch-checked', loading && 'ant-switch-loading', className)}
      onClick={() => {
        setInternal(!active)
        onChange?.(!active)
      }}
      {...props}
    >
      <span className="ant-switch-handle" />
      <span className="ant-switch-inner">{active ? checkedChildren : unCheckedChildren}</span>
    </button>
  )
}

interface SegmentedOption {
  label: ReactNode
  value: any
  icon?: ReactNode
  disabled?: boolean
}

interface SegmentedProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  options: Array<SegmentedOption | string | number>
  value?: any
  defaultValue?: any
  block?: boolean
  onChange?: (value: any) => void
}

export function Segmented({ options, value, defaultValue, block, onChange, className, ...props }: SegmentedProps) {
  const normalized = options.map((option) => (typeof option === 'object' ? option : { label: option, value: option }))
  const [internal, setInternal] = useState(defaultValue ?? normalized[0]?.value)
  const selected = value ?? internal
  return (
    <div className={cx('rh-segmented ant-segmented', block && 'ant-segmented-block', className)} {...props}>
      {normalized.map((option) => (
        <button
          key={String(option.value)}
          type="button"
          disabled={option.disabled}
          className={cx('ant-segmented-item', selected === option.value && 'ant-segmented-item-selected')}
          onClick={() => {
            setInternal(option.value)
            onChange?.(option.value)
          }}
        >
          <span className="ant-segmented-item-label">{option.icon}{option.label}</span>
        </button>
      ))}
    </div>
  )
}

interface SliderProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'value' | 'defaultValue' | 'type'> {
  value?: number | number[]
  defaultValue?: number | number[]
  min?: number
  max?: number
  step?: number
  marks?: Record<number, ReactNode>
  range?: boolean
  tooltip?: { formatter?: (value?: number) => ReactNode }
  onChange?: (value: any) => void
}

export function Slider({ value, defaultValue = 0, min = 0, max = 100, step = 1, marks, range: _range, tooltip: _tooltip, onChange, className, ...props }: SliderProps) {
  const [internal, setInternal] = useState(defaultValue)
  const current = value ?? internal
  const displayValue = Array.isArray(current) ? current[0] ?? min : current
  return (
    <div className={cx('rh-slider ant-slider', className)}>
      <input
        type="range"
        value={displayValue}
        min={min}
        max={max}
        step={step}
        onChange={(event) => {
          const next = Number(event.target.value)
          setInternal(next)
          onChange?.(Array.isArray(current) ? [next, current[1] ?? next] : next)
        }}
        {...props}
      />
      {marks && <div className="rh-slider__marks">{Object.entries(marks).map(([mark, label]) => <span key={mark}>{label}</span>)}</div>}
    </div>
  )
}

interface PickerProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'defaultValue' | 'onChange' | 'size'> {
  value?: any
  defaultValue?: any
  format?: string
  showTime?: boolean | Record<string, any>
  minuteStep?: number
  size?: SizeToken
  disabledDate?: (current: Dayjs) => boolean
  onChange?: (value: Dayjs | null, dateString: string) => void
  [key: string]: any
}

function toDateInput(value: any, includeTime = false) {
  if (!value) return ''
  const parsed = dayjs.isDayjs(value) ? value : dayjs(value)
  if (!parsed.isValid()) return ''
  return parsed.format(includeTime ? 'YYYY-MM-DDTHH:mm' : 'YYYY-MM-DD')
}

function toTimeInput(value: any) {
  if (!value) return ''
  const parsed = dayjs.isDayjs(value) ? value : dayjs(value)
  return parsed.isValid() ? parsed.format('HH:mm') : String(value)
}

function DatePickerRoot({ value, defaultValue, showTime, format: _format, minuteStep: _minuteStep, size: _size, disabledDate: _disabledDate, onChange, className, id, style, ...props }: PickerProps) {
  const [internal, setInternal] = useState(defaultValue)
  const current = value ?? internal
  return (
    <span className={cx('rh-picker ant-picker', className)} style={style}>
      <span className="ant-picker-input">
        <input
          id={id}
          {...props}
          type={showTime ? 'datetime-local' : 'date'}
          value={toDateInput(current, Boolean(showTime))}
          onChange={(event) => {
            const parsed = event.target.value ? dayjs(event.target.value) : null
            setInternal(parsed)
            onChange?.(parsed, event.target.value)
          }}
        />
      </span>
      <span className="ant-picker-suffix">⌄</span>
    </span>
  )
}

interface RangePickerProps extends Omit<HTMLAttributes<HTMLSpanElement>, 'onChange' | 'defaultValue'> {
  value?: [any, any] | null
  defaultValue?: [any, any] | null
  format?: string
  placeholder?: [string, string]
  disabledDate?: (current: Dayjs) => boolean
  onChange?: (value: [any, any] | null, dateStrings: [string, string]) => void
  [key: string]: any
}

function RangePicker({ value, defaultValue = null, placeholder, format: _format, disabledDate: _disabledDate, onChange, className, id, style, ...props }: RangePickerProps) {
  const [internal, setInternal] = useState<[any, any] | null>(defaultValue)
  const current = value ?? internal
  const update = (index: 0 | 1, next: string) => {
    const raw: [any, any] = current ? [...current] as [any, any] : [null, null]
    raw[index] = next ? dayjs(next) : null
    const normalized: [any, any] | null = raw[0] || raw[1] ? raw : null
    setInternal(normalized)
    onChange?.(normalized, [toDateInput(raw[0]), toDateInput(raw[1])])
  }
  return (
    <span className={cx('rh-picker rh-picker-range ant-picker ant-picker-range', className)} style={style} {...props}>
      <span className="ant-picker-input">
        <input id={id} type="date" value={toDateInput(current?.[0])} placeholder={placeholder?.[0]} onChange={(event) => update(0, event.target.value)} />
      </span>
      <span className="ant-picker-separator">–</span>
      <span className="ant-picker-input">
        <input type="date" value={toDateInput(current?.[1])} placeholder={placeholder?.[1]} onChange={(event) => update(1, event.target.value)} />
      </span>
    </span>
  )
}

export const DatePicker = Object.assign(DatePickerRoot, { RangePicker })

export function TimePicker({ value, defaultValue, onChange, minuteStep, format: _format, showTime: _showTime, size: _size, disabledDate: _disabledDate, className, id, style, ...props }: PickerProps) {
  const [internal, setInternal] = useState(defaultValue)
  const current = value ?? internal
  return (
    <span className={cx('rh-picker ant-picker', className)} style={style}>
      <span className="ant-picker-input">
        <input
          id={id}
          {...props}
          type="time"
          step={minuteStep ? minuteStep * 60 : undefined}
          value={toTimeInput(current)}
          onChange={(event) => {
            const parsed = event.target.value ? dayjs(`1970-01-01T${event.target.value}`) : null
            setInternal(parsed)
            onChange?.(parsed, event.target.value)
          }}
        />
      </span>
      <span className="ant-picker-suffix">⌄</span>
    </span>
  )
}

interface PaginationProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  current?: number
  defaultCurrent?: number
  pageSize?: number
  total?: number
  showSizeChanger?: boolean
  pageSizeOptions?: string[]
  showTotal?: (total: number, range?: [number, number]) => ReactNode
  onChange?: (page: number, pageSize: number) => void
  [key: string]: any
}

export function Pagination({
  current,
  defaultCurrent = 1,
  pageSize = 10,
  total = 0,
  showTotal,
  showSizeChanger: _showSizeChanger,
  pageSizeOptions: _pageSizeOptions,
  onChange,
  className,
  ...props
}: PaginationProps) {
  const [internal, setInternal] = useState(defaultCurrent)
  const page = current ?? internal
  const pages = Math.max(1, Math.ceil(total / pageSize))
  const go = (next: number) => {
    const normalized = Math.min(pages, Math.max(1, next))
    setInternal(normalized)
    onChange?.(normalized, pageSize)
  }
  return (
    <div className={cx('rh-pagination ant-pagination', className)} {...props}>
      {showTotal && <span className="rh-pagination__total">{showTotal(total, [(page - 1) * pageSize + 1, Math.min(total, page * pageSize)])}</span>}
      <button type="button" className="ant-pagination-prev" disabled={page <= 1} onClick={() => go(page - 1)}>‹</button>
      {Array.from({ length: pages }, (_, index) => index + 1).slice(0, 7).map((item) => (
        <button
          type="button"
          key={item}
          className={cx('ant-pagination-item', item === page && 'ant-pagination-item-active')}
          onClick={() => go(item)}
        >
          {item}
        </button>
      ))}
      <button type="button" className="ant-pagination-next" disabled={page >= pages} onClick={() => go(page + 1)}>›</button>
    </div>
  )
}

interface TableProps<T> extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  columns?: ColumnsType<T>
  dataSource?: T[]
  rowKey?: keyof T | string | ((record: T) => Key)
  loading?: boolean | { spinning?: boolean }
  locale?: { emptyText?: ReactNode }
  pagination?: false | PaginationProps
  rowClassName?: string | ((record: T, index: number) => string)
  rowSelection?: {
    selectedRowKeys?: Key[]
    onChange?: (selectedRowKeys: Key[], selectedRows: T[]) => void
  }
  expandable?: {
    expandedRowRender?: (record: T, index: number) => ReactNode
    rowExpandable?: (record: T) => boolean
    expandedRowKeys?: Key[]
    defaultExpandedRowKeys?: Key[]
    onExpand?: (expanded: boolean, record: T) => void
  }
  scroll?: { x?: number | string; y?: number | string }
  size?: SizeToken | string
  onRow?: (record: T, index?: number) => HTMLAttributes<HTMLTableRowElement>
  onChange?: (...args: any[]) => void
  [key: string]: any
}

function recordKey<T>(record: T, index: number, rowKey?: TableProps<T>['rowKey']): Key {
  if (typeof rowKey === 'function') return rowKey(record)
  if (typeof rowKey === 'string') return (record as any)[rowKey] ?? index
  return (record as any).key ?? (record as any).id ?? index
}

function cellValue<T>(record: T, dataIndex: FieldName | undefined) {
  if (dataIndex === undefined) return undefined
  return getValueAt(record as FieldValues, dataIndex)
}

export function Table<T extends object>({
  columns = [],
  dataSource = [],
  rowKey,
  loading,
  locale,
  pagination,
  rowClassName,
  rowSelection,
  expandable,
  scroll,
  size: _size,
  onRow,
  className,
  ...props
}: TableProps<T>) {
  const [expanded, setExpanded] = useState<Key[]>(expandable?.defaultExpandedRowKeys ?? [])
  const selectedKeys = rowSelection?.selectedRowKeys ?? []
  const visibleColumns = columns.filter((column) => responsiveVisible(column.responsive))
  const hasPagination = pagination !== false && pagination !== undefined
  const page = hasPagination ? pagination.current ?? pagination.defaultCurrent ?? 1 : 1
  const pageSize = hasPagination ? pagination.pageSize ?? 10 : dataSource.length || 10
  const pagedData = !hasPagination ? dataSource : dataSource.slice((page - 1) * pageSize, page * pageSize)
  const isLoading = typeof loading === 'object' ? loading.spinning : loading
  const tableStyle = scroll?.x ? { minWidth: typeof scroll.x === 'number' ? `${scroll.x}px` : scroll.x } : undefined

  const toggleSelected = (key: Key, record: T, checked: boolean) => {
    const nextKeys = checked ? [...selectedKeys, key] : selectedKeys.filter((item) => item !== key)
    const selectedRows = dataSource.filter((item, index) => nextKeys.includes(recordKey(item, index, rowKey)))
    if (checked && !selectedRows.includes(record)) selectedRows.push(record)
    rowSelection?.onChange?.(nextKeys, selectedRows)
  }

  return (
    <div className={cx('rh-table-wrapper ant-table-wrapper', className)} {...props}>
      <div className="ant-table">
        <div className="ant-table-container">
          <div className="ant-table-content" style={scroll?.x ? { overflowX: 'auto' } : undefined}>
            <table style={tableStyle}>
              <thead className="ant-table-thead">
                <tr>
                  {rowSelection && <th className="ant-table-selection-column" />}
                  {expandable && <th className="ant-table-row-expand-icon-cell" />}
                  {visibleColumns.map((column, index) => (
                    <th key={String(column.key ?? column.dataIndex ?? index)} style={{ width: column.width, textAlign: column.align }}>
                      {column.title}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="ant-table-tbody">
                {pagedData.length === 0 ? (
                  <tr>
                    <td colSpan={visibleColumns.length + (rowSelection ? 1 : 0) + (expandable ? 1 : 0)}>
                      {isLoading && <div className="rh-table__loading">Загрузка...</div>}
                      {locale?.emptyText ?? 'Нет данных'}
                    </td>
                  </tr>
                ) : pagedData.map((record, index) => {
                  const key = recordKey(record, index, rowKey)
                  const expandedKeys = expandable?.expandedRowKeys ?? expanded
                  const canExpand = expandable?.expandedRowRender && (expandable.rowExpandable?.(record) ?? true)
                  const isExpanded = expandedKeys.includes(key)
                  const resolvedRowClassName = typeof rowClassName === 'function' ? rowClassName(record, index) : rowClassName
                  const rowProps = onRow?.(record, index) ?? {}
                  return (
                    <Fragment key={key}>
                      <tr {...rowProps} className={cx('ant-table-row', resolvedRowClassName, rowProps.className)}>
                        {rowSelection && (
                          <td className="ant-table-selection-column">
                            <input
                              type="checkbox"
                              checked={selectedKeys.includes(key)}
                              onChange={(event) => toggleSelected(key, record, event.target.checked)}
                            />
                          </td>
                        )}
                        {expandable && (
                          <td className="ant-table-row-expand-icon-cell">
                            {canExpand && (
                              <button
                                type="button"
                                className="ant-table-row-expand-icon"
                                onClick={() => setExpanded((current) => {
                                  const nextExpanded = !current.includes(key)
                                  expandable?.onExpand?.(nextExpanded, record)
                                  return nextExpanded ? [...current, key] : current.filter((item) => item !== key)
                                })}
                              >
                                {isExpanded ? '−' : '+'}
                              </button>
                            )}
                          </td>
                        )}
                        {visibleColumns.map((column, columnIndex) => {
                          const value = cellValue(record, column.dataIndex)
                          return (
                            <td
                              key={String(column.key ?? column.dataIndex ?? columnIndex)}
                              className={cx(column.className, column.ellipsis && 'ant-table-cell-ellipsis')}
                              style={{ width: column.width, textAlign: column.align }}
                            >
                              <div className="rh-table__cell-inner">
                                {column.render ? column.render(value, record, index) : value as ReactNode}
                              </div>
                            </td>
                          )
                        })}
                      </tr>
                      {canExpand && isExpanded && (
                        <tr key={`${String(key)}-expanded`} className="ant-table-expanded-row">
                          <td colSpan={visibleColumns.length + (rowSelection ? 1 : 0) + 1}>{expandable.expandedRowRender?.(record, index)}</td>
                        </tr>
                      )}
                    </Fragment>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      {hasPagination && (
        <div className="ant-table-pagination">
          <Pagination {...pagination} total={pagination.total ?? dataSource.length} pageSize={pageSize} />
        </div>
      )}
    </div>
  )
}

interface ModalProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  open?: boolean
  visible?: boolean
  title?: ReactNode
  footer?: ReactNode
  width?: number | string
  okText?: ReactNode
  cancelText?: ReactNode
  confirmLoading?: boolean
  onOk?: () => void
  onCancel?: () => void
  okButtonProps?: ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean; danger?: boolean }
  cancelButtonProps?: ButtonHTMLAttributes<HTMLButtonElement>
  destroyOnClose?: boolean
  destroyOnHidden?: boolean
  closable?: boolean
  centered?: boolean
  forceRender?: boolean
  [key: string]: any
}

export function Modal({
  open,
  visible,
  title,
  footer,
  width,
  okText,
  cancelText,
  confirmLoading,
  okButtonProps,
  cancelButtonProps,
  onOk,
  onCancel,
  className,
  children,
  destroyOnClose: _destroyOnClose,
  destroyOnHidden: _destroyOnHidden,
  closable: _closable,
  centered: _centered,
  forceRender: _forceRender,
  ...props
}: ModalProps) {
  if (!(open ?? visible)) return null
  const defaultFooter = footer === undefined ? (
    <>
      <button type="button" className="rh-btn ant-btn ant-btn-default" onClick={onCancel} {...cancelButtonProps}>{cancelText ?? 'Отмена'}</button>
      <button type="button" className="rh-btn ant-btn ant-btn-primary" disabled={confirmLoading || okButtonProps?.loading || okButtonProps?.disabled} onClick={onOk} {...okButtonProps}>{okText ?? 'OK'}</button>
    </>
  ) : footer
  return (
    <div className="rh-modal-root">
      <button type="button" className="rh-modal-mask" aria-label="Закрыть окно" onClick={onCancel} />
      <div className={cx('rh-modal ant-modal', className)} role="dialog" aria-modal="true" style={{ width }} {...props}>
        <div className="ant-modal-content">
          {title && <div className="ant-modal-header"><div className="ant-modal-title">{title}</div></div>}
          <div className="ant-modal-body">{children}</div>
          {defaultFooter && <div className="ant-modal-footer">{defaultFooter}</div>}
        </div>
      </div>
    </div>
  )
}

interface DrawerProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  open?: boolean
  visible?: boolean
  title?: ReactNode
  placement?: 'left' | 'right' | 'top' | 'bottom'
  width?: number | string
  extra?: ReactNode
  footer?: ReactNode
  onClose?: () => void
  size?: SizeToken | 'default' | 'large' | string
  forceRender?: boolean
  [key: string]: any
}

export function Drawer({ open, visible, title, placement = 'right', width = 420, extra, footer, onClose, size: _size, forceRender: _forceRender, className, children, ...props }: DrawerProps) {
  if (!(open ?? visible)) return null
  return (
    <div className={cx('rh-drawer-root ant-drawer', `ant-drawer-${placement}`)}>
      <button type="button" className="rh-modal-mask" aria-label="Закрыть панель" onClick={onClose} />
      <div className={cx('rh-drawer ant-drawer-content', className)} style={{ width }} role="dialog" aria-modal="true" {...props}>
        <div className="ant-drawer-header">
          <div className="ant-drawer-title">{title}</div>
          {extra}
          <button type="button" className="rh-drawer__close" aria-label="Закрыть" onClick={onClose}>×</button>
        </div>
        <div className="ant-drawer-body">{children}</div>
        {footer && <div className="ant-drawer-footer">{footer}</div>}
      </div>
    </div>
  )
}

function renderMenuItems(
  items: MenuItem[] = [],
  menuOnClick?: MenuProps['onClick'],
  onItemClick?: () => void,
  selectedKeys: Key[] = [],
) {
  return items.map((item, index) => {
    if (item.type === 'divider') return <li key={String(item.key ?? `divider-${index}`)} className="ant-dropdown-menu-item-divider" role="separator" />
    if (item.type === 'group') {
      return (
        <li key={String(item.key ?? `group-${textFromNode(item.label)}-${index}`)} className="ant-dropdown-menu-item-group">
          <div className="ant-dropdown-menu-item-group-title">{item.label}</div>
          <ul role="group">{renderMenuItems(item.children, menuOnClick, onItemClick, selectedKeys)}</ul>
        </li>
      )
    }
    const selected = item.key !== undefined && selectedKeys.some((key) => String(key) === String(item.key))
    return (
      <li
        key={String(item.key)}
        className={cx(
          'ant-dropdown-menu-item',
          selected && 'ant-dropdown-menu-item-selected',
          item.danger && 'ant-dropdown-menu-item-danger',
          item.disabled && 'ant-dropdown-menu-item-disabled',
        )}
        role="none"
      >
        <button
          type="button"
          role="menuitem"
          aria-current={selected ? 'page' : undefined}
          disabled={item.disabled}
          onClick={(event) => {
            item.onClick?.()
            menuOnClick?.({ key: String(item.key), item, domEvent: event })
            onItemClick?.()
          }}
        >
          {item.icon}
          <span>{item.label}</span>
        </button>
      </li>
    )
  })
}

interface DropdownProps extends HTMLAttributes<HTMLDivElement> {
  menu?: MenuProps
  trigger?: string[]
  placement?: string
  dropdownRender?: (menu: ReactNode) => ReactNode
  children?: ReactNode
  classNames?: Record<string, any>
  [key: string]: any
}

export function Dropdown({ menu, dropdownRender, className, classNames, trigger: _trigger, placement = 'bottomLeft', children, ...props }: DropdownProps) {
  const [open, setOpen] = useState(false)
  const close = () => setOpen(false)
  const menuNode = <ul className="ant-dropdown-menu" role="menu">{renderMenuItems(menu?.items, menu?.onClick, close, menu?.selectedKeys)}</ul>
  const childArray = Children.toArray(children)
  const trigger = childArray.length === 1 && isValidElement(childArray[0])
    ? cloneElement(childArray[0] as ReactElement<Record<string, any>>, {
      'aria-expanded': open,
      onClick: (event: ReactMouseEvent<HTMLElement>) => {
        setOpen((value) => !value)
        ;(childArray[0] as ReactElement<Record<string, any>>).props.onClick?.(event)
      },
    })
    : children
  const rootClassName = typeof classNames?.root === 'string' ? classNames.root : undefined
  const popupRootClassName = typeof classNames?.popup?.root === 'string' ? classNames.popup.root : undefined

  return (
    <span className={cx('rh-dropdown ant-dropdown', className)} {...props}>
      {trigger}
      {open && (
        <div
          className={cx('rh-dropdown__overlay ant-dropdown', `rh-dropdown__overlay--${placement}`, rootClassName, popupRootClassName)}
          data-placement={placement}
        >
          {dropdownRender ? dropdownRender(menuNode) : menuNode}
        </div>
      )}
    </span>
  )
}

interface PopoverProps extends Omit<HTMLAttributes<HTMLSpanElement>, 'content' | 'title'> {
  title?: ReactNode
  content?: ReactNode
  children?: ReactNode
  trigger?: string | string[]
  open?: boolean
  onOpenChange?: (open: boolean) => void
  placement?: string
  arrow?: boolean
  [key: string]: any
}

export function Popover({ title, content, className, children, trigger: _trigger, open, onOpenChange, placement: _placement, arrow: _arrow, ...props }: PopoverProps) {
  const [internalOpen, setInternalOpen] = useState(false)
  const visible = open ?? internalOpen
  const childArray = Children.toArray(children)
  const trigger = childArray.length === 1 && isValidElement(childArray[0])
    ? cloneElement(childArray[0] as ReactElement<Record<string, any>>, {
      onClick: (event: ReactMouseEvent<HTMLElement>) => {
        const next = !visible
        setInternalOpen(next)
        onOpenChange?.(next)
        ;(childArray[0] as ReactElement<Record<string, any>>).props.onClick?.(event)
      },
    })
    : children
  return (
    <span className={cx('rh-popover ant-popover', className)} {...props}>
      {trigger}
      {visible && (
        <span className="ant-popover-inner" role="tooltip">
          {title && <span className="ant-popover-title">{title}</span>}
          {content && <span className="ant-popover-inner-content">{content}</span>}
        </span>
      )}
    </span>
  )
}

interface TooltipProps extends Omit<HTMLAttributes<HTMLSpanElement>, 'title'> {
  title?: ReactNode
  children?: ReactNode
  placement?: string
  [key: string]: any
}

export function Tooltip({ title, className, children, placement: _placement, ...props }: TooltipProps) {
  return (
    <span className={cx('rh-tooltip ant-tooltip', className)} title={textFromNode(title)} {...props}>
      {children}
    </span>
  )
}

interface PopconfirmProps extends Omit<PopoverProps, 'content'> {
  okText?: ReactNode
  cancelText?: ReactNode
  description?: ReactNode
  okButtonProps?: ButtonHTMLAttributes<HTMLButtonElement> & { danger?: boolean; loading?: boolean }
  cancelButtonProps?: ButtonHTMLAttributes<HTMLButtonElement>
  onConfirm?: (event?: ReactMouseEvent<HTMLElement>) => void
  onCancel?: (event?: ReactMouseEvent<HTMLElement>) => void
  children?: ReactNode
}

const popconfirmRootStyle: CSSProperties = {
  position: 'relative',
  display: 'inline-flex',
}

const popconfirmOverlayStyle: CSSProperties = {
  position: 'absolute',
  right: 0,
  bottom: 'calc(100% + 10px)',
  zIndex: 1050,
  minWidth: 220,
}

export function Popconfirm({ title, description, okText, cancelText, okButtonProps, cancelButtonProps, onConfirm, onCancel, children, className, style, ...props }: PopconfirmProps) {
  const [open, setOpen] = useState(false)
  const childArray = Children.toArray(children)
  const trigger = childArray.length === 1 && isValidElement(childArray[0])
    ? cloneElement(childArray[0] as ReactElement<Record<string, any>>, {
      onClick: (event: ReactMouseEvent<HTMLElement>) => {
        event.preventDefault()
        setOpen(true)
        ;(childArray[0] as ReactElement<Record<string, any>>).props.onClick?.(event)
      },
    })
    : children
  return (
    <span className={cx('rh-popconfirm ant-popover ant-popconfirm', className)} style={{ ...popconfirmRootStyle, ...style }} {...props}>
      {trigger}
      {open && (
        <span className="ant-popover-inner" role="tooltip" style={popconfirmOverlayStyle}>
          <span className="ant-popover-inner-content">{title}{description && <small>{description}</small>}</span>
          <span className="ant-popconfirm-buttons">
            <button type="button" className="ant-btn" onClick={(event) => { setOpen(false); onCancel?.(event) }} {...cancelButtonProps}>{cancelText ?? 'Нет'}</button>
            <button type="button" className={cx('ant-btn ant-btn-primary', okButtonProps?.danger && 'ant-btn-dangerous')} onClick={(event) => { setOpen(false); onConfirm?.(event) }} {...okButtonProps}>{okText ?? 'Да'}</button>
          </span>
        </span>
      )}
    </span>
  )
}

interface UploadComponentProps extends UploadProps {
  drag?: boolean
}

const LIST_IGNORE = Symbol('LIST_IGNORE')

function UploadRoot({ accept, multiple, disabled, showUploadList = true, fileList, beforeUpload, onChange, children, className, drag, ...props }: UploadComponentProps) {
  const [internalFiles, setInternalFiles] = useState<UploadFile[]>([])
  const files = fileList ?? internalFiles
  const handleFiles = async (selectedFiles: File[]) => {
    const nextFiles: UploadFile[] = []
    for (const file of selectedFiles) {
      const decision = await beforeUpload?.(file, selectedFiles)
      if (decision === LIST_IGNORE) continue
      const uploadFile: UploadFile = { uid: `${Date.now()}-${file.name}`, name: file.name, originFileObj: file, status: decision === false ? 'ready' : 'done' }
      nextFiles.push(uploadFile)
      onChange?.({ file: uploadFile, fileList: [...files, uploadFile] })
    }
    if (nextFiles.length > 0) setInternalFiles((current) => [...current, ...nextFiles])
  }
  return (
    <div className={cx('rh-upload ant-upload-wrapper', className)} {...props}>
      <label className={cx('ant-upload', drag ? 'ant-upload-drag' : 'ant-upload-select')}>
        <input
          type="file"
          accept={accept}
          multiple={multiple}
          disabled={disabled}
          hidden
          onChange={(event) => {
            void handleFiles(Array.from(event.target.files ?? []))
            event.currentTarget.value = ''
          }}
        />
        <span className="ant-upload">{children}</span>
      </label>
      {showUploadList && files.length > 0 && (
        <div className="ant-upload-list">
          {files.map((file) => <div key={file.uid ?? file.name} className="ant-upload-list-item">{file.name}</div>)}
        </div>
      )}
    </div>
  )
}

function Dragger(props: UploadProps) {
  return <UploadRoot {...props} drag />
}

export const Upload = Object.assign(UploadRoot, { Dragger, LIST_IGNORE })

interface DescriptionsProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: ReactNode
  bordered?: boolean
  column?: number | Record<string, number>
  size?: SizeToken
  children?: ReactNode
}

interface DescriptionItemProps extends HTMLAttributes<HTMLDivElement> {
  label?: ReactNode
  span?: number
}

function DescriptionItem({ label, children, className, ...props }: DescriptionItemProps) {
  return (
    <div className={cx('ant-descriptions-item', className)} {...props}>
      <div className="ant-descriptions-item-label">{label}</div>
      <div className="ant-descriptions-item-content">{children}</div>
    </div>
  )
}

export const Descriptions = Object.assign(
  function DescriptionsRoot({ title, bordered, column: _column, size: _size, children, className, ...props }: DescriptionsProps) {
    return (
      <div className={cx('rh-descriptions ant-descriptions', bordered && 'ant-descriptions-bordered', className)} {...props}>
        {title && <div className="ant-descriptions-title">{title}</div>}
        <div className="ant-descriptions-view">
          <div className="ant-descriptions-row">{children}</div>
        </div>
      </div>
    )
  },
  { Item: DescriptionItem },
)

interface CollapseItem {
  key: Key
  label: ReactNode
  children?: ReactNode
}

interface CollapseProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  items?: CollapseItem[]
  defaultActiveKey?: Key | Key[]
  activeKey?: Key | Key[]
  accordion?: boolean
  onChange?: (key: Key | Key[]) => void
  ghost?: boolean
  [key: string]: any
}

function normalizeKeys(key?: Key | Key[]) {
  if (key == null) return []
  return Array.isArray(key) ? key : [key]
}

function Panel({ header, children }: { header?: ReactNode; children?: ReactNode }) {
  return <details className="ant-collapse-item"><summary className="ant-collapse-header">{header}</summary><div className="ant-collapse-content">{children}</div></details>
}

export const Collapse = Object.assign(
  function CollapseRoot({ items, defaultActiveKey, activeKey, onChange, className, ghost: _ghost, children, ...props }: CollapseProps) {
    const [internal, setInternal] = useState<Key[]>(normalizeKeys(defaultActiveKey))
    const active = normalizeKeys(activeKey ?? internal)
    const nodes = items
      ? items.map((item) => {
        const open = active.includes(item.key)
        return (
          <details
            key={String(item.key)}
            className="ant-collapse-item"
            open={open}
            onToggle={(event) => {
              const next = event.currentTarget.open ? [item.key] : active.filter((key) => key !== item.key)
              setInternal(next)
              onChange?.(next)
            }}
          >
            <summary className="ant-collapse-header">{item.label}</summary>
            <div className="ant-collapse-content">{item.children}</div>
          </details>
        )
      })
      : children
    return <div className={cx('rh-collapse ant-collapse', className)} {...props}>{nodes}</div>
  },
  { Panel },
)

interface TabsItem {
  key: string
  label: ReactNode
  children?: ReactNode
}

interface TabsProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  items?: TabsItem[]
  activeKey?: string
  defaultActiveKey?: string
  onChange?: (activeKey: string) => void
  children?: ReactNode
}

function TabPane({ children }: { tab?: ReactNode; children?: ReactNode }) {
  return <>{children}</>
}

export const Tabs = Object.assign(
  function TabsRoot({ items = [], activeKey, defaultActiveKey, onChange, className, ...props }: TabsProps) {
    const [internal, setInternal] = useState(defaultActiveKey ?? items[0]?.key)
    const active = activeKey ?? internal
    const activeItem = items.find((item) => item.key === active) ?? items[0]
    return (
      <div className={cx('rh-tabs ant-tabs', className)} {...props}>
        <div className="ant-tabs-nav">
          {items.map((item) => (
            <button
              key={item.key}
              type="button"
              className={cx('ant-tabs-tab', item.key === active && 'ant-tabs-tab-active')}
              onClick={() => {
                setInternal(item.key)
                onChange?.(item.key)
              }}
            >
              <span className="ant-tabs-tab-btn">{item.label}</span>
            </button>
          ))}
        </div>
        <div className="ant-tabs-content-holder">{activeItem?.children}</div>
      </div>
    )
  },
  { TabPane },
)

interface StepsProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  current?: number
  items?: Array<{ title?: ReactNode; description?: ReactNode; icon?: ReactNode; status?: string }>
  direction?: 'horizontal' | 'vertical'
  size?: SizeToken | string
  onChange?: (current: number) => void
  [key: string]: any
}

export function Steps({ current = 0, items = [], direction = 'horizontal', size: _size, onChange, className, ...props }: StepsProps) {
  return (
    <div className={cx('rh-steps ant-steps', direction === 'vertical' && 'ant-steps-vertical', className)} {...props}>
      {items.map((item, index) => (
        <button type="button" key={index} className={cx('ant-steps-item', index === current && 'ant-steps-item-active', index < current && 'ant-steps-item-finish', item.status && `ant-steps-item-${item.status}`)} onClick={() => onChange?.(index)}>
          <div className="ant-steps-item-icon">{item.icon ?? index + 1}</div>
          <div className="ant-steps-item-content">
            <div className="ant-steps-item-title">{item.title}</div>
            {item.description && <div className="ant-steps-item-description">{item.description}</div>}
          </div>
        </button>
      ))}
    </div>
  )
}

interface QRCodeProps extends HTMLAttributes<HTMLDivElement> {
  value?: string
  size?: number
}

export function QRCode({ value = '', size = 160, className, ...props }: QRCodeProps) {
  return (
    <div className={cx('rh-qr ant-qrcode', className)} style={{ width: size, height: size }} aria-label={value} {...props}>
      <div className="rh-qr__grid" />
    </div>
  )
}

interface ColorPickerProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'type'> {
  value?: string
  defaultValue?: string
  onChange?: (color: { toHexString: () => string }, hex: string) => void
  [key: string]: any
}

export function ColorPicker({ value, defaultValue = '#0f766e', onChange, className, ...props }: ColorPickerProps) {
  const [internal, setInternal] = useState(defaultValue)
  const current = value ?? internal
  return (
    <input
      className={cx('rh-color-picker ant-color-picker-trigger', className)}
      type="color"
      value={current}
      onChange={(event) => {
        setInternal(event.target.value)
        onChange?.({ toHexString: () => event.target.value }, event.target.value)
      }}
      {...props}
    />
  )
}
