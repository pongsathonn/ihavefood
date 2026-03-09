import MenuList from '@/components/menu-list'
import {
  getCustomer,
  getDeliveryFee,
  getRestaurant,
  listCoupons,
} from '@/lib/fetchs'
import { cookies } from 'next/headers'

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  let customer
  try {
    customer = await getCustomer()
  } catch (err) {
    if (err instanceof Error) console.error(err.message)
    return <p>Failed to load customer</p>
  }

  let fee: number | undefined = undefined
  try {
    const defaultAddrId: string | undefined = (await cookies()).get(
      'default_address_id',
    )?.value
    if (!defaultAddrId) throw new Error('default address not found')
    fee = await getDeliveryFee(customer.customerId, defaultAddrId, id)
  } catch (err) {
    if (err instanceof Error) console.error(err.message)
  }

  try {
    const [coupons, restaurant] = await Promise.all([
      listCoupons(),
      getRestaurant(id),
    ])
    return (
      <MenuList coupons={coupons} restaurant={restaurant} deliveryFee={fee} />
    )
  } catch (err) {
    if (err instanceof Error) console.error(err.message)
    return <p>Something went wrong</p>
  }

  // return (
  //   <MenuList coupons={coupons} restaurant={restaurant} deliveryFee={fee} />
  // )
}
