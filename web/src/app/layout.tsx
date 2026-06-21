import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";
import { StarField } from "@/components/background/StarField";
import { MountainSilhouette } from "@/components/background/MountainSilhouette";
import { Aurora } from "@/components/background/Aurora";

const inter = Inter({
  subsets: ["latin"],
  weight: ["300", "400", "500"],
  variable: "--font-inter",
  display: "swap",
});

export const metadata: Metadata = {
  title: "GitWhy - Bộ nhớ quyết định",
  description:
    "Truy vấn lý do đằng sau mọi quyết định mà AI agent của bạn đưa ra.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={inter.variable}>
      <body>
        <Aurora />
        <StarField />
        <MountainSilhouette />
        <Providers>
          <div className="relative z-[2]">{children}</div>
        </Providers>
      </body>
    </html>
  );
}
