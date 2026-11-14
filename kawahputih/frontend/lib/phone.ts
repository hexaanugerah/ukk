// Helper nomor telepon internasional: pilih kode negara, ketik nomor tanpa 0.
// Format tersimpan: E.164 (contoh: +6281234567890) — sesuai batas backend
// (maks 20 karakter, min 8 untuk booking).

export interface CountryCode {
  dial: string; // termasuk "+", mis. "+62"
  iso: string; // kode negara, mis. "ID"
  name: string;
}

/** Daftar kode negara yang bisa dipilih di samping input telepon. */
export const COUNTRY_CODES: CountryCode[] = [
  { dial: "+62", iso: "ID", name: "Indonesia" },
  { dial: "+60", iso: "MY", name: "Malaysia" },
  { dial: "+65", iso: "SG", name: "Singapura" },
  { dial: "+66", iso: "TH", name: "Thailand" },
  { dial: "+63", iso: "PH", name: "Filipina" },
  { dial: "+84", iso: "VN", name: "Vietnam" },
  { dial: "+61", iso: "AU", name: "Australia" },
  { dial: "+81", iso: "JP", name: "Jepang" },
  { dial: "+82", iso: "KR", name: "Korea Selatan" },
  { dial: "+86", iso: "CN", name: "China" },
  { dial: "+91", iso: "IN", name: "India" },
  { dial: "+971", iso: "AE", name: "Uni Emirat Arab" },
  { dial: "+44", iso: "GB", name: "Inggris" },
  { dial: "+1", iso: "US", name: "Amerika Serikat / Kanada" },
];

export const DEFAULT_DIAL = "+62";

const DIALS_BY_LENGTH = [...COUNTRY_CODES].sort((a, b) => b.dial.length - a.dial.length);

export interface ParsedPhone {
  dial: string;
  /** Digit nasional tanpa kode negara & tanpa awalan 0. */
  national: string;
}

/**
 * Memecah nomor apa pun (08..., 62..., +62..., +xx...) menjadi kode negara
 * dan digit nasional. Nomor lokal "0..." dianggap Indonesia.
 */
export function splitPhone(value?: string | null): ParsedPhone {
  const clean = (value ?? "").replace(/[\s\-().]/g, "");
  if (!clean) return { dial: DEFAULT_DIAL, national: "" };
  if (clean.startsWith("+")) {
    const hit = DIALS_BY_LENGTH.find((c) => clean.startsWith(c.dial));
    if (hit) return { dial: hit.dial, national: clean.slice(hit.dial.length) };
    return { dial: DEFAULT_DIAL, national: clean.slice(1) };
  }
  if (clean.startsWith("0")) return { dial: DEFAULT_DIAL, national: clean.slice(1) };
  if (clean.startsWith("62")) return { dial: DEFAULT_DIAL, national: clean.slice(2) };
  return { dial: DEFAULT_DIAL, national: clean };
}

/** Menggabungkan kode negara + digit nasional menjadi format E.164 ("+62..."). */
export function joinPhone(dial: string, national: string): string {
  const nat = (national ?? "").replace(/\D/g, "");
  return nat ? `${dial}${nat}` : "";
}

/** Validasi nomor lengkap (boleh format lama "08..." maupun E.164). */
export function validatePhoneFull(
  value?: string | null,
  required = false
): { valid: boolean; message?: string } {
  if (!value || !value.trim()) {
    return required ? { valid: false, message: "Nomor telepon wajib diisi" } : { valid: true };
  }
  const { dial, national } = splitPhone(value);
  if (/\D/.test(national)) {
    return { valid: false, message: "Nomor telepon hanya boleh berisi angka" };
  }
  if (dial === DEFAULT_DIAL) {
    if (!national.startsWith("8")) {
      return { valid: false, message: "Nomor Indonesia diawali 8, contoh: 81234567890 (tanpa 0)" };
    }
    if (national.length < 9 || national.length > 12) {
      return { valid: false, message: "Nomor Indonesia harus 9-12 digit" };
    }
    return { valid: true };
  }
  if (national.length < 6 || national.length > 14) {
    return { valid: false, message: "Nomor telepon harus 6-14 digit" };
  }
  return { valid: true };
}
