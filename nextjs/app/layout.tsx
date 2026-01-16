import './globals.css'

export default function RootLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <html lang="en">
            <title>IHAVEFOOD</title>
            <body>{children}</body>
        </html>
    )
}
