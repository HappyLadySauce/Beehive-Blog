import type { Metadata } from "next";
import type { ReactNode } from "react";

import { StudioLayout } from "@/components/studio/StudioLayout";

export const metadata: Metadata = {
  title: "Studio",
  robots: {
    index: false,
    follow: false
  }
};

export default function StudioRouteLayout({ children }: { children: ReactNode }) {
  return <StudioLayout>{children}</StudioLayout>;
}
