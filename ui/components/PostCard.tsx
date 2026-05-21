import type { PublicPost } from "@/lib/api/types";
import { PublicContentCard } from "@/components/PublicContentCard";

export function PostCard({ post }: { post: PublicPost }) {
  return <PublicContentCard content={{ ...post, type: "article", typeLabel: "文章", href: `/posts/${post.slug}` }} />;
}
