"use client";

import { useState } from "react";

const DEFAULT_PLACEHOLDER = "/placeholder.svg";

interface ProductImageProps {
  src?: string | null;
  alt: string;
  className?: string;
}

export function ProductImage({ src, alt, className = "h-full w-full object-cover" }: ProductImageProps) {
  const [imgSrc, setImgSrc] = useState<string>(src && src.trim() !== "" ? src : DEFAULT_PLACEHOLDER);

  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={imgSrc}
      alt={alt}
      className={className}
      loading="lazy"
      onError={() => {
        if (imgSrc !== DEFAULT_PLACEHOLDER) {
          setImgSrc(DEFAULT_PLACEHOLDER);
        }
      }}
    />
  );
}
