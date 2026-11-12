// Nominal dari backend selalu bilangan bulat dalam sen (lihat field
// price_cents / unit_cents), bukan pecahan, sehingga pembagian 100 untuk
// format dilakukan di satu tempat ini saja, bukan di setiap pemanggil.
export function formatCurrency(cents: number, currency = "IDR"): string {
  const amount = cents / 100;
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}
