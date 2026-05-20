export function isStudioContentEditorPath(pathname: string | null | undefined) {
  if (!pathname) return false;
  return pathname === "/studio/content/new" || /^\/studio\/content\/\d+\/edit$/.test(pathname);
}
