"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { registerSchema, type RegisterFormValues } from "@/lib/schemas";
import { useAuth } from "@/lib/auth-context";
import { ApiClientError } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const { register: registerUser } = useAuth();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setError,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      name: "",
      email: "",
      password: "",
      role: "buyer",
    },
  });

  const selectedRole = watch("role");

  async function onSubmit(values: RegisterFormValues) {
    setServerError(null);
    try {
      const user = await registerUser(values);
      router.push(user.role === "seller" ? "/seller/dashboard" : "/");
      router.refresh();
    } catch (err) {
      if (err instanceof ApiClientError) {
        if (err.fields) {
          for (const [field, msg] of Object.entries(err.fields)) {
            if (field in values) {
              setError(field as keyof RegisterFormValues, { message: msg });
            }
          }
        }
        setServerError(err.message);
      } else {
        setServerError("Something went wrong. Please try again.");
      }
    }
  }

  return (
    <main className="mx-auto max-w-md px-6 py-12">
      <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
        <h1 className="text-2xl font-bold text-gray-900">Create an account</h1>
        <p className="mt-1 text-sm text-gray-600">
          Choose whether you want to buy products or open a seller store.
        </p>

        {serverError && (
          <div
            role="alert"
            className="mt-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
          >
            {serverError}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4" noValidate>
          {/* Role selector */}
          <fieldset>
            <legend className="block text-sm font-medium text-gray-700">
              Account Role
            </legend>
            <div className="mt-2 grid grid-cols-2 gap-3">
              <label
                className={`flex cursor-pointer flex-col rounded-lg border p-3 text-sm ${
                  selectedRole === "buyer"
                    ? "border-indigo-600 bg-indigo-50/50 text-indigo-900"
                    : "border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
                }`}
              >
                <input
                  type="radio"
                  value="buyer"
                  {...register("role")}
                  className="sr-only"
                />
                <span className="font-semibold text-gray-900">Buyer</span>
                <span className="mt-0.5 text-xs font-medium text-gray-700">
                  Browse &amp; purchase items
                </span>
              </label>

              <label
                className={`flex cursor-pointer flex-col rounded-lg border p-3 text-sm ${
                  selectedRole === "seller"
                    ? "border-indigo-600 bg-indigo-50/50 text-indigo-900"
                    : "border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
                }`}
              >
                <input
                  type="radio"
                  value="seller"
                  {...register("role")}
                  className="sr-only"
                />
                <span className="font-semibold text-gray-900">Seller</span>
                <span className="mt-0.5 text-xs font-medium text-gray-700">
                  Open a store &amp; sell
                </span>
              </label>
            </div>
            {errors.role && (
              <p className="mt-1 text-xs text-red-600">{errors.role.message}</p>
            )}
          </fieldset>

          <div>
            <label
              htmlFor="name"
              className="block text-sm font-semibold text-gray-900"
            >
              Full Name
            </label>
            <input
              id="name"
              type="text"
              autoComplete="name"
              {...register("name")}
              className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-indigo-500 focus:outline-none"
              placeholder="Alice Smith"
            />
            {errors.name && (
              <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>
            )}
          </div>

          <div>
            <label
              htmlFor="email"
              className="block text-sm font-semibold text-gray-900"
            >
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              {...register("email")}
              className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-indigo-500 focus:outline-none"
              placeholder="you@example.com"
            />
            {errors.email && (
              <p className="mt-1 text-xs text-red-600">{errors.email.message}</p>
            )}
          </div>

          <div>
            <label
              htmlFor="password"
              className="block text-sm font-semibold text-gray-900"
            >
              Password (min 8 characters)
            </label>
            <input
              id="password"
              type="password"
              autoComplete="new-password"
              {...register("password")}
              className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-indigo-500 focus:outline-none"
              placeholder="••••••••"
            />
            {errors.password && (
              <p className="mt-1 text-xs text-red-600">
                {errors.password.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
          >
            {isSubmitting ? "Creating account..." : "Create account"}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-600">
          Already have an account?{" "}
          <Link
            href="/login"
            className="font-semibold text-indigo-600 hover:text-indigo-500"
          >
            Log in
          </Link>
        </p>
      </div>
    </main>
  );
}

