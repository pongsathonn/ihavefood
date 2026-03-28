'use client'

import { OrderStatus } from '@/lib/types'
import { Bike, ChefHat, MapPin, Store } from 'lucide-react'
import React from 'react'

export default function LiveActivity({
  orderId,
  status,
}: {
  orderId: string | undefined
  status: OrderStatus | undefined
}) {
  const steps = [
    { id: 'ORDERED', icon: Store, threshold: 1 },
    { id: 'PREPARING', icon: ChefHat, threshold: 2 },
    { id: 'DELIVERY', icon: Bike, threshold: 3 },
    { id: 'ARRIVED', icon: MapPin, threshold: 4 },
  ]

  const statusMap = {
    0: {
      step: 1,
      progress: 0,
      label: 'Processing...',
    },
    1: {
      step: 1,
      progress: 0,
      label: 'Order Placed',
    },
    2: {
      step: 2,
      progress: 50,
      label: 'Preparing Order',
    },
    3: {
      step: 2,
      progress: 80,
      label: 'Finding Rider',
    },
    4: {
      step: 3,
      progress: 20,
      label: 'Waiting for Pickup',
    },
    5: {
      step: 3,
      progress: 70,
      label: 'On the Way',
    },
    6: {
      step: 4,
      progress: 100,
      label: 'Arrived',
    },
    7: {
      step: 1,
      progress: 0,
      label: 'Cancelled',
    },
  } as const

  const current = statusMap[status] ?? statusMap[0]


  return (
    <div className="w-full py-6 rounded-2xl bg-white border border-slate-100 px-8 shadow-sm">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 bg-amber-50 rounded-full flex items-center justify-center">
            🥘
          </div>
          <div>
            <h3 className="text-slate-900 font-bold text-sm">
              {current.label}
            </h3>
            {/* <p className="text-slate-400 text-xs">
              Arriving in 12 mins
            </p> */}
          </div>
        </div>

        <div className="bg-amber-50 border border-amber-100 px-3 py-1 rounded-full">
          <p className="text-amber-600 font-mono text-xs font-bold">
            #ORDER-{orderId}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {steps.map((step, index) => (
          <React.Fragment key={step.id}>
            <div
              className={`p-2 rounded-full ${current.step >= step.threshold
                ? 'bg-amber-100 text-amber-600'
                : 'bg-slate-50 text-slate-300'
                }`}
            >
              <step.icon className="h-5 w-5" />
            </div>

            {index < steps.length - 1 && (
              <div className="flex-1 h-1 bg-slate-100 rounded-full">
                <div
                  className={`h-1 bg-amber-500 rounded-full transition-all duration-700 ${current.step > step.threshold
                    ? 'w-full'
                    : current.step === step.threshold
                      ? 'w-full'
                      : 'w-0'
                    }`}
                />
              </div>
            )}
          </React.Fragment>
        ))}
      </div>
    </div>
  )
}