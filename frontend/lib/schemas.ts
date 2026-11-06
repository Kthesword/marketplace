import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().trim().email("Please enter a valid email address"),
  password: z.string().min(1, "Password is required"),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

export const registerSchema = z.object({
  name: z.string().trim().min(1, "Name is required"),
  email: z.string().trim().email("Please enter a valid email address"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(72, "Password must not exceed 72 characters"),
  role: z.enum(["buyer", "seller"], {
    message: "Please select a role",
  }),
});

export type RegisterFormValues = z.infer<typeof registerSchema>;

const optionalHttpUrl = z
  .string()
  .trim()
  .refine(
    (v) => v === "" || /^https?:\/\/.+/i.test(v),
    "Must be a valid http:// or https:// URL"
  );

export const storeSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Store name is required")
    .max(100, "Store name must not exceed 100 characters"),
  description: z
    .string()
    .trim()
    .max(1000, "Description must not exceed 1000 characters"),
  logo_url: optionalHttpUrl,
  contact_email: z
    .string()
    .trim()
    .email("Please enter a valid contact email"),
});

export type StoreFormValues = z.infer<typeof storeSchema>;

export const productFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Product name is required")
    .max(140, "Product name must not exceed 140 characters"),
  description: z
    .string()
    .trim()
    .max(2000, "Description must not exceed 2000 characters"),
  price_dollars: z
    .string()
    .trim()
    .refine(
      (v) => /^\d+(\.\d{1,2})?$/.test(v) && Number(v) > 0,
      "Enter a valid price greater than 0 (e.g. 19.99)"
    ),
  stock: z
    .string()
    .trim()
    .refine(
      (v) => /^\d+$/.test(v) && Number(v) >= 0,
      "Stock must be a whole number 0 or greater"
    ),
  image_url: optionalHttpUrl,
  active: z.boolean(),
});

export type ProductFormValues = z.infer<typeof productFormSchema>;
