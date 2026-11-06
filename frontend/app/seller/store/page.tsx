"use client";

import { useCallback, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { storeSchema, type StoreFormValues } from "@/lib/schemas";
import { api, ApiClientError } from "@/lib/api";
import type { Store } from "@/lib/types";
import { ErrorState, LoadingState } from "@/components/states";

export default function SellerStorePage() {
  const [store, setStore] = useState<Store | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
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
  } = useForm<StoreFormValues>({
    resolver: zodResolver(storeSchema),
    defaultValues: {
      name: "",
      description: "",
      logo_url: "",
      contact_email: "",
    },
  });

  const loadStore = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const { store: existing } = await api.seller.getStore();
      setStore(existing);
      reset({
        name: existing.name,
        description: existing.description,
        logo_url: existing.logo_url,
        contact_email: existing.contact_email,
      });
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 404) {
        // 404 simply means the seller hasn't set up their store yet.
        setStore(null);
      } else {
        setLoadError(
          err instanceof Error ? err.message : "Failed to load store"
        );
      }
    } finally {
      setLoading(false);
    }
  }, [reset]);

  useEffect(() => {
    void loadStore();
  }, [loadStore]);

  async function onSubmit(values: StoreFormValues) {
    setBanner(null);
    try {
      const res = store
        ? await api.seller.updateStore(values)
        : await api.seller.createStore(values);
      setStore(res.store);
      reset({
        name: res.store.name,
        description: res.store.description,
        logo_url: res.store.logo_url,
        contact_email: res.store.contact_email,
      });
      setBanner({
        type: "success",
        text: store
          ? "Store details updated!"
          : "Store created! You can now add products.",
      });
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.fields) {
          for (const [k, msg] of Object.entries(err.fields)) {
            if (k in values) {
              setError(k as keyof StoreFormValues, { message: msg });
            }
          }
        }
        setBanner({ type: "error", text: err.message });
      } else {
        setBanner({ type: "error", text: "Failed to save store." });
      }
    }
  }

  return (
    <main className="mx-auto max-w-2xl px-6 py-10">
      <h1 className="text-2xl font-bold text-gray-900">
        {store ? "Manage Your Store" : "Set Up Your Store"}
      </h1>
      <p className="mt-1 text-sm text-gray-600">
        {store
          ? "Update how your storefront appears to buyers across the marketplace."
          : "Create your public storefront before listing products."}
      </p>

      <div className="mt-6">
        {loading ? (
          <LoadingState message="Loading store settings..." />
        ) : loadError ? (
          <ErrorState message={loadError} onRetry={loadStore} />
        ) : (
          <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
            {banner && (
              <div
                role="alert"
                className={`mb-5 rounded-md border px-4 py-3 text-sm ${
                  banner.type === "success"
                    ? "border-emerald-200 bg-emerald-50 text-emerald-800"
                    : "border-red-200 bg-red-50 text-red-700"
                }`}
              >
                {banner.text}
              </div>
            )}

            <form
              onSubmit={handleSubmit(onSubmit)}
              className="space-y-4"
              noValidate
            >
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-gray-700"
                >
                  Store Name *
                </label>
                <input
                  id="name"
                  type="text"
                  {...register("name")}
                  className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                  placeholder="Alice Artisan Crafts"
                />
                {errors.name && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.name.message}
                  </p>
                )}
              </div>

              <div>
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
                  placeholder="Tell buyers about your products and craft..."
                />
                {errors.description && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.description.message}
                  </p>
                )}
              </div>

              <div>
                <label
                  htmlFor="logo_url"
                  className="block text-sm font-medium text-gray-700"
                >
                  Logo URL (optional)
                </label>
                <input
                  id="logo_url"
                  type="url"
                  {...register("logo_url")}
                  className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                  placeholder="https://example.com/logo.png"
                />
                {errors.logo_url && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.logo_url.message}
                  </p>
                )}
              </div>

              <div>
                <label
                  htmlFor="contact_email"
                  className="block text-sm font-medium text-gray-700"
                >
                  Public Contact Email *
                </label>
                <input
                  id="contact_email"
                  type="email"
                  {...register("contact_email")}
                  className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none"
                  placeholder="support@yourstore.com"
                />
                {errors.contact_email && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.contact_email.message}
                  </p>
                )}
              </div>

              <div className="pt-2">
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="rounded-md bg-indigo-600 px-5 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
                >
                  {isSubmitting
                    ? "Saving..."
                    : store
                      ? "Save Changes"
                      : "Create Store"}
                </button>
              </div>
            </form>
          </div>
        )}
      </div>
    </main>
  );
}

