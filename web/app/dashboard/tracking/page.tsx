'use client'

import { useRouter } from "next/navigation"
import { ChevronLeft, Phone, AlertCircle, MapPin, Bike } from "lucide-react"
import LiveActivity from "@/components/live-activity"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

export default function OrderTrackingPage() {
    const router = useRouter()

    return (
        <section className="max-w-6xl mx-auto p-4 md:p-6 space-y-6">
            {/* Header Area */}
            <div className="flex items-center gap-4">
                <Button
                    variant="outline"
                    size="icon"
                    onClick={() => router.back()}
                    className="rounded-full border-slate-200 hover:bg-slate-100"
                >
                    <ChevronLeft className="h-5 w-5 text-slate-600" />
                </Button>
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">Order Tracking</h2>
                    <p className="text-sm text-slate-500">Order #ORD-88291 • Arriving in 15-20 mins</p>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Left Column: Map & Status (Takes 2 columns on large screens) */}
                <div className="lg:col-span-2 space-y-6">
                    <LiveActivity />

                    <Card className="overflow-hidden border-none shadow-md rounded-3xl">
                        <div className="relative aspect-video bg-slate-100">
                            {/* Map Placeholder */}
                            <iframe
                                src="https://www.google.com/maps/embed?pb=!1m14!1m12!1m3!1d15501.444855325562!2d100.523186!3d13.756331!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!5e0!3m2!1sen!2sth!4v1715800000000!5m2!1sen!2sth"
                                width="100%"
                                height="100%"
                                style={{ border: 0 }}
                                loading="lazy"
                                className="grayscale contrast-125 opacity-80"
                            ></iframe>

                            {/* Overlay Badge */}
                            <Badge className="absolute top-4 left-4 bg-white/90 text-slate-900 backdrop-blur hover:bg-white/90 px-3 py-1 border-none shadow-sm">
                                <MapPin className="w-3 h-3 mr-1 text-amber-600" />
                                On the way to you
                            </Badge>
                        </div>
                        <CardContent className="p-4 bg-amber-50 border-t border-amber-100">
                            <p className="text-xs text-amber-700 font-medium flex items-center justify-center">
                                <AlertCircle className="w-3 h-3 mr-2" />
                                Your rider is picking up your order from the restaurant
                            </p>
                        </CardContent>
                    </Card>
                </div>

                {/* Right Column: Details */}
                <div className="space-y-6">
                    {/* Order Details Card */}
                    <Card className="rounded-3xl border-slate-100 shadow-sm">
                        <CardHeader>
                            <CardTitle className="text-lg">Order Details</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-500">Restaurant</span>
                                <span className="font-semibold text-slate-900">Burger King - Siam Paragon</span>
                            </div>
                            <Separator className="bg-slate-100" />
                            <div className="space-y-2">
                                <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Items</span>
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-700">2x Whopper Junior</span>
                                    <span className="font-medium">฿320.00</span>
                                </div>
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-700">1x Large Fries</span>
                                    <span className="font-medium">฿65.00</span>
                                </div>
                            </div>
                            <Separator className="bg-slate-100" />
                            <div className="flex justify-between items-center pt-2">
                                <span className="font-bold text-slate-900 text-lg">Total</span>
                                <span className="font-bold text-amber-600 text-lg">฿385.00</span>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Rider Info Card */}
                    <Card className="rounded-3xl border-slate-100 shadow-sm overflow-hidden">
                        <CardContent className="p-6">
                            <div className="flex items-center gap-4">
                                <Avatar className="h-14 w-14 border-2 border-amber-100">
                                    <AvatarImage src="https://github.com/shadcn.png" />
                                    <AvatarFallback>RD</AvatarFallback>
                                </Avatar>
                                <div className="flex-1">
                                    <h4 className="font-bold text-slate-900">Somsak Delivery</h4>
                                    <p className="text-xs text-slate-500 flex items-center">
                                        <Bike className="w-3 h-3 mr-1" />
                                        Honda Click 125i (1กข 1234)
                                    </p>
                                </div>
                                <Button size="icon" variant="secondary" className="rounded-full bg-amber-100 text-amber-700 hover:bg-amber-200">
                                    <a href="tel:0812345678">
                                        <Phone className="h-4 w-4" />
                                    </a>
                                </Button>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Support Link */}
                    <Button variant="ghost" className="w-full text-slate-400 hover:text-red-500 hover:bg-red-50/50 text-xs py-8 border-2 border-dashed border-slate-100 rounded-3xl">
                        Having an issue? Report a Problem
                    </Button>
                </div>
            </div>
        </section>
    )
}