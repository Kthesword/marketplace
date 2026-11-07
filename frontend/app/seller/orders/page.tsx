"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiClientError } from "@/lib/api";
import { formatCents } from "@/lib/money";
import type { Order, OrderStatus, Paginated } from "@/lib/types";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";
import { Pagination } from "@/components/Pagination";
import { StatusBadge } from "@/components/StatusBadge";

const FILTER_OPTIONS: { label: string; value: OrderStatus | "All" }[] = [
  { label: "All", value: "All" },
  { label: "Pending", value: "Pending" },
  { label: "Shipped", value: "Shipped" },
  { label: "Delivered", value: "Delivered" },
  { label: "Cancelled", value: "Cancelled" },
];

export default function SellerOrdersPage() {
  const [statusFilter, setStatusFilter] = useState<OrderStatus | "All">("All");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Paginated<Order> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingId, setUpdatingId] = useState<number | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(
        await api.seller.listOrders({
          status: statusFilter === "All" ? undefined : statusFilter,
          page,
          pageSize: 10,
        })
      );
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load seller orders"
      );
    } finally {
      setLoading(false);
    }
  }, [statusFilter, page]);

  useEffect(() => {
    void load();
  }, [load]);

  function handleFilterChange(next: OrderStatus | "All") {
    setStatusFilter(next);
    setPage(1);
  }

  async function advanceStatus(orderId: number, nextStatus: OrderStatus) {
    setUpdatingId(orderId);
    setActionError(null);
    try {
      const res = await api.seller.updateOrderStatus(orderId, nextStatus);
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
        err instanceof ApiClientError
          ? err.message
          : "Failed to update order status"
      );
      await load();
    } finally {
      setUpdatingId(null);
    }
  }

  return (
    <main className="mx-auto max-w-5xl px-6 py-10">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Received Orders</h1>
          <p className="mt-1 text-sm text-gray-600">
            Fulfill customer orders for your store: Pending &rarr; Shipped &rarr;
            Delivered.
          </p>
        </div>

        {/* Status filter pills */}
        <div className="flex flex-wrap gap-1.5 rounded-lg border border-gray-200 bg-white p-1">
          {FILTER_OPTIONS.map((opt) => {
            const active = statusFilter === opt.value;
            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => handleFilterChange(opt.value)}
                className={`rounded-md px-3 py-1.5 text-xs font-semibold transition ${
                  active
                    ? "bg-indigo-600 text-white"
                    : "text-gray-600 hover:bg-gray-100"
                }`}
              >
                {opt.label}
              </button>
            );
          })}
        </div>
      </div>

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
          <LoadingState message="Loading received orders..." />
        ) : error ? (
          <ErrorState message={error} onRetry={load} />
        ) : !data || data.items.length === 0 ? (
          <EmptyState
            title={
              statusFilter === "All"
                ? "No orders received yet"
                : `No ${statusFilter.toLowerCase()} orders`
            }
            description="Orders placed by buyers for your store's products will appear here."
          />
        ) : (
          <>
            <div className="space-y-4">
              {data.items.map((o) => {
                const busy = updatingId === o.id;
                return (
                  <div
                    key={o.id}
                    className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h2 className="text-lg font-bold text-gray-900">
                          Order #{o.id}
                        </h2>
                        <p className="mt-0.5 text-xs font-medium text-gray-700">
                          Buyer:{" "}
                          <span className="font-bold text-gray-900">
                            {o.buyer_name || `Buyer #${o.buyer_id}`}
                          </span>{" "}
                          · Placed {new Date(o.created_at).toLocaleString()}
                        </p>
                      </div>

                      <div className="flex flex-wrap items-center gap-3">
                        <StatusBadge status={o.status} />

                        {o.status === "Pending" && (
                          <button
                            type="button"
                            disabled={busy}
                            onClick={() => advanceStatus(o.id, "Shipped")}
                            className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-500 disabled:opacity-50"
                          >
                            {busy ? "Updating..." : "Mark as Shipped"}
                          </button>
                        )}

                        {o.status === "Shipped" && (
                          <button
                            type="button"
                            disabled={busy}
                            onClick={() => advanceStatus(o.id, "Delivered")}
                            className="rounded-md bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-emerald-500 disabled:opacity-50"
                          >
                            {busy ? "Updating..." : "Mark as Delivered"}
                          </button>
                        )}
                      </div>
                    </div>

                    <table className="mt-4 min-w-full divide-y divide-gray-200 text-sm">
                      <thead className="text-left text-xs font-bold uppercase tracking-wider text-gray-700">
                        <tr>
                          <th className="py-2">Product (Snapshot)</th>
                          <th className="py-2">Unit Price</th>
                          <th className="py-2">Qty</th>
                          <th className="py-2 text-right">Line Total</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {o.items.map((it) => (
                          <tr key={it.id}>
                            <td className="py-2.5 font-bold text-gray-900">
                              {it.product_name}
                              {!it.product_id && (
                                <span className="ml-2 text-xs font-semibold text-gray-600">
                                  (deleted product)
                                </span>
                              )}
                            </td>
                            <td className="py-2.5 font-medium text-gray-800">
                              {formatCents(it.unit_price_cents)}
                            </td>
                            <td className="py-2.5 font-medium text-gray-800">
                              {it.quantity}
                            </td>
                            <td className="py-2.5 text-right font-bold text-gray-900">
                              {formatCents(it.line_total_cents)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>

                    <div className="mt-3 flex justify-end border-t border-gray-100 pt-3">
                      <span className="text-base font-bold text-gray-900">
                        Order Total: {formatCents(o.total_cents)}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>

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

