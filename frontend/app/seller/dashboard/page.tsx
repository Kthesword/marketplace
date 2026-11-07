"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { api, ApiClientError } from "@/lib/api";
import type { Product, Store } from "@/lib/types";
import { ErrorState, LoadingState } from "@/components/states";

export default function SellerDashboardPage() {
  const { user } = useAuth();
  const [store, setStore] = useState<Store | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadDashboard = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [storeRes, productsRes] = await Promise.allSettled([
        api.seller.getStore(),
        api.seller.listProducts(1, 100),
      ]);

      if (storeRes.status === "fulfilled") {
        setStore(storeRes.value.store);
      } else if (
        storeRes.reason instanceof ApiClientError &&
        storeRes.reason.status === 404
      ) {
        setStore(null);
      } else {
        throw storeRes.reason;
      }

      if (productsRes.status === "fulfilled") {
        setProducts(productsRes.value.items);
      } else {
        throw productsRes.reason;
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load dashboard data"
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);

  const activeCount = products.filter((p) => p.active).length;
  const outOfStockCount = products.filter((p) => p.stock === 0).length;

  return (
    <main className="mx-auto max-w-5xl px-6 py-10">
      <h1 className="text-2xl font-bold text-gray-900">Seller Dashboard</h1>
      <p className="mt-1 text-sm text-gray-600">
        Welcome back{user ? `, ${user.name}` : ""}! Manage your storefront, catalog, and customer orders.
      </p>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading seller overview..." />
        ) : error ? (
          <ErrorState message={error} onRetry={loadDashboard} />
        ) : (
          <div className="space-y-6">
            {!store ? (
              <div className="rounded-xl border border-amber-300 bg-amber-50 p-6">
                <h2 className="text-base font-semibold text-amber-900">
                  Step 1: Set up your store
                </h2>
                <p className="mt-1 text-sm text-amber-800">
                  You need a store name and contact email before you can publish products.
                </p>
                <Link
                  href="/seller/store"
                  className="mt-4 inline-block rounded-md bg-amber-600 px-4 py-2 text-sm font-semibold text-white hover:bg-amber-500"
                >
                  Create Store Now
                </Link>
              </div>
            ) : (
              <div className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
                <div>
                  <span className="text-xs font-semibold uppercase tracking-wider text-indigo-600">
                    Active Storefront
                  </span>
                  <h2 className="mt-1 text-xl font-bold text-gray-900">
                    {store.name}
                  </h2>
                  <p className="mt-0.5 text-sm font-medium text-gray-700">
                    {store.description || "No description set"} ·{" "}
                    <span className="font-semibold text-gray-900">{store.contact_email}</span>
                  </p>
                </div>
                <div className="flex gap-3">
                  <Link
                    href={`/sellers/${store.seller_id}`}
                    className="rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                  >
                    View Public Profile
                  </Link>
                  <Link
                    href="/seller/store"
                    className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
                  >
                    Edit Store
                  </Link>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
                <p className="text-xs font-bold uppercase tracking-wider text-gray-700">
                  Total Products
                </p>
                <p className="mt-2 text-3xl font-extrabold text-gray-900">
                  {products.length}
                </p>
              </div>
              <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
                <p className="text-xs font-bold uppercase tracking-wider text-gray-700">
                  Active Listings
                </p>
                <p className="mt-2 text-3xl font-extrabold text-emerald-700">
                  {activeCount}
                </p>
              </div>
              <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
                <p className="text-xs font-bold uppercase tracking-wider text-gray-700">
                  Out of Stock
                </p>
                <p className="mt-2 text-3xl font-extrabold text-amber-700">
                  {outOfStockCount}
                </p>
              </div>
            </div>

            <div className="flex flex-wrap gap-3">
              <Link
                href="/seller/products"
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
              >
                Manage Products
              </Link>
              <Link
                href="/seller/orders"
                className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
              >
                View Received Orders
              </Link>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
