import type { PublicContent } from "@/lib/api/types";
import type React from "react";

const gradients = [
  "linear-gradient(135deg, #44c08a 0%, #259f75 100%)",
  "linear-gradient(135deg, #2b69f2 0%, #1e40af 100%)",
  "linear-gradient(135deg, #ff7a18 0%, #f43f5e 100%)",
  "linear-gradient(135deg, #243b55 0%, #141e30 100%)",
  "linear-gradient(135deg, #f6b73c 0%, #d8891b 100%)"
];

export function coverStyle(content: Pick<PublicContent, "coverUrl" | "slug">): React.CSSProperties {
  if (content.coverUrl) {
    return {
      backgroundImage: `linear-gradient(180deg, rgba(0,0,0,0.08), rgba(0,0,0,0.12)), url(${content.coverUrl})`
    };
  }
  return { background: gradients[hashSlug(content.slug) % gradients.length] };
}

function hashSlug(slug: string) {
  let hash = 0;
  for (const char of slug) hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  return hash;
}
