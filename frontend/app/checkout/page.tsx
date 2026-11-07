"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, ApiClientError } from "@/lib/api";
import { formatCents } from "@/lib/money";
import type { Cart, CartItem } from "@/lib/types";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";

export default function CheckoutPage() {
  const router = useRouter();
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [placing, setPlacing] = useState(false);
  const [checkoutError, setCheckoutError] = useState<string | null>(null);

  const loadCart = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      setCart(await api.cart.get());
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Failed to load cart");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCart();
  }, [loadCart]);

  async function placeOrder() {
    setPlacing(true);
    setCheckoutError(null);
    try {
      await api.orders.checkout(); // mock payment: always succeeds
      router.push("/orders");
      router.refresh();
    } catch (err) {
      setCheckoutError(
        err instanceof ApiClientError
          ? err.message
          : "Checkout failed. Please try again."
      );
      // Stock/prices may have changed: refresh what the buyer sees.
      await loadCart();
    } finally {
      setPlacing(false);
    }
  }

  // Group cart lines by seller to preview how the purchase will be split.
  const groups = new Map<number, { store: string; items: CartItem[] }>();
  for (const item of cart?.items ?? []) {
    const g = groups.get(item.seller_id) ?? { store: item.store_name, items: [] };
    g.items.push(item);
    groups.set(item.seller_id, g);
  }

  return (
    <main className="mx-auto max-w-3xl px-6 py-10">
      <h1 className="text-2xl font-bold text-gray-900">Checkout</h1>
      <p className="mt-1 text-sm text-gray-600">
        Review your order. Payment is simulated, so no card is needed.
      </p>

      {checkoutError && (
        <div
          role="alert"
          className="mt-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          {checkoutError}{" "}
          <Link href="/cart" className="font-semibold underline">
            Review cart
          </Link>
        </div>
      )}

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Preparing checkout..." />
        ) : loadError ? (
          <ErrorState message={loadError} onRetry={loadCart} />
        ) : !cart || cart.items.length === 0 ? (
          <EmptyState
            title="Nothing to check out"
            description="Your cart is empty."
            action={
              <Link
                href="/"
                className="inline-block rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
              >
                Browse products
              </Link>
            }
          />
        ) : (
          <div className="space-y-6">
            {groups.size > 1 && (
              <p className="rounded-md border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-800">
                Your items come from {groups.size} sellers, so this purchase
                will create {groups.size} separate orders (one per seller).
              </p>
            )}

            {[...groups.entries()].map(([sellerId, g]) => {
              const subtotal = g.items.reduce((s, i) => s + i.line_total_cents, 0);
              return (
                <section
                  key={sellerId}
                  className="rounded-xl border border-gray-200 bg-white shadow-sm"
                >
                  <header className="flex items-center justify-between border-b border-gray-100 px-5 py-3">
                    <Link
                      href={`/sellers/${sellerId}`}
                      className="font-semibold text-gray-900 hover:text-indigo-600"
                    >
                      {g.store}
                    </Link>
                    <span className="text-sm text-gray-500">
                      Subtotal {formatCents(subtotal)}
                    </span>
                  </header>
                  <ul className="divide-y divide-gray-100">
                    {g.items.map((i) => (
                      <li
                        key={i.product_id}
                        className="flex items-center justify-between px-5 py-3 text-sm"
                      >
                        <span className="text-gray-800">
                          {i.name}{" "}
                          <span className="text-gray-500">
                            × {i.quantity} @ {formatCents(i.price_cents)}
                          </span>
                        </span>
                        <span className="font-medium text-gray-900">
                          {formatCents(i.line_total_cents)}
                        </span>
                      </li>
                    ))}
                  </ul>
                </section>
              );
            })}

            <div className="flex flex-col items-start justify-between gap-4 rounded-xl border border-gray-200 bg-white p-6 shadow-sm sm:flex-row sm:items-center">
              <p className="text-xl font-bold text-gray-900">
                Total: {formatCents(cart.total_cents)}
              </p>
              <button
                type="button"
                onClick={placeOrder}
                disabled={placing}
                className="rounded-md bg-indigo-600 px-6 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
              >
                {placing ? "Placing order..." : "Place order (mock payment)"}
              </button>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}

