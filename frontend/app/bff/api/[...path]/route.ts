import { NextRequest, NextResponse } from "next/server";
import { getAuthToken } from "@/lib/auth-cookie";

const API_URL = process.env.API_URL ?? "http://localhost:8080";

// Catch-all BFF proxy: forwards /bff/api/<anything> to Go's /api/<anything>.
// Automatically reads the httpOnly `mp_token` cookie and sends it as
// `Authorization: Bearer <jwt>` to the Go API.
async function proxy(
  req: NextRequest,
  context: { params: Promise<{ path: string[] }> }
): Promise<NextResponse> {
  const { path } = await context.params;
  const targetUrl = new URL(`/api/${path.join("/")}`, API_URL);
  targetUrl.search = req.nextUrl.search;

  const headers = new Headers();
  const contentType = req.headers.get("content-type");
  if (contentType) {
    headers.set("content-type", contentType);
  }

  const token = await getAuthToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const hasBody = req.method !== "GET" && req.method !== "HEAD";
  const body = hasBody ? await req.text() : undefined;

  try {
    const upstream = await fetch(targetUrl, {
      method: req.method,
      headers,
      body,
      cache: "no-store",
    });

    if (upstream.status === 204) {
      return new NextResponse(null, { status: 204 });
    }

    const respText = await upstream.text();
    return new NextResponse(respText, {
      status: upstream.status,
      headers: {
        "content-type":
          upstream.headers.get("content-type") ?? "application/json",
      },
    });
  } catch {
    return NextResponse.json(
      { error: { code: "bad_gateway", message: "Backend API is unreachable" } },
      { status: 502 }
    );
  }
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
