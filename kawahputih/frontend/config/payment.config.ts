// Payment config — daftar metode pembayaran yang didukung. Kode di sini HARUS sama persis dengan konstanta method di backend (payments/entity.go) karena dikirim apa adanya pada POST /payments/process.

export type PaymentMethodCode =
  | "QRIS"
  | "GOPAY"
  | "OVO"
  | "DANA"
  | "SHOPEEPAY"
  | "BCA_VA"
  | "BNI_VA"
  | "MANDIRI_VA"
  | "BRI_VA"
  | "CARD"
  | "ALFAMART"
  | "INDOMARET";

/** Cara pembayaran diselesaikan — menentukan instruksi yang ditampilkan. */
export type PaymentMethodKind = "qris" | "ewallet" | "va" | "card" | "retail";

export interface PaymentMethodOption {
  code: PaymentMethodCode;
  label: string;
  kind: PaymentMethodKind;
  hint: string;
}

export interface PaymentMethodGroup {
  key: string;
  label: string;
  options: PaymentMethodOption[];
}

export const PAYMENT_METHOD_GROUPS: PaymentMethodGroup[] = [
  {
    key: "qris",
    label: "QRIS",
    options: [
      {
        code: "QRIS",
        label: "QRIS",
        kind: "qris",
        hint: "Pindai dengan aplikasi bank atau e-wallet apa pun",
      },
    ],
  },
  {
    key: "ewallet",
    label: "E-Wallet",
    options: [
      { code: "GOPAY", label: "GoPay", kind: "ewallet", hint: "Bayar lewat aplikasi Gojek" },
      { code: "OVO", label: "OVO", kind: "ewallet", hint: "Bayar lewat aplikasi OVO" },
      { code: "DANA", label: "DANA", kind: "ewallet", hint: "Bayar lewat aplikasi DANA" },
      { code: "SHOPEEPAY", label: "ShopeePay", kind: "ewallet", hint: "Bayar lewat aplikasi Shopee" },
    ],
  },
  {
    key: "va",
    label: "Transfer Bank / ATM",
    options: [
      { code: "BCA_VA", label: "BCA Virtual Account", kind: "va", hint: "ATM, m-BCA, atau internet banking" },
      { code: "BNI_VA", label: "BNI Virtual Account", kind: "va", hint: "ATM, BNI Mobile, atau internet banking" },
      { code: "MANDIRI_VA", label: "Mandiri Virtual Account", kind: "va", hint: "ATM, Livin', atau internet banking" },
      { code: "BRI_VA", label: "BRI Virtual Account", kind: "va", hint: "ATM, BRImo, atau internet banking" },
    ],
  },
  {
    key: "card",
    label: "Kartu Kredit / Debit",
    options: [
      { code: "CARD", label: "Kartu Kredit / Debit", kind: "card", hint: "Visa, Mastercard, atau JCB" },
    ],
  },
  {
    key: "retail",
    label: "Gerai Retail",
    options: [
      { code: "ALFAMART", label: "Alfamart", kind: "retail", hint: "Bayar tunai di kasir Alfamart" },
      { code: "INDOMARET", label: "Indomaret", kind: "retail", hint: "Bayar tunai di kasir Indomaret" },
    ],
  },
];

/** Semua opsi dalam satu array datar (untuk lookup kode → label/kind). */
export const PAYMENT_METHODS: PaymentMethodOption[] = PAYMENT_METHOD_GROUPS.flatMap(
  (group) => group.options
);

export function findPaymentMethod(code: string): PaymentMethodOption | undefined {
  return PAYMENT_METHODS.find((m) => m.code === code);
}

/** Label ramah-baca untuk sebuah kode metode (fallback: kode aslinya). */
export function paymentMethodLabel(code: string): string {
  return findPaymentMethod(code)?.label ?? code;
}
