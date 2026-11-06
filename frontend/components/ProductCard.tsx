import Link from "next/link";
import type { Product } from "@/lib/types";
import { formatCents } from "@/lib/money";

export function ProductCard({ product }: { product: Product }) {
  const outOfStock = product.stock === 0;

  return (
    <div className="group flex flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition hover:shadow-md">
      <Link
        href={`/products/${product.id}`}
        className="relative aspect-[4/3] w-full overflow-hidden bg-gray-100"
      >
        {product.image_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={product.image_url}
            alt={product.name}
            className="h-full w-full object-cover transition duration-300 group-hover:scale-105"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center text-xs font-medium text-gray-400">
            No image
          </div>
        )}
        {outOfStock && (
          <span className="absolute right-2 top-2 rounded-full bg-red-600 px-2.5 py-0.5 text-xs font-semibold text-white">
            Out of stock
          </span>
        )}
      </Link>

      <div className="flex flex-1 flex-col justify-between p-4">
        <div>
          <Link
            href={`/sellers/${product.seller_id}`}
            className="text-xs font-medium text-indigo-600 hover:underline"
          >
            {product.store_name || `Seller #${product.seller_id}`}
          </Link>
          <Link href={`/products/${product.id}`} className="mt-1 block">
            <h3 className="line-clamp-1 font-semibold text-gray-900 group-hover:text-indigo-600">
              {product.name}
            </h3>
          </Link>
          {product.description && (
            <p className="mt-1 line-clamp-2 text-xs text-gray-500">
              {product.description}
            </p>
          )}
        </div>

        <div className="mt-4 flex items-center justify-between border-t border-gray-100 pt-3">
          <span className="text-base font-bold text-gray-900">
            {formatCents(product.price_cents)}
          </span>
          <span className="text-xs text-gray-500">
            {outOfStock ? "0 left" : `${product.stock} in stock`}
          </span>
        </div>
      </div>
    </div>
  );
}

