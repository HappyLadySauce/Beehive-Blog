import { notFound } from "next/navigation";

import { StudioContentEditorPage } from "@/components/studio/StudioContentEditorPage";

type StudioContentEditRouteProps = {
  params: Promise<{ id: string }>;
};

export default async function StudioContentEditRoute({ params }: StudioContentEditRouteProps) {
  const { id } = await params;
  const contentId = Number(id);
  if (!Number.isSafeInteger(contentId) || contentId <= 0) {
    notFound();
  }
  return <StudioContentEditorPage contentId={contentId} mode="edit" />;
}
