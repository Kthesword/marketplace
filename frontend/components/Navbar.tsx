"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

export function Navbar() {
  const { user, loading, logout } = useAuth();
  const router = useRouter();

  async function handleLogout() {
    await logout();
    router.push("/login");
    router.refresh();
  }

  return (
    <header className="border-b border-gray-200 bg-white">
      <nav className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/" className="text-xl font-bold text-indigo-600">
          Marketplace
        </Link>

        <div className="flex items-center gap-5 text-sm font-medium text-gray-700">
          <Link href="/" className="hover:text-indigo-600">
            Browse
          </Link>

          {!loading && user?.role === "buyer" && (
            <>
              <Link href="/cart" className="hover:text-indigo-600">
                Cart
              </Link>
              <Link href="/orders" className="hover:text-indigo-600">
                My Orders
              </Link>
            </>
          )}

          {!loading && user?.role === "seller" && (
            <>
              <Link href="/seller/dashboard" className="hover:text-indigo-600">
                Seller Dashboard
              </Link>
              <Link href="/seller/store" className="hover:text-indigo-600">
                Store
              </Link>
              <Link href="/seller/products" className="hover:text-indigo-600">
                Products
              </Link>
              <Link href="/seller/orders" className="hover:text-indigo-600">
                Orders
              </Link>
            </>
          )}

          {loading ? (
            <span className="text-xs text-gray-400">...</span>
          ) : user ? (
            <div className="flex items-center gap-3">
              <span className="rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-700">
                {user.name} ({user.role})
              </span>
              <button
                type="button"
                onClick={handleLogout}
                className="rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
              >
                Log out
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Link href="/login" className="hover:text-indigo-600">
                Log in
              </Link>
              <Link
                href="/register"
                className="rounded-md bg-indigo-600 px-3 py-1.5 text-white hover:bg-indigo-500"
              >
                Register
              </Link>
            </div>
          )}
        </div>
      </nav>
    </header>
  );
}
