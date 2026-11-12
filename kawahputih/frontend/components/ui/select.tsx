import { cn } from "@/lib/utils";
import { SelectHTMLAttributes, forwardRef } from "react";

// select bawaan yang ditata agar cocok dengan Input/Textarea. Sengaja sederhana agar bekerja dengan register() react-hook-form (uncontrolled).
export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, children, ...props }, ref) => (
    <select
      ref={ref}
      className={cn(
        "w-full rounded-input border border-borderc bg-white px-3 py-2 text-sm text-text-primary outline-none transition-all duration-150 focus:border-brand focus:ring-1 focus:ring-brand dark:border-borderc-dark dark:bg-slate-900 dark:text-text-primary-dark",
        className
      )}
      {...props}
    >
      {children}
    </select>
  )
);
Select.displayName = "Select";
