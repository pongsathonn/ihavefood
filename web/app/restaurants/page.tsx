import { createCustomerAddress } from '@/app/restaurants/actions'
import AddressPrompt from '@/components/address-prompt'
import PromotionCard from '@/components/promotion-card'
import RestaurantList from '@/components/restaurant-list'
import { cookies } from 'next/headers'

export default async function Page() {
  // const promotions = PromotionImages()

  const defaultAddrId: string | undefined = (await cookies()).get(
    'default_address_id',
  )?.value

  return (
    <div className="h-[80vh] bg-linear-to-r from-slate-900 to-slate-700 rounded-b-3xl">
      <section className="flex justify-center p-4 overflow-x-hidden">
        <PromotionCard />
      </section>
      <section className="bg-white p-6 rounded-3xl shadow-lg border border-gray-100">
        <h2 className="text-2xl font-bold text-gray-800 mb-4">
          Restaurants near you
        </h2>
        {!defaultAddrId ? (
          <AddressPrompt onConfirmAddress={createCustomerAddress} />
        ) : (
          <RestaurantList customerAddrId={defaultAddrId} />
        )}
      </section>
    </div>
  )
}
