import type { NextConfig } from "next";

const goApiBaseUrl = process.env.BEEHIVE_API_BASE_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: `${goApiBaseUrl}/api/v1/:path*`
      }
    ];
  }
};

export default nextConfig;
