'use server'

import {
    authentication,
    getCustomerId,
    getDefaultAddressId,
    getSession,
} from '@/lib/session'
import { DeliveryStatus, OrderStatus, PlaceOrder, PlaceOrderSchema } from '@/lib/types'

type CartOrder = {
    restaurantId: string
    cartItems: ({ itemId: string, quantity: number, note: string })[]
    appliedCoupon: string
    discount?: number
}

export const updateDeliveryStatus = async function ({
    orderId,
    status,
}: {
    orderId: string,
    status: DeliveryStatus,
}): Promise<void> {

    if (status === 0) {
        throw new Error('Status should not be undefined')
    }

    const { isAuth } = await authentication()
    if (!isAuth) {
        throw new Error('Unauthorized')
    }

    const serverUrl = process.env.SERVER_URL
    const DEMO_RIDER_ID = process.env.DEMO_RIDER_ID
    if (!serverUrl) throw new Error('SERVER_URL is not defined')
    if (!DEMO_RIDER_ID) throw new Error('DEMO_RIDER_ID is not defined')
    const token = await getSession()

    const res = await fetch(`${serverUrl}/api/deliveries/${orderId}/status`, {
        method: 'PATCH',
        headers: {
            Accept: 'application/json',
            Cookie: `session=${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            rider_id: DEMO_RIDER_ID,
            status: status
        }),
    })

    if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to update delivery status: ${errorText}`)
    }
}
