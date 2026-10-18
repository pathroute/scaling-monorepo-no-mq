import { useId } from "react";

export type TextFieldProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
};

/** A labelled text input. */
export function TextField({ label, value, onChange, placeholder }: TextFieldProps) {
  const id = useId();
  return (
    <div className="sk-field">
      <label htmlFor={id}>{label}</label>
      <input id={id} value={value} placeholder={placeholder} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}
