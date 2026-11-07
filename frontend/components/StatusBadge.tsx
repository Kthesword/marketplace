import type { OrderStatus } from "@/lib/types";

const styles: Record<OrderStatus, string> = {
  Pending: "bg-amber-100 text-amber-800",
  Shipped: "bg-blue-100 text-blue-800",
  Delivered: "bg-emerald-100 text-emerald-800",
  Cancelled: "bg-gray-200 text-gray-600",
};

export function StatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold ${styles[status]}`}
    >
      {status}
    </span>
  );
}

