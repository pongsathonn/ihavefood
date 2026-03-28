'use server'

import {
    authentication,
    getCustomerId,
    getDefaultAddressId,
    getSession,
} from '@/lib/session'
import { PlaceOrder, PlaceOrderSchema } from '@/lib/types'

type CartOrder = {
    restaurantId: string
    cartItems: ({ itemId: string, quantity: number, note: string })[]
    appliedCoupon: string
    discount?: number
}

export const createPlaceOrder = async function (
    cartOrder: CartOrder,
): Promise<PlaceOrder> {
    const { isAuth } = await authentication()
    if (!isAuth) {
        throw new Error('Unauthorized')
    }

    const serverUrl = process.env.SERVER_URL
    if (!serverUrl) throw new Error('SERVER_URL is not defined')
    const token = await getSession()
    const customerId = await getCustomerId()
    const addrId = await getDefaultAddressId()
    if (!addrId) throw new Error('Default Address ID is not defined')

    // TODO: create type for payload
    const orderPayload = {
        request_id: crypto.randomUUID(),
        customer_id: customerId,
        merchant_id: cartOrder.restaurantId,
        items: cartOrder.cartItems,
        coupon_code: cartOrder.appliedCoupon,
        discount: cartOrder.discount,
        customer_address_id: addrId,
        payment_methods: 'PAYMENT_METHOD_CREDIT_CARD',
    }

    const res = await fetch(`${serverUrl}/api/orders/place_order`, {
        method: 'POST',
        headers: {
            Accept: 'application/json',
            Cookie: `session=${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(orderPayload),
    })

    if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to create order: ${errorText}`)
    }

    const result = PlaceOrderSchema.safeParse(await res.json())
    if (!result.success) {
        throw result.error
    }

    return result.data


    async function TODO() {

    }
}
