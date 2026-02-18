import './globals.css'

export default function RootLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <html lang="en" suppressHydrationWarning>
            <title>IHAVEFOOD</title>
            <body>
                {children}
            </body>
        </html >
    )
}


