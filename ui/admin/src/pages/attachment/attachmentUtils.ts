import type { Attachment, AttachmentFamilyResponse } from '../../api/attachment';

export function formatSize(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

export function collectFamilyMembers(f: AttachmentFamilyResponse): Attachment[] {
  return [f.root, ...f.children];
}

export function findMember(f: AttachmentFamilyResponse | null, id: number): Attachment | undefined {
  if (!f) return undefined;
  return collectFamilyMembers(f).find((a) => a.id === id);
}
