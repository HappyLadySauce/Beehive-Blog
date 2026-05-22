import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { PublicContentDetail } from "@/components/PublicContentDetail";
import { getPublicContent, listPublicContents } from "@/lib/api/content";

type PageProps = {
  params: Promise<{ slug: string }>;
};

export async function generateStaticParams() {
  const projects = await listPublicContents("project");
  return projects.map((project) => ({ slug: project.slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const project = await getPublicContent("project", slug);
  if (!project) return {};

  return {
    title: project.title,
    description: project.description,
    alternates: {
      canonical: `/projects/${project.slug}`
    },
    openGraph: {
      title: project.title,
      description: project.description,
      type: "article",
      publishedTime: project.publishedAt
    },
    twitter: {
      card: "summary_large_image",
      title: project.title,
      description: project.description
    }
  };
}

export default async function ProjectDetailPage({ params }: PageProps) {
  const { slug } = await params;
  const project = await getPublicContent("project", slug);
  if (!project) notFound();

  return await PublicContentDetail({ content: project, eyebrow: "Project" });
}
