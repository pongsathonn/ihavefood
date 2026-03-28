'use client'

import { Bike, ChevronLeft, Phone } from "lucide-react"
import { notFound } from "next/navigation"
import { useEffect, useMemo } from "react"

import LiveActivity from "@/components/live-activity"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"

import { useOrderDetail } from "@/hooks/use-order-detail"
import { useRestaurantWithEst } from "@/hooks/use-restaurant"
import Link from "next/link"
import useOrderStatus from "@/hooks/use-order-status"
import useDeliveryStatus from "@/hooks/use-delivery-status"
import { DeliveryStatusSchema } from "@/lib/types"

export default function OrderTrackingPage() {
    const orderDetail = useOrderDetail((state) => state.orderDetail)
    const getRestaurantById = useRestaurantWithEst((state) => state.getRestaurantById)
    // TODO: real time tracking order status. might use stream grpc
    // const orderStatus = useOrderStatus((state) => state.status)

    const restaurant = useMemo(() => {
        return orderDetail?.restaurantId ? getRestaurantById(orderDetail.restaurantId) : null
    }, [orderDetail?.restaurantId, getRestaurantById])

    const subTotal = useMemo(() => {
        if (!orderDetail?.items || !restaurant?.menu) return 0
        return orderDetail.items.reduce((acc, item) => {
            const menu = restaurant.menu.find((m) => m.itemId === item.itemId)
            return acc + (menu?.price || 0) * item.quantity
        }, 0)
    }, [orderDetail?.items, restaurant?.menu])

    if (!orderDetail || !restaurant) {
        return notFound()
    }

    const ORDER_TRIMMED = orderDetail.orderId.slice(-4)
    const TODO_RIDER_NAME = 'Somsak Jaidee'
    const TODO_BIKE_NAME = 'Honda Wave (1234)'


    // fake update Delivery status for every 3 sec
    demoDeliveryStatusTicker()

    return (
        <section className="w-full p-4 md:p-8 max-w-5xl mx-auto space-y-6">
            <div className="flex items-start md:items-center gap-4">
                <Link href="/"> <ChevronLeft className="h-5 w-5 text-slate-600" /> </Link>
                <div className="min-w-0">
                    <h2 className="text-2xl font-bold text-slate-900 truncate">Order Tracking</h2>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 space-y-6">
                    <LiveActivity orderId={ORDER_TRIMMED} status={orderDetail.orderStatus} />

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
                        <Card className="rounded-3xl border-slate-100 shadow-sm overflow-hidden h-fit">
                            <CardContent className="p-6">
                                {orderDetail.orderStatus > 3 ? (
                                    <div className="flex items-center w-full gap-4">
                                        <Avatar className="h-14 w-14 border-2 border-amber-100 shrink-0">
                                            <AvatarImage src="https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/RIDER_PROFILE.png" />
                                            <AvatarFallback>RD</AvatarFallback>
                                        </Avatar>
                                        <div className="flex-1 min-w-0">
                                            <h4 className="font-bold text-slate-900 truncate">{TODO_RIDER_NAME}</h4>
                                            <p className="text-xs text-slate-500 flex items-center">
                                                <Bike className="w-3 h-3 mr-1 text-amber-500" />
                                                {TODO_BIKE_NAME}
                                            </p>
                                        </div>
                                        <Button size="icon" variant="secondary" className="rounded-full bg-amber-100 text-amber-700 hover:bg-amber-200 shrink-0" asChild>
                                            <a href="tel:0812345678">
                                                <Phone className="h-4 w-4" />
                                            </a>
                                        </Button>
                                    </div>
                                ) : (
                                    <div className="flex items-center w-full gap-4 animate-pulse">
                                        <Skeleton className="h-14 w-14 rounded-full shrink-0" />
                                        <div className="flex-1 space-y-2">
                                            <Skeleton className="h-4 w-24" />
                                            <div className="flex items-center gap-2">
                                                <Skeleton className="h-3 w-32" />
                                                <span className="flex gap-1">
                                                    <span className="w-1 h-1 bg-slate-300 rounded-full animate-bounce" />
                                                    <span className="w-1 h-1 bg-slate-300 rounded-full animate-bounce [animation-delay:-0.3s]" />
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </CardContent>
                        </Card>

                        <Card className="rounded-3xl border-slate-100 shadow-sm">
                            <CardHeader className="pb-3">
                                <CardTitle className="text-lg">Order Summary</CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-500">Restaurant</span>
                                    <span className="font-semibold text-slate-900">{restaurant.restaurantName}</span>
                                </div>
                                <Separator className="opacity-50" />

                                <div className="space-y-3">
                                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Items</span>
                                    {orderDetail.items.map((item) => {
                                        const menu = restaurant.menu.find((m) => m.itemId === item.itemId)
                                        return (
                                            <div className="flex justify-between text-sm" key={item.itemId}>
                                                <span className="text-slate-700">
                                                    <span className="text-slate-700 mr-2">{item.quantity}x</span>
                                                    {menu?.foodName}
                                                </span>
                                                <span className="font-medium text-slate-900">฿{(menu?.price || 0) * item.quantity}</span>
                                            </div>
                                        )
                                    })}
                                </div>

                                <Separator className="opacity-50" />

                                <div className="space-y-1.5 text-sm text-slate-600">
                                    <div className="flex justify-between">
                                        <span>Delivery Fee</span>
                                        <span>฿{orderDetail.deliveryFee}</span>
                                    </div>
                                    {(orderDetail?.discount ?? 0) > 0 && (
                                        <div className="flex justify-between text-red-600 font-medium">
                                            <span>Discount</span>
                                            <span>-฿{orderDetail!.discount}</span>
                                        </div>
                                    )}
                                </div>

                                <Separator className="bg-slate-100" />

                                <div className="flex justify-between items-center pt-1">
                                    <span className="font-bold text-slate-900">Total</span>
                                    <span className="font-bold text-amber-600 text-xl">฿{orderDetail.total}</span>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </section>
    )
}

function demoDeliveryStatusTicker(orderId: string) {
    const setStatus = useDeliveryStatus((state) => state.setStatus)

    useEffect(() => {
        const steps: (1 | 2 | 3 | 4)[] = [1, 2, 3, 4]
        let index = 0

        const interval = setInterval(() => {
            const status = steps[index]
            setStatus(orderId, status)
            index++
            if (index >= steps.length) {
                clearInterval(interval)
            }
        }, 3000)
        return () => clearInterval(interval)
    }, [setStatus])
    return null
}