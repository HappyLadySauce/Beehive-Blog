import React, { ChangeEvent } from 'react';
import CustomSelect from './CustomSelect';

// ─── 公共类型 ────────────────────────────────────────────────────────────────

interface BaseFieldProps {
  label: string;
  /** 标签右侧的辅助说明文字（灰色小字）*/
  hint?: string;
  /** 显示红色星号 */
  required?: boolean;
  disabled?: boolean;
  className?: string;
}

const inputBase =
  'w-full px-3 text-[clamp(0.86rem,0.14vw+0.82rem,0.98rem)] border border-border rounded bg-input-background text-foreground admin-control-md focus:ring-2 focus:ring-ring focus:border-transparent disabled:bg-muted disabled:text-muted-foreground';

function FieldLabel({
  label,
  hint,
  required,
}: Pick<BaseFieldProps, 'label' | 'hint' | 'required'>) {
  return (
    <label className="mb-1 block text-[clamp(0.86rem,0.13vw+0.82rem,0.96rem)] font-medium text-foreground">
      {label}
      {required && <span className="text-red-500 ml-0.5">*</span>}
      {hint && <span className="text-muted-foreground font-normal ml-1">{hint}</span>}
    </label>
  );
}

// ─── TextField ───────────────────────────────────────────────────────────────

interface TextFieldProps extends BaseFieldProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: 'text' | 'email' | 'url';
}

/**
 * 单行文本输入框。
 *
 * @example
 * <TextField label="名称" value={form.name} onChange={(v) => setForm(f => ({ ...f, name: v }))} required />
 */
export function TextField({
  label,
  hint,
  required,
  disabled,
  className,
  value,
  onChange,
  placeholder,
  type = 'text',
}: TextFieldProps) {
  return (
    <div className={className}>
      <FieldLabel label={label} hint={hint} required={required} />
      <input
        type={type}
        value={value}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={inputBase}
      />
    </div>
  );
}

// ─── PasswordField ───────────────────────────────────────────────────────────

interface PasswordFieldProps extends BaseFieldProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

/**
 * 密码输入框，与 TextField 视觉保持一致。
 *
 * @example
 * <PasswordField label="初始密码" value={form.password} onChange={(v) => setForm(f => ({ ...f, password: v }))} required />
 */
export function PasswordField({
  label,
  hint,
  required,
  disabled,
  className,
  value,
  onChange,
  placeholder,
}: PasswordFieldProps) {
  return (
    <div className={className}>
      <FieldLabel label={label} hint={hint} required={required} />
      <input
        type="password"
        value={value}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={inputBase}
        autoComplete="new-password"
      />
    </div>
  );
}

// ─── TextareaField ───────────────────────────────────────────────────────────

interface TextareaFieldProps extends BaseFieldProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  rows?: number;
}

/**
 * 多行文本域，默认禁止手动拉伸。
 *
 * @example
 * <TextareaField label="描述" value={form.description} onChange={(v) => setForm(f => ({ ...f, description: v }))} />
 */
export function TextareaField({
  label,
  hint,
  required,
  disabled,
  className,
  value,
  onChange,
  placeholder,
  rows = 3,
}: TextareaFieldProps) {
  return (
    <div className={className}>
      <FieldLabel label={label} hint={hint} required={required} />
      <textarea
        value={value}
        onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onChange(e.target.value)}
        placeholder={placeholder}
        rows={rows}
        disabled={disabled}
        className={`${inputBase} resize-none`}
      />
    </div>
  );
}

// ─── SelectField ─────────────────────────────────────────────────────────────

interface SelectOption {
  value: string;
  label: string;
}

interface SelectFieldProps extends BaseFieldProps {
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
}

/**
 * 下拉选择框。
 *
 * @example
 * <SelectField
 *   label="角色"
 *   value={form.role}
 *   onChange={(v) => setForm(f => ({ ...f, role: v as UserRole }))}
 *   options={[{ value: 'user', label: '用户' }, { value: 'admin', label: '管理员' }]}
 * />
 */
export function SelectField({
  label,
  hint,
  required,
  disabled,
  className,
  value,
  onChange,
  options,
}: SelectFieldProps) {
  return (
    <div className={className}>
      <FieldLabel label={label} hint={hint} required={required} />
      <CustomSelect
        value={value}
        onChange={onChange}
        disabled={disabled}
        options={options}
        className="w-full"
        ariaLabel={label}
      />
    </div>
  );
}

// ─── ColorField ──────────────────────────────────────────────────────────────

interface ColorFieldProps extends BaseFieldProps {
  value: string;
  onChange: (value: string) => void;
}

/**
 * 颜色选择器：颜色块 + 十六进制文本同步输入，适用于标签颜色等场景。
 *
 * @example
 * <ColorField label="颜色" value={form.color} onChange={(v) => setForm(f => ({ ...f, color: v }))} />
 */
export function ColorField({
  label,
  hint,
  required,
  disabled,
  className,
  value,
  onChange,
}: ColorFieldProps) {
  return (
    <div className={className}>
      <FieldLabel label={label} hint={hint} required={required} />
      <div className="flex items-center gap-3">
        <input
          type="color"
          value={value}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
          disabled={disabled}
          className="h-[var(--admin-control-md)] w-[var(--admin-control-md)] rounded border border-border cursor-pointer p-0.5 disabled:cursor-not-allowed"
        />
        <input
          type="text"
          value={value}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
          placeholder="#3B82F6"
          disabled={disabled}
          className={`flex-1 ${inputBase}`}
        />
      </div>
    </div>
  );
}
