import type { FormInstance } from 'antd'

/**
 * Reads current DOM input values for the given Antd Form field names and
 * writes any that diverge from form state back into the form. Call this
 * right before submit to defeat Chrome password-manager autofill that
 * populates `input.value` without firing React-compatible change events.
 *
 * Each name must match a `name` prop on `<Form.Item name="...">` AND an
 * `id="..."` on the underlying input (antd derives the id from name when no
 * custom id is set — the common case here).
 *
 * The generic is intentionally loose: antd's FormInstance<T> has a recursive
 * NamePath type that's painful to satisfy from a plain string list, so we
 * accept `FormInstance<any>` and trust callers to pass matching field names.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function syncAntdFormFromDOM(form: FormInstance<any>, names: readonly string[]): void {
  const patch: Record<string, string> = {}
  for (const name of names) {
    const el = document.getElementById(name) as HTMLInputElement | null
    if (!el) continue
    const domValue = el.value ?? ''
    const formValue = form.getFieldValue(name) as string | undefined
    if (domValue && domValue !== (formValue ?? '')) {
      patch[name] = domValue
    }
  }
  if (Object.keys(patch).length > 0) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    form.setFieldsValue(patch as any)
  }
}
