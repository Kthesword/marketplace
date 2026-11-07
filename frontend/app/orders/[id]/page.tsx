"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import { formatCents } from "@/lib/money";
import type { Order } from "@/lib/types";
import { ErrorState, LoadingState } from "@/components/states";
import { StatusBadge } from "@/components/StatusBadge";

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>();
  const orderId = Number(params.id);

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!Number.isFinite(orderId) || orderId <= 0) {
      setError("Invalid order ID");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      setOrder((await api.orders.get(orderId)).order);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Order not found");
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <main className="mx-auto max-w-3xl px-6 py-10">
      <Link href="/orders" className="text-sm font-medium text-indigo-600 hover:underline">
        &larr; Back to my orders
      </Link>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading order..." />
        ) : error || !order ? (
          <ErrorState message={error ?? "Order not found"} onRetry={load} />
        ) : (
          <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h1 className="text-xl font-bold text-gray-900">Order #{order.id}</h1>
                <p className="mt-1 text-sm text-gray-500">
                  Placed {new Date(order.created_at).toLocaleString()} · sold by{" "}
                  <Link
                    href={`/sellers/${order.seller_id}`}
                    className="text-indigo-600 hover:underline"
                  >
                    {order.store_name || `Seller #${order.seller_id}`}
                  </Link>
                </p>
              </div>
              <StatusBadge status={order.status} />
            </div>

            <table className="mt-6 min-w-full divide-y divide-gray-200 text-sm">
              <thead className="text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <tr>
                  <th className="py-2">Item</th>
                  <th className="py-2">Price</th>
                  <th className="py-2">Qty</th>
                  <th className="py-2 text-right">Total</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {order.items.map((i) => (
                  <tr key={i.id}>
                    <td className="py-3 font-medium text-gray-900">
                      {i.product_id ? (
                        <Link href={`/products/${i.product_id}`} className="hover:text-indigo-600">
                          {i.product_name}
                        </Link>
                      ) : (
                        <span title="This product was removed by the seller">
                          {i.product_name}
                        </span>
                      )}
                    </td>
                    <td className="py-3 text-gray-700">{formatCents(i.unit_price_cents)}</td>
                    <td className="py-3 text-gray-700">{i.quantity}</td>
                    <td className="py-3 text-right font-medium text-gray-900">
                      {formatCents(i.line_total_cents)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            <p className="mt-4 border-t border-gray-100 pt-4 text-right text-lg font-bold text-gray-900">
              Total: {formatCents(order.total_cents)}
            </p>
            <p className="mt-2 text-xs text-gray-400">
              Item names and prices are the ones you paid at purchase time.
            </p>
          </div>
        )}
      </div>
    </main>
  );
}

