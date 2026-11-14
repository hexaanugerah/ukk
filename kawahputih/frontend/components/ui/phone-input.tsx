"use client";

import { useEffect, useState } from "react";
import { COUNTRY_CODES, DEFAULT_DIAL, joinPhone, splitPhone } from "@/lib/phone";

interface PhoneInputProps {
  /** Nomor lengkap terkontrol, mis. "+6281234567890" atau "0812..." (lama). */
  value: string;
  /** Dipanggil dengan nomor lengkap format E.164; kosong jika belum diisi. */
  onChange: (full: string) => void;
  id?: string;
  autoComplete?: string;
  placeholder?: string;
  invalid?: boolean;
  /** Class untuk kotak input (tanpa class rounded penuh — dipasang internal). */
  inputClassName?: string;
  /** Class untuk dropdown kode negara. */
  selectClassName?: string;
  ariaLabel?: string;
}

// Input telepon dengan pilihan kode negara di sebelah kiri. User cukup
// mengetik nomor nasional tanpa awalan 0 (mis. 81234567890) — kode negara
// otomatis ditempelkan saat disimpan (format E.164).
export function PhoneInput({
  value,
  onChange,
  id,
  autoComplete = "tel-national",
  placeholder = "81234567890",
  invalid,
  inputClassName = "h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-brand-600",
  selectClassName = "h-10 rounded-md border border-slate-300 bg-slate-50 px-2 text-sm text-slate-600 outline-none focus:border-brand-600",
  ariaLabel = "Nomor telepon",
}: PhoneInputProps) {
  const parsed = splitPhone(value);
  // Kode negara dipertahankan lokal agar tetap terpilih meski nomor kosong.
  const [dial, setDial] = useState(parsed.dial);
  useEffect(() => {
    if (parsed.national) setDial(parsed.dial);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  function handleNational(raw: string) {
    const nat = raw.replace(/\D/g, "").replace(/^0+/, "");
    onChange(joinPhone(dial, nat));
  }

  function handleDial(next: string) {
    setDial(next);
    if (parsed.national) onChange(joinPhone(next, parsed.national));
  }

  return (
    <span className="flex gap-2">
      <select
        value={dial}
        onChange={(e) => handleDial(e.target.value)}
        className={selectClassName}
        aria-label="Kode negara"
      >
        {COUNTRY_CODES.map((c) => (
          <option key={c.dial} value={c.dial}>
            {c.dial} ({c.iso})
          </option>
        ))}
      </select>
      <input
        id={id}
        type="tel"
        inputMode="numeric"
        value={parsed.national}
        onChange={(e) => handleNational(e.target.value)}
        placeholder={placeholder}
        autoComplete={autoComplete}
        className={inputClassName}
        aria-label={ariaLabel}
        aria-invalid={invalid}
      />
    </span>
  );
}

export { DEFAULT_DIAL };
