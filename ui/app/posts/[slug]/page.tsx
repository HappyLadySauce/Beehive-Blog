import type { Metadata } from "next";
import { notFound } from "next/navigation";

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

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline: post.title,
    description: post.description,
    datePublished: post.publishedAt,
    keywords: post.tags.join(", ")
  };

  return (
    <main className="page">
      <article className="article">
        <p className="eyebrow">Article</p>
        <h1>{post.title}</h1>
        <div className="article-meta">
          <time dateTime={post.publishedAt}>{new Intl.DateTimeFormat("zh-CN").format(new Date(post.publishedAt))}</time>
          <span>{post.readingMinutes} 分钟阅读</span>
        </div>
        <div className="tag-row" aria-label="标签">
          {post.tags.map((tag) => (
            <span key={tag}>{tag}</span>
          ))}
        </div>
        <div className="article-body">
          {post.body.split("\n").map((paragraph) => (
            <p key={paragraph}>{paragraph}</p>
          ))}
        </div>
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      </article>
    </main>
  );
}
