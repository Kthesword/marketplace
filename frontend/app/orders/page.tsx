"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { formatCents } from "@/lib/money";
import type { Order, Paginated } from "@/lib/types";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";
import { Pagination } from "@/components/Pagination";
import { StatusBadge } from "@/components/StatusBadge";

export default function OrdersPage() {
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Paginated<Order> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(await api.orders.list(page, 10));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load orders");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    void load();
  }, [load]);

  // Count orders per checkout_id so we can label purchases that were split by seller.
  const perCheckout = new Map<string, number>();
  for (const o of data?.items ?? []) {
    perCheckout.set(o.checkout_id, (perCheckout.get(o.checkout_id) ?? 0) + 1);
  }

  return (
    <main className="mx-auto max-w-4xl px-6 py-10">
      <h1 className="text-2xl font-bold text-gray-900">My Orders</h1>
      <p className="mt-1 text-sm text-gray-600">
        Track the status of everything you have purchased.
      </p>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading your orders..." />
        ) : error ? (
          <ErrorState message={error} onRetry={load} />
        ) : !data || data.items.length === 0 ? (
          <EmptyState
            title="No orders yet"
            description="When you check out, your orders will appear here."
            action={
              <Link
                href="/"
                className="inline-block rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
              >
                Start shopping
              </Link>
            }
          />
        ) : (
          <>
            <ul className="space-y-4">
              {data.items.map((o) => (
                <li
                  key={o.id}
                  className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm"
                >
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <Link
                        href={`/orders/${o.id}`}
                        className="font-semibold text-gray-900 hover:text-indigo-600"
                      >
                        Order #{o.id}
                      </Link>
                      <p className="text-xs text-gray-500">
                        {new Date(o.created_at).toLocaleString()} · from{" "}
                        <Link
                          href={`/sellers/${o.seller_id}`}
                          className="text-indigo-600 hover:underline"
                        >
                          {o.store_name || `Seller #${o.seller_id}`}
                        </Link>
                        {(perCheckout.get(o.checkout_id) ?? 0) > 1 &&
                          " · part of a multi-seller purchase"}
                      </p>
                    </div>
                    <div className="flex items-center gap-4">
                      <StatusBadge status={o.status} />
                      <span className="font-bold text-gray-900">
                        {formatCents(o.total_cents)}
                      </span>
                    </div>
                  </div>
                  <p className="mt-3 line-clamp-1 text-sm text-gray-600">
                    {o.items.map((i) => `${i.product_name} × ${i.quantity}`).join(", ")}
                  </p>
                </li>
              ))}
            </ul>
            <Pagination
              page={data.page}
              pageSize={data.page_size}
              total={data.total}
              onPageChange={setPage}
            />
          </>
        )}
      </div>
    </main>
  );
}

