export const accessCookieName = "beehive.access";
export const refreshCookieName = "beehive.refresh";

export function secureCookieEnabled() {
  return process.env.NODE_ENV === "production";
}
