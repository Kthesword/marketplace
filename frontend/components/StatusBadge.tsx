import type { OrderStatus } from "@/lib/types";

const styles: Record<OrderStatus, string> = {
  Pending: "bg-amber-100 text-amber-900 border border-amber-300",
  Shipped: "bg-blue-100 text-blue-900 border border-blue-300",
  Delivered: "bg-emerald-100 text-emerald-900 border border-emerald-300",
  Cancelled: "bg-gray-200 text-gray-900 border border-gray-400",
};

export function StatusBadge({ status }: { status: OrderStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold ${styles[status]}`}
    >
      {status}
    </span>
  );
}
