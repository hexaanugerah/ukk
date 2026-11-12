import { NextRequest, NextResponse } from "next/server";

// Middleware edge-level: hanya cek KEHADIRAN cookie sesi mock (bukan validasi
// role — role dicek di client oleh useRouteGuard, karena sesi mock tersimpan
// di localStorage). Ini UX optimization: redirect cepat sebelum halaman
// protected sempat ter-render, bukan security boundary.
const PROTECTED_PREFIXES = ["/dashboard", "/admin", "/manager", "/petugas"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasSession = request.cookies.has("kpr_access_token");

  const isProtected = PROTECTED_PREFIXES.some((p) => pathname.startsWith(p));
  if (isProtected && !hasSession) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/admin/:path*", "/manager/:path*", "/petugas/:path*"],
};
