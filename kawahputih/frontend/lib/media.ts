// Helper untuk URL media (berkas unggahan) yang disajikan backend di /uploads.

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://localhost:8080/api/v1";

// Origin backend: base URL API tanpa akhiran /api/v1.
export const backendOrigin = API_BASE_URL.replace(/\/api\/v1\/?$/, "");

/**
 * Mengubah path media relatif (mis. "/uploads/gallery/x.jpg") menjadi URL
 * absolut ke backend. URL absolut dilewatkan apa adanya.
 */
export function mediaUrl(path?: string | null): string | undefined {
  if (!path) return undefined;
  if (/^https?:\/\//i.test(path)) return path;
  if (path.startsWith("/")) return `${backendOrigin}${path}`;
  return path;
}
