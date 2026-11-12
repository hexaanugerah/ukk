import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";
import { ButtonHTMLAttributes, forwardRef } from "react";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-btn font-body text-sm font-semibold transition-all duration-150 disabled:opacity-60 disabled:pointer-events-none active:scale-[0.98]",
  {
    variants: {
      variant: {
        primary: "bg-brand text-white shadow-sm hover:bg-brand-dark",
        outline: "border border-borderc hover:bg-slate-50 dark:border-borderc-dark dark:hover:bg-slate-800",
        secondary: "bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700",
        ghost: "hover:bg-slate-100 dark:hover:bg-slate-800",
        destructive: "bg-danger text-white hover:brightness-90",
        success: "bg-success text-white hover:brightness-90",
      },
      size: {
        sm: "h-8 px-3",
        md: "h-10 px-4",
        lg: "h-12 px-6 text-base",
      },
    },
    defaultVariants: { variant: "primary", size: "md" },
  }
);

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  loading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, loading, children, disabled, ...props }, ref) => (
    <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} disabled={disabled || loading} {...props}>
      {loading && <Loader2 className="h-4 w-4 animate-spin" />}
      {children}
    </button>
  )
);
Button.displayName = "Button";
