import { createCustomerAddress } from '@/actions/customer-actions'
import { getNearbyRestaurants } from '@/actions/restaurant-actions'
import { getCustomer } from '@/lib/fetchs'
import { updateDefaultAddressId } from '@/lib/session'
import { cookies } from 'next/headers'
import AddressPrompt from '@/components/address-prompt'
import PromotionCard from '@/components/promotion-card'
import RestaurantList from '@/components/restaurant-list'

export default async function Page() {
  const defaultAddrId = await getDefaultAddressid()
  if (!defaultAddrId) {
    return <AddressPrompt onConfirmAddress={createCustomerAddress} />
  }
  const restaurants = await getNearbyRestaurants(defaultAddrId)

  return (
    <div className="h-[80vh] bg-linear-to-r from-slate-900 to-slate-700 rounded-b-3xl">
      <section className="flex justify-center p-4 overflow-x-hidden">
        <PromotionCard />
      </section>
      <section className="bg-white p-6 rounded-3xl shadow-lg border border-gray-100">
        <h2 className="text-2xl font-bold text-gray-800 mb-4">
          Restaurants near you
        </h2>
        <RestaurantList restaurants={restaurants} />
      </section>
    </div>
  )
}

async function getDefaultAddressid(): Promise<string | undefined> {
  // - check from cookie (bff)
  const defaultAddrId = (await cookies()).get('default_address_id')?.value

  // - check from backend (server)
  if (!defaultAddrId) {
    const c = await getCustomer()
    if (c.defaultAddressId) {
      updateDefaultAddressId(c.defaultAddressId)
    }
  }

  return defaultAddrId
}
