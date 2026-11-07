import type {
  ApiErrorPayload,
  Cart,
  CheckoutResult,
  LoginInput,
  Order,
  Paginated,
  Product,
  ProductInput,
  RegisterInput,
  SellerProfile,
  Store,
  StoreInput,
  User,
} from "@/lib/types";

export class ApiClientError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(
    status: number,
    code: string,
    message: string,
    fields?: Record<string, string>
  ) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const data = (await res.json().catch(() => null)) as
    | T
    | ApiErrorPayload
    | null;

  if (!res.ok) {
    const errObj =
      data && typeof data === "object" && "error" in data
        ? data.error
        : { code: "unknown_error", message: `Request failed (${res.status})` };
    throw new ApiClientError(
      res.status,
      errObj.code,
      errObj.message,
      errObj.fields
    );
  }

  return data as T;
}

export const api = {
  auth: {
    register(input: RegisterInput): Promise<{ user: User }> {
      return request<{ user: User }>("/bff/auth/register", {
        method: "POST",
        body: JSON.stringify(input),
      });
    },
    login(input: LoginInput): Promise<{ user: User }> {
      return request<{ user: User }>("/bff/auth/login", {
        method: "POST",
        body: JSON.stringify(input),
      });
    },
    logout(): Promise<{ ok: boolean }> {
      return request<{ ok: boolean }>("/bff/auth/logout", {
        method: "POST",
      });
    },
    me(): Promise<{ user: User | null }> {
      return request<{ user: User | null }>("/bff/auth/me");
    },
  },

  seller: {
    getStore(): Promise<{ store: Store }> {
      return request<{ store: Store }>("/bff/api/seller/store");
    },
    createStore(input: StoreInput): Promise<{ store: Store }> {
      return request<{ store: Store }>("/bff/api/seller/store", {
        method: "POST",
        body: JSON.stringify(input),
      });
    },
    updateStore(input: StoreInput): Promise<{ store: Store }> {
      return request<{ store: Store }>("/bff/api/seller/store", {
        method: "PUT",
        body: JSON.stringify(input),
      });
    },
    listProducts(page = 1, pageSize = 20): Promise<Paginated<Product>> {
      return request<Paginated<Product>>(
        `/bff/api/seller/products?page=${page}&page_size=${pageSize}`
      );
    },
    createProduct(input: ProductInput): Promise<{ product: Product }> {
      return request<{ product: Product }>("/bff/api/seller/products", {
        method: "POST",
        body: JSON.stringify(input),
      });
    },
    updateProduct(
      id: number,
      input: ProductInput
    ): Promise<{ product: Product }> {
      return request<{ product: Product }>(`/bff/api/seller/products/${id}`, {
        method: "PUT",
        body: JSON.stringify(input),
      });
    },
    deleteProduct(id: number): Promise<void> {
      return request<void>(`/bff/api/seller/products/${id}`, {
        method: "DELETE",
      });
    },
  },

  catalog: {
    listProducts(params?: {
      q?: string;
      page?: number;
      pageSize?: number;
    }): Promise<Paginated<Product>> {
      const sp = new URLSearchParams();
      if (params?.q) sp.set("q", params.q);
      if (params?.page) sp.set("page", String(params.page));
      if (params?.pageSize) sp.set("page_size", String(params.pageSize));
      const qs = sp.toString();
      return request<Paginated<Product>>(
        `/bff/api/products${qs ? `?${qs}` : ""}`
      );
    },
    getProduct(id: number): Promise<{ product: Product }> {
      return request<{ product: Product }>(`/bff/api/products/${id}`);
    },
    getSellerProfile(sellerId: number): Promise<SellerProfile> {
      return request<SellerProfile>(`/bff/api/sellers/${sellerId}`);
    },
    listSellerProducts(
      sellerId: number,
      page = 1,
      pageSize = 12
    ): Promise<Paginated<Product>> {
      return request<Paginated<Product>>(
        `/bff/api/sellers/${sellerId}/products?page=${page}&page_size=${pageSize}`
      );
    },
  },

  cart: {
    get(): Promise<Cart> {
      return request<Cart>("/bff/api/cart/items");
    },
    addItem(productId: number, quantity: number): Promise<Cart> {
      return request<Cart>("/bff/api/cart/items", {
        method: "POST",
        body: JSON.stringify({ product_id: productId, quantity }),
      });
    },
    updateItem(productId: number, quantity: number): Promise<Cart> {
      return request<Cart>(`/bff/api/cart/items/${productId}`, {
        method: "PATCH",
        body: JSON.stringify({ quantity }),
      });
    },
    removeItem(productId: number): Promise<Cart> {
      return request<Cart>(`/bff/api/cart/items/${productId}`, {
        method: "DELETE",
      });
    },
  },

  orders: {
    // No arguments on purpose: the server reads the cart from the database,
    // so the browser can never tamper with prices or quantities.
    checkout(): Promise<CheckoutResult> {
      return request<CheckoutResult>("/bff/api/orders/checkout", {
        method: "POST",
        body: "{}",
      });
    },
    list(page = 1, pageSize = 10): Promise<Paginated<Order>> {
      return request<Paginated<Order>>(
        `/bff/api/orders?page=${page}&page_size=${pageSize}`
      );
    },
    get(id: number): Promise<{ order: Order }> {
      return request<{ order: Order }>(`/bff/api/orders/${id}`);
    },
  },
};
