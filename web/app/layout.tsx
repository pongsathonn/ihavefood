import Header from '@/components/header'
import { TooltipProvider } from '@/components/ui/tooltip'
import { Toaster } from "@/components/ui/sonner"
import './globals.css'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <title>IHAVEFOOD</title>
      </head>
      <body>
        <TooltipProvider>
          <Header />
          {children}
        </TooltipProvider>
        <Toaster richColors position="bottom-center" />
      </body>
    </html>
  )
}
