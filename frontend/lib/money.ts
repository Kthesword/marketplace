const usdFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
});

// Formats integer cents (e.g. 2499) as "$24.99".
export function formatCents(cents: number): string {
  return usdFormatter.format(cents / 100);
}

// Converts a dollar string/number (e.g. "19.99") to integer cents (1999).
// Math.round avoids IEEE-754 float drift like 19.99 * 100 === 1998.9999999999998.
export function dollarsToCents(dollars: number): number {
  return Math.round(dollars * 100);
}

// Converts integer cents (2499) to a 2-decimal dollar string ("24.99") for form inputs.
export function centsToDollars(cents: number): string {
  return (cents / 100).toFixed(2);
}

