"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Paginated, Product } from "@/lib/types";
import { ProductCard } from "@/components/ProductCard";
import { Pagination } from "@/components/Pagination";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";

const PAGE_SIZE = 12;

export default function HomePage() {
  const [searchInput, setSearchInput] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<Paginated<Product> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.catalog.listProducts({
        q: activeQuery,
        page,
        pageSize: PAGE_SIZE,
      });
      setData(res);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load marketplace products"
      );
    } finally {
      setLoading(false);
    }
  }, [activeQuery, page]);

  useEffect(() => {
    void loadCatalog();
  }, [loadCatalog]);

  function handleSearch(e: FormEvent) {
    e.preventDefault();
    setPage(1);
    setActiveQuery(searchInput.trim());
  }

  function handleClear() {
    setSearchInput("");
    setActiveQuery("");
    setPage(1);
  }

  return (
    <main className="mx-auto max-w-6xl px-6 py-10">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-gray-900">
            Explore the Marketplace
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            Discover unique products directly from independent sellers.
          </p>
        </div>

        <form onSubmit={handleSearch} className="flex w-full gap-2 sm:w-auto">
          <input
            type="search"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search products..."
            className="w-full rounded-md border border-gray-300 bg-white px-3.5 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none sm:w-72"
          />
          <button
            type="submit"
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
          >
            Search
          </button>
          {activeQuery && (
            <button
              type="button"
              onClick={handleClear}
              className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Clear
            </button>
          )}
        </form>
      </div>

      {activeQuery && (
        <p className="mt-4 text-sm text-gray-600">
          Showing results for{" "}
          <span className="font-semibold text-gray-900">
            &ldquo;{activeQuery}&rdquo;
          </span>
        </p>
      )}

      <div className="mt-8">
        {loading ? (
          <LoadingState message="Loading products..." />
        ) : error ? (
          <ErrorState message={error} onRetry={loadCatalog} />
        ) : !data || data.items.length === 0 ? (
          <EmptyState
            title={
              activeQuery
                ? `No products matching "${activeQuery}"`
                : "No products listed yet"
            }
            description={
              activeQuery
                ? "Try a different keyword or clear the search filter."
                : "Check back soon or register as a seller to list the first item!"
            }
          />
        ) : (
          <>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {data.items.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
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
