import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Valt - AI Secret Vault",
  description: "Human-in-the-loop secret management for AI agents",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-background antialiased">
        {children}
      </body>
    </html>
  );
}
