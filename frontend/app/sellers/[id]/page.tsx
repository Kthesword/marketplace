"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import type { Paginated, Product, SellerProfile } from "@/lib/types";
import { ProductCard } from "@/components/ProductCard";
import { Pagination } from "@/components/Pagination";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";

const PAGE_SIZE = 12;

export default function SellerProfilePage() {
  const params = useParams<{ id: string }>();
  const sellerId = Number(params.id);

  const [profile, setProfile] = useState<SellerProfile | null>(null);
  const [products, setProducts] = useState<Paginated<Product> | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSellerPage = useCallback(async () => {
    if (!Number.isFinite(sellerId) || sellerId <= 0) {
      setError("Invalid seller ID");
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [prof, prods] = await Promise.all([
        api.catalog.getSellerProfile(sellerId),
        api.catalog.listSellerProducts(sellerId, page, PAGE_SIZE),
      ]);
      setProfile(prof);
      setProducts(prods);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Seller storefront not found"
      );
    } finally {
      setLoading(false);
    }
  }, [sellerId, page]);

  useEffect(() => {
    void loadSellerPage();
  }, [loadSellerPage]);

  return (
    <main className="mx-auto max-w-6xl px-6 py-10">
      <Link
        href="/"
        className="text-sm font-medium text-indigo-600 hover:underline"
      >
        &larr; Back to marketplace
      </Link>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading seller storefront..." />
        ) : error || !profile ? (
          <ErrorState
            message={error ?? "Seller not found"}
            onRetry={loadSellerPage}
          />
        ) : (
          <>
            <section className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm sm:p-8">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-4">
                  {profile.store.logo_url ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={profile.store.logo_url}
                      alt={profile.store.name}
                      className="h-16 w-16 rounded-full border border-gray-200 object-cover"
                    />
                  ) : (
                    <div className="flex h-16 w-16 items-center justify-center rounded-full bg-indigo-100 text-xl font-bold text-indigo-700">
                      {profile.store.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div>
                    <h1 className="text-2xl font-bold text-gray-900">
                      {profile.store.name}
                    </h1>
                    <p className="text-sm font-medium text-gray-700">
                      Managed by <span className="font-semibold text-gray-900">{profile.seller_name}</span> · Contact:{" "}
                      <a
                        href={`mailto:${profile.store.contact_email}`}
                        className="font-semibold text-indigo-700 hover:underline"
                      >
                        {profile.store.contact_email}
                      </a>
                    </p>
                  </div>
                </div>
              </div>
              {profile.store.description && (
                <p className="mt-4 text-sm text-gray-700">
                  {profile.store.description}
                </p>
              )}
            </section>

            <section className="mt-8">
              <h2 className="text-lg font-bold text-gray-900">
                Products from {profile.store.name}
              </h2>

              <div className="mt-4">
                {!products || products.items.length === 0 ? (
                  <EmptyState
                    title="No active products"
                    description="This seller has not published any active products yet."
                  />
                ) : (
                  <>
                    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                      {products.items.map((p) => (
                        <ProductCard key={p.id} product={p} />
                      ))}
                    </div>
                    <Pagination
                      page={products.page}
                      pageSize={products.page_size}
                      total={products.total}
                      onPageChange={setPage}
                    />
                  </>
                )}
              </div>
            </section>
          </>
        )}
      </div>
    </main>
  );
}

