"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiClientError } from "@/lib/api";
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
  const [cancellingId, setCancellingId] = useState<number | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

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

  async function handleCancel(orderId: number) {
    setCancellingId(orderId);
    setActionError(null);
    try {
      const res = await api.orders.cancel(orderId);
      setData((prev) =>
        prev
          ? {
              ...prev,
              items: prev.items.map((o) => (o.id === orderId ? res.order : o)),
            }
          : prev
      );
    } catch (err) {
      setActionError(
        err instanceof ApiClientError ? err.message : "Failed to cancel order"
      );
      await load();
    } finally {
      setCancellingId(null);
    }
  }

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

      {actionError && (
        <div
          role="alert"
          className="mt-4 rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-800"
        >
          {actionError}
        </div>
      )}

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
                        className="font-bold text-gray-900 hover:text-indigo-700"
                      >
                        Order #{o.id}
                      </Link>
                      <p className="text-xs font-medium text-gray-700">
                        {new Date(o.created_at).toLocaleString()} · from{" "}
                        <Link
                          href={`/sellers/${o.seller_id}`}
                          className="font-semibold text-indigo-700 hover:underline"
                        >
                          {o.store_name || `Seller #${o.seller_id}`}
                        </Link>
                        {(perCheckout.get(o.checkout_id) ?? 0) > 1 &&
                          " · part of a multi-seller purchase"}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <StatusBadge status={o.status} />
                      {o.status === "Pending" && (
                        <button
                          type="button"
                          disabled={cancellingId === o.id}
                          onClick={() => handleCancel(o.id)}
                          className="rounded-md border border-red-300 bg-red-50 px-2.5 py-1 text-xs font-semibold text-red-700 hover:bg-red-100 disabled:opacity-50"
                        >
                          {cancellingId === o.id ? "Cancelling..." : "Cancel"}
                        </button>
                      )}
                      <span className="font-extrabold text-gray-900">
                        {formatCents(o.total_cents)}
                      </span>
                    </div>
                  </div>
                  <p className="mt-3 line-clamp-1 text-sm font-medium text-gray-800">
                    {o.items
                      .map((i) => `${i.product_name} × ${i.quantity}`)
                      .join(", ")}
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
