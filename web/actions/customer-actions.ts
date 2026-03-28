'use server'

import { authentication, getCustomerId, getSession } from '@/lib/session'
import { Address, AddressSchema } from '@/lib/types'
import { cookies } from 'next/headers'
import { NewAddress } from './restaurant-actions'

export const createCustomerAddress = async function (
    newAddress: NewAddress,
): Promise<Address> {
    const { isAuth } = await authentication()
    if (!isAuth) {
        throw new Error('Unauthorized')
    }

    const customerId = await getCustomerId()
    const serverUrl = process.env.SERVER_URL
    if (!serverUrl) throw new Error('SERVER_URL is not defined')
    const token = await getSession()
    const res = await fetch(`${serverUrl}/api/customers/${customerId}/address`, {
        method: 'POST',
        headers: {
            Accept: 'application/json',
            Cookie: `session=${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            address: {
                address_name: newAddress.addressName,
                sub_district: newAddress.subDistrict,
                district: newAddress.district,
                province: newAddress.province,
                postal_code: newAddress.postalCode,
            },
        }),
    })
    if (!res.ok) {
        const errorText = await res.text()
        throw new Error(`Failed to create customer address: ${errorText}`)
    }
    const createdAddr = AddressSchema.safeParse(await res.json()).data
    if (!createdAddr) {
        throw new Error('Failed to parse Custoemr address')
    }

    ; (await cookies()).set('default_address_id', createdAddr.addressId)

    return createdAddr
}
