import { OrderStatus } from '@/lib/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type OrderStatusState = {
    status: OrderStatus | undefined
    setStatus: (status: OrderStatus) => void
    clearStatus: () => void
}

const useOrderStatus = create<OrderStatusState>()(
    persist(
        (set) => ({
            status: undefined,
            setStatus: (status: OrderStatus) => set({ status: status }),
            clearStatus: () => set({ status: undefined }),
        }),
        {
            name: 'delivery-status',
        },
    ),
)

export default useOrderStatus