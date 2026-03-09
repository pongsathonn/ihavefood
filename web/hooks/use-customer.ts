import { Customer } from '@/lib/types'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type CustomerState = {
  customer: Customer | undefined
  setCustomer: (data: Customer) => void
  clearCustomer: () => void
}

const useCustomer = create<CustomerState>()(
  persist(
    (set) => ({
      customer: undefined,
      setCustomer: (data) => set({ customer: data }),
      clearCustomer: () => set({ customer: undefined }),
    }),
    {
      name: 'customer-storage',
    },
  ),
)

export default useCustomer
