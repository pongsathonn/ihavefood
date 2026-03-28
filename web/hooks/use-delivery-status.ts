import { updateDeliveryStatus } from '@/actions/delivery-actions'
import { DeliveryStatus } from '@/lib/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type DeliveryStatusState = {
    status: DeliveryStatus | undefined
    setStatus: (orderId: string, status: DeliveryStatus) => Promise<void>
    clearStatus: () => void
}

const useDeliveryStatus = create<DeliveryStatusState>()(
    persist(
        (set) => ({
            status: undefined,
            setStatus: async (orderId: string, status: DeliveryStatus) => {
                try {
                    await updateDeliveryStatus({ orderId, status })
                    set({ status: status })
                } catch (error) {
                    console.error("Failed to update delivery status:", error)
                    throw error
                }
            },

            clearStatus: () => set({ status: undefined }),
        }),
        {
            name: 'delivery-status',
        },
    ),
)

export default useDeliveryStatus