import { PlaceOrder } from '@/lib/types';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type OrderStore = {
    orderDetail: PlaceOrder | undefined;
    setOrderDetail: (data: PlaceOrder) => void;
    clearOrderDetail: () => void;
};

export const useOrderDetail = create<OrderStore>()(
    persist(
        (set) => ({
            orderDetail: undefined,
            setOrderDetail: (data) => set({ orderDetail: data }),
            clearOrderDetail: () => set({ orderDetail: undefined }),
        }),
        {
            name: 'order-storage',
        }
    )
);