"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { productFormSchema, type ProductFormValues } from "@/lib/schemas";
import { api, ApiClientError } from "@/lib/api";
import { centsToDollars, dollarsToCents, formatCents } from "@/lib/money";
import type { Product } from "@/lib/types";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";

const emptyForm: ProductFormValues = {
  name: "",
  description: "",
  price_dollars: "",
  stock: "1",
  image_url: "",
  active: true,
};

export default function SellerProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [banner, setBanner] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ProductFormValues>({
    resolver: zodResolver(productFormSchema),
    defaultValues: emptyForm,
  });

  const loadProducts = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const page = await api.seller.listProducts(1, 50);
      setProducts(page.items);
    } catch (err) {
      setLoadError(
        err instanceof Error ? err.message : "Failed to load products"
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadProducts();
  }, [loadProducts]);

  function openCreate() {
    setEditingId(null);
    reset(emptyForm);
    setBanner(null);
    setShowForm(true);
  }

  function openEdit(p: Product) {
    setEditingId(p.id);
    reset({
      name: p.name,
      description: p.description,
      price_dollars: centsToDollars(p.price_cents),
      stock: String(p.stock),
      image_url: p.image_url,
      active: p.active,
    });
    setBanner(null);
    setShowForm(true);
  }

  async function onSubmit(values: ProductFormValues) {
    setBanner(null);
    const payload = {
      name: values.name,
      description: values.description,
      price_cents: dollarsToCents(Number(values.price_dollars)),
      stock: Number(values.stock),
      image_url: values.image_url,
      active: values.active,
    };

    try {
      if (editingId !== null) {
        await api.seller.updateProduct(editingId, payload);
        setBanner({ type: "success", text: "Product updated!" });
      } else {
        await api.seller.createProduct(payload);
        setBanner({ type: "success", text: "Product created!" });
      }
      setShowForm(false);
      setEditingId(null);
      reset(emptyForm);
      await loadProducts();
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.fields) {
          for (const [k, msg] of Object.entries(err.fields)) {
            const targetKey = k === "price_cents" ? "price_dollars" : k;
            if (targetKey in values) {
              setError(targetKey as keyof ProductFormValues, { message: msg });
            }
          }
        }
        setBanner({ type: "error", text: err.message });
      } else {
        setBanner({ type: "error", text: "Failed to save product." });
      }
    }
  }

  async function handleDelete(p: Product) {
    if (!window.confirm(`Delete "${p.name}"? This cannot be undone.`)) {
      return;
    }
    setBanner(null);
    try {
      await api.seller.deleteProduct(p.id);
      setBanner({ type: "success", text: `Deleted "${p.name}".` });
      await loadProducts();
    } catch (err) {
      setBanner({
        type: "error",
        text: err instanceof Error ? err.message : "Failed to delete product.",
      });
    }
  }

  return (
    <main className="mx-auto max-w-5xl px-6 py-10">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Your Products</h1>
          <p className="mt-1 text-sm text-gray-600">
            Add, edit, or deactivate items in your store catalog.
          </p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
        >
          + Add Product
        </button>
      </div>

      {banner && (
        <div
          role="alert"
          className={`mt-6 rounded-md border px-4 py-3 text-sm ${
            banner.type === "success"
              ? "border-emerald-200 bg-emerald-50 text-emerald-800"
              : "border-red-200 bg-red-50 text-red-700"
          }`}
        >
          {banner.text}{" "}
          {banner.text.includes("store") && (
            <Link href="/seller/store" className="font-semibold underline">
              Go to Store Setup
            </Link>
          )}
        </div>
      )}

      {showForm && (
        <section className="mt-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">
              {editingId !== null ? "Edit Product" : "New Product"}
            </h2>
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Cancel
            </button>
          </div>

          <form
            onSubmit={handleSubmit(onSubmit)}
            className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2"
            noValidate
          >
            <div className="sm:col-span-2">
              <label
                htmlFor="name"
                className="block text-sm font-medium text-gray-700"
              >
                Product Name *
              </label>
              <input
                id="name"
                type="text"
                {...register("name")}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                placeholder="Hand-thrown Ceramic Mug"
              />
              {errors.name && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.name.message}
                </p>
              )}
            </div>

            <div className="sm:col-span-2">
              <label
                htmlFor="description"
                className="block text-sm font-medium text-gray-700"
              >
                Description
              </label>
              <textarea
                id="description"
                rows={3}
                {...register("description")}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                placeholder="Describe materials, dimensions, care instructions..."
              />
              {errors.description && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.description.message}
                </p>
              )}
            </div>

            <div>
              <label
                htmlFor="price_dollars"
                className="block text-sm font-medium text-gray-700"
              >
                Price (USD) *
              </label>
              <input
                id="price_dollars"
                type="text"
                inputMode="decimal"
                {...register("price_dollars")}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                placeholder="24.99"
              />
              {errors.price_dollars && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.price_dollars.message}
                </p>
              )}
            </div>

            <div>
              <label
                htmlFor="stock"
                className="block text-sm font-medium text-gray-700"
              >
                Stock Quantity *
              </label>
              <input
                id="stock"
                type="number"
                min={0}
                {...register("stock")}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
              />
              {errors.stock && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.stock.message}
                </p>
              )}
            </div>

            <div className="sm:col-span-2">
              <label
                htmlFor="image_url"
                className="block text-sm font-medium text-gray-700"
              >
                Image URL (optional)
              </label>
              <input
                id="image_url"
                type="url"
                {...register("image_url")}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                placeholder="https://images.unsplash.com/..."
              />
              {errors.image_url && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.image_url.message}
                </p>
              )}
            </div>

            <div className="flex items-center gap-2 sm:col-span-2">
              <input
                id="active"
                type="checkbox"
                {...register("active")}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <label htmlFor="active" className="text-sm text-gray-700">
                Active (visible to buyers in the public catalog)
              </label>
            </div>

            <div className="flex gap-3 pt-2 sm:col-span-2">
              <button
                type="submit"
                disabled={isSubmitting}
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
              >
                {isSubmitting
                  ? "Saving..."
                  : editingId !== null
                    ? "Save Changes"
                    : "Create Product"}
              </button>
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </form>
        </section>
      )}

      <div className="mt-8">
        {loading ? (
          <LoadingState message="Loading your products..." />
        ) : loadError ? (
          <ErrorState message={loadError} onRetry={loadProducts} />
        ) : products.length === 0 ? (
          <EmptyState
            title="No products yet"
            description="Create your first product so buyers can discover your store."
            action={
              <button
                type="button"
                onClick={openCreate}
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
              >
                + Add Product
              </button>
            }
          />
        ) : (
          <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <tr>
                  <th className="px-6 py-3">Product</th>
                  <th className="px-6 py-3">Price</th>
                  <th className="px-6 py-3">Stock</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {products.map((p) => (
                  <tr key={p.id}>
                    <td className="px-6 py-4">
                      <div className="font-semibold text-gray-900">{p.name}</div>
                      {p.description && (
                        <div className="line-clamp-1 text-xs text-gray-500">
                          {p.description}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 font-medium text-gray-900">
                      {formatCents(p.price_cents)}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={
                          p.stock === 0
                            ? "font-semibold text-red-600"
                            : "text-gray-700"
                        }
                      >
                        {p.stock}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                          p.active
                            ? "bg-emerald-100 text-emerald-800"
                            : "bg-gray-100 text-gray-600"
                        }`}
                      >
                        {p.active ? "Active" : "Hidden"}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="inline-flex gap-3">
                        <button
                          type="button"
                          onClick={() => openEdit(p)}
                          className="font-medium text-indigo-600 hover:text-indigo-500"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(p)}
                          className="font-medium text-red-600 hover:text-red-500"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </main>
  );
}

