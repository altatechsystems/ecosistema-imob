import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "Imobiliária - Encontre seu imóvel ideal",
  description: "Encontre apartamentos, casas e imóveis comerciais para venda e aluguel",
  keywords: ["imóveis", "apartamentos", "casas", "venda", "aluguel", "imobiliária"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR">
      <body className={`${inter.variable} font-sans antialiased`}>
        {children}
      </body>
    </html>
  );
}
