import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "K8S Graph Explorer - Kubernetes Topology Visualizer",
  description: "Visualize and explore your Kubernetes cluster topology with an interactive graph view",
  keywords: ["kubernetes", "k8s", "graph", "topology", "visualization", "cluster"],
};

import { ApolloWrapper } from "@/lib/apollo-wrapper";
import { Toaster } from "sonner";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ApolloWrapper>
          {children}
          <Toaster position="top-center" />
        </ApolloWrapper>
      </body>
    </html>
  );
}
