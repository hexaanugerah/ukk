import { format, parseISO } from "date-fns";
import { id } from "date-fns/locale";

export function formatDate(dateStr: string, pattern = "d MMMM yyyy"): string {
  try {
    return format(parseISO(dateStr), pattern, { locale: id });
  } catch {
    return dateStr;
  }
}

export function formatDateTime(dateStr: string): string {
  return formatDate(dateStr, "d MMMM yyyy, HH:mm");
}
