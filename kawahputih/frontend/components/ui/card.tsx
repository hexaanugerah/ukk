import { cn } from "@/lib/utils";
import { HTMLAttributes } from "react";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "rounded-card border border-borderc bg-white shadow-sm dark:border-borderc-dark dark:bg-surface-dark",
        className
      )}
      {...props}
    />
  );
}
