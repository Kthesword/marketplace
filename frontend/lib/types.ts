// Shared TypeScript types mirroring the Go backend JSON DTOs.

export type Role = "seller" | "buyer";

export interface User {
  id: number;
  email: string;
  name: string;
  role: Role;
  created_at: string;
}

export interface ApiErrorPayload {
  error: {
    code: string;
    message: string;
    fields?: Record<string, string>;
  };
}

export interface Paginated<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
}

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
  role: Role;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface Store {
  id: number;
  seller_id: number;
  name: string;
  description: string;
  logo_url: string;
  contact_email: string;
  created_at: string;
  updated_at: string;
}

export interface StoreInput {
  name: string;
  description: string;
  logo_url: string;
  contact_email: string;
}

export interface SellerProfile {
  seller_id: number;
  seller_name: string;
  joined_at: string;
  store: Store;
}

export interface Product {
  id: number;
  seller_id: number;
  store_name: string;
  name: string;
  description: string;
  price_cents: number;
  stock: number;
  image_url: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductInput {
  name: string;
  description: string;
  price_cents: number;
  stock: number;
  image_url: string;
  active: boolean;
}

export interface CartItem {
  product_id: number;
  seller_id: number;
  store_name: string;
  name: string;
  image_url: string;
  price_cents: number;
  stock: number;
  quantity: number;
  line_total_cents: number;
}

export interface Cart {
  items: CartItem[];
  total_cents: number;
}

export type OrderStatus = "Pending" | "Shipped" | "Delivered" | "Cancelled";

export interface OrderItem {
  id: number;
  product_id: number | null; // null if the product was deleted after purchase
  product_name: string; // snapshot at purchase time
  unit_price_cents: number; // snapshot at purchase time
  quantity: number;
  line_total_cents: number;
}

export interface Order {
  id: number;
  checkout_id: string;
  buyer_id: number;
  buyer_name?: string;
  seller_id: number;
  store_name: string;
  status: OrderStatus;
  total_cents: number;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface CheckoutResult {
  checkout_id: string;
  orders: Order[];
}
