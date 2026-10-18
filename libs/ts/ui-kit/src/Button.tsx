import type { ButtonHTMLAttributes } from "react";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary";
};

/** The suite's button; defaults to type="button" so it never submits a form by accident. */
export function Button({ variant = "primary", type = "button", className, ...rest }: ButtonProps) {
  const classes = ["sk-button", `sk-button--${variant}`, className].filter(Boolean).join(" ");
  return <button type={type} className={classes} {...rest} />;
}
