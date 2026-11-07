"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { formatCents } from "@/lib/money";
import type { Product } from "@/lib/types";
import { ErrorState, LoadingState } from "@/components/states";

export default function ProductDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const { user } = useAuth();
  const productId = Number(params.id);

  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [adding, setAdding] = useState(false);
  const [feedback, setFeedback] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const loadProduct = useCallback(async () => {
    if (!Number.isFinite(productId) || productId <= 0) {
      setError("Invalid product ID");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const { product: p } = await api.catalog.getProduct(productId);
      setProduct(p);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Product not found");
    } finally {
      setLoading(false);
    }
  }, [productId]);

  useEffect(() => {
    void loadProduct();
  }, [loadProduct]);

  async function handleAddToCart() {
    if (!user) {
      router.push("/login");
      return;
    }
    if (!product) return;

    setAdding(true);
    setFeedback(null);
    try {
      await api.cart.addItem(product.id, quantity);
      setFeedback({
        type: "success",
        text: `Added ${quantity} × "${product.name}" to your cart!`,
      });
    } catch (err) {
      setFeedback({
        type: "error",
        text: err instanceof Error ? err.message : "Could not add to cart",
      });
    } finally {
      setAdding(false);
    }
  }

  return (
    <main className="mx-auto max-w-4xl px-6 py-10">
      <Link
        href="/"
        className="text-sm font-medium text-indigo-600 hover:underline"
      >
        &larr; Back to catalog
      </Link>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading product details..." />
        ) : error || !product ? (
          <ErrorState
            message={error ?? "Product not found"}
            onRetry={loadProduct}
          />
        ) : (
          <div className="grid grid-cols-1 gap-8 rounded-xl border border-gray-200 bg-white p-6 shadow-sm md:grid-cols-2 md:p-8">
            <div className="aspect-square overflow-hidden rounded-lg bg-gray-100">
              {product.image_url ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={product.image_url}
                  alt={product.name}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center text-sm text-gray-400">
                  No image available
                </div>
              )}
            </div>

            <div className="flex flex-col justify-between">
              <div>
                <Link
                  href={`/sellers/${product.seller_id}`}
                  className="text-xs font-semibold uppercase tracking-wider text-indigo-600 hover:underline"
                >
                  Sold by {product.store_name || `Seller #${product.seller_id}`}
                </Link>
                <h1 className="mt-2 text-2xl font-bold text-gray-900">
                  {product.name}
                </h1>
                <p className="mt-3 text-2xl font-bold text-indigo-600">
                  {formatCents(product.price_cents)}
                </p>

                <div className="mt-4 border-t border-gray-100 pt-4">
                  <h2 className="text-xs font-bold uppercase tracking-wider text-gray-700">
                    Description
                  </h2>
                  <p className="mt-2 whitespace-pre-line text-sm font-normal text-gray-800">
                    {product.description || "No description provided."}
                  </p>
                </div>
              </div>

              <div className="mt-8 border-t border-gray-100 pt-6">
                <p className="text-sm font-semibold text-gray-800">
                  Availability:{" "}
                  {product.stock > 0 ? (
                    <span className="font-bold text-emerald-700">
                      {product.stock} in stock
                    </span>
                  ) : (
                    <span className="font-bold text-red-600">Out of stock</span>
                  )}
                </p>

                {feedback && (
                  <div
                    role="alert"
                    className={`mt-4 rounded-md border px-4 py-3 text-sm font-medium ${
                      feedback.type === "success"
                        ? "border-emerald-300 bg-emerald-50 text-emerald-900"
                        : "border-red-300 bg-red-50 text-red-800"
                    }`}
                  >
                    {feedback.text}{" "}
                    {feedback.type === "success" && (
                      <Link href="/cart" className="font-bold underline">
                        View Cart
                      </Link>
                    )}
                  </div>
                )}

                {user?.role === "seller" ? (
                  <p className="mt-4 text-xs font-medium text-gray-700">
                    You are signed in as a seller. Sign in with a buyer account to purchase items.
                  </p>
                ) : (
                  <div className="mt-4 flex items-center gap-3">
                    <label htmlFor="qty" className="sr-only">
                      Quantity
                    </label>
                    <input
                      id="qty"
                      type="number"
                      min={1}
                      max={Math.max(1, product.stock)}
                      disabled={product.stock === 0}
                      value={quantity}
                      onChange={(e) =>
                        setQuantity(
                          Math.max(
                            1,
                            Math.min(
                              product.stock,
                              Number(e.target.value) || 1
                            )
                          )
                        )
                      }
                      className="w-20 rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                    />
                    <button
                      type="button"
                      disabled={product.stock === 0 || adding}
                      onClick={handleAddToCart}
                      className="flex-1 rounded-md bg-indigo-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      {product.stock === 0
                        ? "Out of Stock"
                        : adding
                          ? "Adding..."
                          : "Add to Cart"}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}

