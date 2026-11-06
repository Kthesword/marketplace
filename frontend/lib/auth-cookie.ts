import { cookies } from "next/headers";

export const AUTH_COOKIE = "mp_token";

// Cookie options used when setting the JWT cookie on login/register.
// httpOnly: browser JS (document.cookie) cannot read it — protects against XSS token theft.
// sameSite: "lax" — sent on normal navigation and same-origin requests, blocked on cross-site POSTs (CSRF defense).
export async function setAuthCookie(token: string) {
  const jar = await cookies();
  jar.set(AUTH_COOKIE, token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 60 * 60 * 24, // 24 hours (matches backend JWT_TTL)
  });
}

export async function clearAuthCookie() {
  const jar = await cookies();
  jar.delete(AUTH_COOKIE);
}

export async function getAuthToken(): Promise<string | undefined> {
  const jar = await cookies();
  return jar.get(AUTH_COOKIE)?.value;
}

