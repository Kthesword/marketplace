import { NextRequest, NextResponse } from "next/server";
import { AUTH_COOKIE } from "@/lib/auth-cookie";
import type { Role } from "@/lib/types";

interface DecodedJwtPayload {
  uid?: number;
  role?: Role;
  exp?: number;
}

// Decodes the middle part of a JWT (header.payload.signature) to inspect
// `role` and `exp` for UX route guarding.
// Security note: actual authorization is ALWAYS enforced by the Go API using
// the cryptographic HMAC signature.
function decodePayload(token: string): DecodedJwtPayload | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const json = atob(base64);
    const payload = JSON.parse(json) as DecodedJwtPayload;
    if (payload.exp && payload.exp * 1000 < Date.now()) {
      return null; // expired
    }
    return payload;
  } catch {
    return null;
  }
}

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get(AUTH_COOKIE)?.value;
  const claims = token ? decodePayload(token) : null;

  const isSellerRoute = pathname.startsWith("/seller");
  const isBuyerRoute =
    pathname.startsWith("/cart") ||
    pathname.startsWith("/checkout") ||
    pathname.startsWith("/orders");
  const isAuthPage = pathname === "/login" || pathname === "/register";

  // 1. Unauthenticated user visiting a protected route -> /login
  if ((isSellerRoute || isBuyerRoute) && !claims) {
    const loginUrl = new URL("/login", req.url);
    return NextResponse.redirect(loginUrl);
  }

  // 2. Role mismatch: buyer trying to open /seller/* -> /
  if (isSellerRoute && claims?.role !== "seller") {
    return NextResponse.redirect(new URL("/", req.url));
  }

  // 3. Role mismatch: seller trying to open buyer-only routes (/cart, /checkout, /orders) -> /seller/dashboard
  if (isBuyerRoute && claims?.role !== "buyer") {
    return NextResponse.redirect(new URL("/seller/dashboard", req.url));
  }

  // 4. Already logged in visiting /login or /register -> redirect to their home
  if (isAuthPage && claims) {
    const dest = claims.role === "seller" ? "/seller/dashboard" : "/";
    return NextResponse.redirect(new URL(dest, req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/seller/:path*",
    "/cart/:path*",
    "/checkout/:path*",
    "/orders/:path*",
    "/login",
    "/register",
  ],
};

