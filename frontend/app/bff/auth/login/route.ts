import { NextRequest, NextResponse } from "next/server";
import { setAuthCookie } from "@/lib/auth-cookie";

const API_URL = process.env.API_URL ?? "http://localhost:8080";

export async function POST(req: NextRequest) {
  const body = await req.text();
  try {
    const upstream = await fetch(`${API_URL}/api/auth/login`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body,
      cache: "no-store",
    });

    const data = await upstream.json();
    if (!upstream.ok) {
      return NextResponse.json(data, { status: upstream.status });
    }

    // Store JWT in httpOnly cookie and strip token from browser response.
    await setAuthCookie(data.token);
    return NextResponse.json({ user: data.user }, { status: 200 });
  } catch {
    return NextResponse.json(
      { error: { code: "bad_gateway", message: "Backend API is unreachable" } },
      { status: 502 }
    );
  }
}

