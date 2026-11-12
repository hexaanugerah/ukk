import { cn } from "@/lib/utils";
import { InputHTMLAttributes, forwardRef } from "react";

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
  "w-full rounded-input border border-borderc bg-white px-3 py-2 text-sm text-text-primary outline-none transition-all duration-150 placeholder:text-text-muted focus:border-brand focus:ring-1 focus:ring-brand dark:border-borderc-dark dark:bg-slate-900 dark:text-text-primary-dark",
  className
)}
      {...props}
    />
  )
);
Input.displayName = "Input";
