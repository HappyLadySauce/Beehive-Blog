import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { PublicContentDetail } from "@/components/PublicContentDetail";
import { getPublicContent, listPublicContents } from "@/lib/api/content";

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateStaticParams() {
  const notes = await listPublicContents("note");
  return notes.map((note) => ({ slug: note.slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const note = await getPublicContent("note", slug);
  if (!note) return {};

  return {
    title: note.title,
    description: note.description,
    alternates: {
      canonical: `/notes/${note.slug}`
    },
    openGraph: {
      title: note.title,
      description: note.description,
      type: "article",
      publishedTime: note.publishedAt
    },
    twitter: {
      card: "summary_large_image",
      title: note.title,
      description: note.description
    }
  };
}

export default async function NoteDetailPage({ params }: PageProps) {
  const { slug } = await params;
  const note = await getPublicContent("note", slug);
  if (!note) notFound();

  return await PublicContentDetail({ content: note, eyebrow: "Note" });
}
