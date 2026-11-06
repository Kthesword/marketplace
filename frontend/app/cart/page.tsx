"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { formatCents } from "@/lib/money";
import type { Cart, CartItem } from "@/lib/types";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";

export default function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [busyProductId, setBusyProductId] = useState<number | null>(null);

  const loadCart = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const c = await api.cart.get();
      setCart(c);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load cart");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCart();
  }, [loadCart]);

  async function handleQuantityChange(item: CartItem, nextQty: number) {
    setActionError(null);
    setBusyProductId(item.product_id);
    try {
      const updated =
        nextQty <= 0
          ? await api.cart.removeItem(item.product_id)
          : await api.cart.updateItem(item.product_id, nextQty);
      setCart(updated);
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : "Could not update cart item"
      );
    } finally {
      setBusyProductId(null);
    }
  }

  async function handleRemove(productId: number) {
    setActionError(null);
    setBusyProductId(productId);
    try {
      const updated = await api.cart.removeItem(productId);
      setCart(updated);
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : "Could not remove cart item"
      );
    } finally {
      setBusyProductId(null);
    }
  }

  // Group count of unique sellers to show the buyer how their order will be split at checkout.
  const sellerCount = cart
    ? new Set(cart.items.map((i) => i.seller_id)).size
    : 0;

  return (
    <main className="mx-auto max-w-4xl px-6 py-10">
      <h1 className="text-2xl font-bold text-gray-900">Your Shopping Cart</h1>
      <p className="mt-1 text-sm text-gray-600">
        Review your items and adjust quantities before checking out.
      </p>

      {actionError && (
        <div
          role="alert"
          className="mt-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
          {actionError}
        </div>
      )}

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading your cart..." />
        ) : error ? (
          <ErrorState message={error} onRetry={loadCart} />
        ) : !cart || cart.items.length === 0 ? (
          <EmptyState
            title="Your cart is empty"
            description="Browse the marketplace to add items from independent sellers."
            action={
              <Link
                href="/"
                className="inline-block rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
              >
                Explore Products
              </Link>
            }
          />
        ) : (
          <div className="space-y-6">
            <div className="divide-y divide-gray-200 rounded-xl border border-gray-200 bg-white shadow-sm">
              {cart.items.map((item) => {
                const busy = busyProductId === item.product_id;
                return (
                  <div
                    key={item.product_id}
                    className="flex flex-col justify-between gap-4 p-5 sm:flex-row sm:items-center"
                  >
                    <div className="flex items-center gap-4">
                      <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg bg-gray-100">
                        {item.image_url ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={item.image_url}
                            alt={item.name}
                            className="h-full w-full object-cover"
                          />
                        ) : (
                          <div className="flex h-full w-full items-center justify-center text-xs text-gray-400">
                            No img
                          </div>
                        )}
                      </div>
                      <div>
                        <Link
                          href={`/sellers/${item.seller_id}`}
                          className="text-xs font-medium text-indigo-600 hover:underline"
                        >
                          {item.store_name}
                        </Link>
                        <Link
                          href={`/products/${item.product_id}`}
                          className="block font-semibold text-gray-900 hover:text-indigo-600"
                        >
                          {item.name}
                        </Link>
                        <p className="text-xs text-gray-500">
                          {formatCents(item.price_cents)} each · {item.stock} in
                          stock
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center justify-between gap-6 sm:justify-end">
                      <div className="flex items-center rounded-md border border-gray-300">
                        <button
                          type="button"
                          disabled={busy}
                          onClick={() =>
                            handleQuantityChange(item, item.quantity - 1)
                          }
                          className="px-3 py-1 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                          aria-label="Decrease quantity"
                        >
                          &minus;
                        </button>
                        <span className="px-3 py-1 text-sm font-medium text-gray-900">
                          {item.quantity}
                        </span>
                        <button
                          type="button"
                          disabled={busy || item.quantity >= item.stock}
                          onClick={() =>
                            handleQuantityChange(item, item.quantity + 1)
                          }
                          className="px-3 py-1 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                          aria-label="Increase quantity"
                        >
                          +
                        </button>
                      </div>

                      <div className="w-24 text-right font-bold text-gray-900">
                        {formatCents(item.line_total_cents)}
                      </div>

                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => handleRemove(item.product_id)}
                        className="text-xs font-medium text-red-600 hover:text-red-500 disabled:opacity-50"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex flex-col items-start justify-between gap-4 rounded-xl border border-gray-200 bg-white p-6 shadow-sm sm:flex-row sm:items-center">
              <div>
                <p className="text-sm text-gray-600">
                  Items from{" "}
                  <span className="font-semibold text-gray-900">
                    {sellerCount} {sellerCount === 1 ? "seller" : "sellers"}
                  </span>
                  {sellerCount > 1 &&
                    " (will create one order per seller at checkout)"}
                </p>
                <p className="mt-1 text-xl font-bold text-gray-900">
                  Total: {formatCents(cart.total_cents)}
                </p>
              </div>

              <Link
                href="/checkout"
                className="rounded-md bg-indigo-600 px-6 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
              >
                Proceed to Checkout
              </Link>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}

