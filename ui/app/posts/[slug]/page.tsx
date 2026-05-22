import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { PublicContentDetail } from "@/components/PublicContentDetail";
import { getPublicPost, listPublicPosts } from "@/lib/api/content";

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateStaticParams() {
  const posts = await listPublicPosts();
  return posts.map((post) => ({ slug: post.slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPublicPost(slug);
  if (!post) return {};

  return {
    title: post.title,
    description: post.description,
    alternates: {
      canonical: `/posts/${post.slug}`
    },
    openGraph: {
      title: post.title,
      description: post.description,
      type: "article",
      publishedTime: post.publishedAt
    },
    twitter: {
      card: "summary_large_image",
      title: post.title,
      description: post.description
    }
  };
}

export default async function PostDetailPage({ params }: PageProps) {
  const { slug } = await params;
  const post = await getPublicPost(slug);
  if (!post) notFound();

  return await PublicContentDetail({
    content: { ...post, type: "article", typeLabel: "文章", href: `/posts/${post.slug}` },
    eyebrow: "Article"
  });
}
