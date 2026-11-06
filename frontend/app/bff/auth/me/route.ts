import { NextResponse } from "next/server";
import { clearAuthCookie, getAuthToken } from "@/lib/auth-cookie";

const API_URL = process.env.API_URL ?? "http://localhost:8080";

export async function GET() {
  const token = await getAuthToken();
  if (!token) {
    return NextResponse.json({ user: null }, { status: 200 });
  }

  try {
    const upstream = await fetch(`${API_URL}/api/auth/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });

    if (!upstream.ok) {
      // Token expired or user deleted: clean up the stale cookie.
      await clearAuthCookie();
      return NextResponse.json({ user: null }, { status: 200 });
    }

    const data = await upstream.json();
    return NextResponse.json({ user: data.user }, { status: 200 });
  } catch {
    return NextResponse.json(
      { error: { code: "bad_gateway", message: "Backend API is unreachable" } },
      { status: 502 }
    );
  }
}

