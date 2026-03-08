import { TooltipProvider } from '@/components/ui/tooltip'
import './globals.css'
import Header from '@/components/header'

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
      </body>
    </html>
  )
}
