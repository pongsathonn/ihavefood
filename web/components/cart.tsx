'use client'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Coupon, MenuItem } from '@/lib/types'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'

export default function Cart({
  cartItems,
  deliveryFee,
  coupons,
  onRemoveMenuItem,
}: {
  cartItems: (MenuItem & { quantity: number })[]
  deliveryFee: number
  coupons: Coupon[]
  onRemoveMenuItem: (itemId: string) => void
}) {
  const [couponInput, setCouponInput] = useState('')
  const [appliedCoupon, setAppliedCoupon] = useState<string>('')

  const [couponMsg, setCouponMsg] = useState<
    | {
        text: string
        type: 'error' | 'success'
      }
    | undefined
  >(undefined)

  const handleApplyCoupon = (code: string) => {
    const trimmed = code.trim().toUpperCase()
    const matched = coupons.find((c) => c.code === trimmed)

    if (!matched) {
      setCouponMsg({
        text: 'Invalid coupon code. Please try again.',
        type: 'error',
      })
      setAppliedCoupon('')
      return
    }

    setAppliedCoupon(trimmed)
    setCouponInput(trimmed)

    if (matched.percentDiscount) {
      const percent = matched.percentDiscount.percent
      setDiscount(() => foodTotal * (percent / 100))
      setCouponMsg({
        text: `Coupon applied! You got ${percent}% off`,
        type: 'success',
      })
    }

    if (matched.freeDelivery) {
      setDiscount(deliveryFee)
      setCouponMsg({
        text: `Coupon applied! Free Delivery.`,
        type: 'success',
      })
    }
  }

  const handleRemoveCoupon = () => {
    if (!appliedCoupon) return
    setAppliedCoupon('')
    setCouponInput('')
    setCouponMsg(undefined)
    setDiscount(0)
  }

  const foodTotal = cartItems.reduce(
    (sum, item) => sum + item.price * item.quantity,
    0,
  )

  // discount FREEDLEIVERY = deliveryFee
  // discount coupon = calculate with coupon such as SAVE20 will make discount 20% of foodTotal(exclude delivery fee)
  const [discount, setDiscount] = useState(0)
  const total = foodTotal + deliveryFee - discount

  return (
    <aside className="col-span-1">
      <Card className="sticky top-40 rounded-3xl shadow-lg border-gray-100">
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-gray-800">
            Your Cart
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Cart Items List */}
          <div className="min-h-25 ">
            {cartItems.length <= 0 ? (
              <p className="text-gray-500 italic text-sm">
                Your cart is empty.
              </p>
            ) : (
              cartItems.map((item) => {
                return (
                  <div
                    key={item.itemId}
                    className="flex justify-between items-center py-3 border-b border-gray-50 last:border-0"
                  >
                    <div className="flex flex-col">
                      <span className="font-medium text-sm text-gray-800">
                        {item.foodName}
                        <span className="ml-2 text-muted-foreground font-normal">
                          x{item.quantity || 1}
                        </span>
                      </span>
                      <span className="text-xs text-amber-600 font-semibold">
                        ฿
                        {(
                          (item.price || 0) * (item.quantity || 1)
                        ).toLocaleString()}
                      </span>
                    </div>

                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onRemoveMenuItem(item.itemId)}
                      className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                    >
                      <Trash2 className="h-4 w-4" />{' '}
                    </Button>
                  </div>
                )
              })
            )}
          </div>

          {/* Coupon Section */}
          <div className="space-y-3">
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Available Coupons
            </h3>

            <div className="flex flex-wrap gap-2">
              {coupons.map((c) => {
                const isApplied = appliedCoupon === c.code

                return (
                  <Badge
                    key={c.code}
                    variant="secondary"
                    className={`px-3 py-1 transition-colors ${
                      isApplied
                        ? 'opacity-40 cursor-not-allowed'
                        : 'cursor-pointer hover:bg-gray-200'
                    }`}
                    onClick={() => handleApplyCoupon(c.code)}
                  >
                    {c.code}
                  </Badge>
                )
              })}
            </div>

            <div className="flex gap-2">
              <Input
                placeholder={appliedCoupon ? appliedCoupon : 'Enter code'}
                className="bg-gray-50/50"
                type="text"
                value={couponInput}
                disabled={!!appliedCoupon}
                onChange={(e) => {
                  if (!appliedCoupon) {
                    setCouponInput(e.target.value)
                  }
                }}
              />
              {appliedCoupon ? (
                <Button
                  variant="ghost"
                  className="text-red-500 hover:text-red-500"
                  size="sm"
                  onClick={handleRemoveCoupon}
                >
                  Remove
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!couponInput.trim()}
                  onClick={() => handleApplyCoupon(couponInput)}
                >
                  Apply
                </Button>
              )}
            </div>

            {couponMsg && (
              <p
                className={
                  couponMsg.type === 'error' ? 'text-red-500' : 'text-green-500'
                }
              >
                {couponMsg.text}
              </p>
            )}
          </div>

          <Separator />

          {/* Pricing Breakdown */}
          <div className="space-y-2">
            <PriceRow label="Food" amount={`฿${foodTotal}`} />
            <PriceRow label="Delivery" amount={`฿${deliveryFee}`} />
            <PriceRow label="Discount" amount={`-฿${discount}`} isDiscount />
            <div className="flex justify-between items-center pt-2">
              <span className="text-lg font-bold text-gray-800">Total</span>
              <span className="text-xl font-bold text-amber-600">฿{total}</span>
            </div>
          </div>
        </CardContent>

        <CardFooter>
          <Button
            size="lg"
            className="w-full rounded-full bg-amber-500 hover:bg-amber-600 text-white font-bold shadow-md"
            // onClick={() => handleOrderPlace()}
          >
            Order Now
          </Button>
        </CardFooter>
      </Card>
    </aside>
  )
}

function PriceRow({
  label,
  amount,
  isDiscount = false,
}: {
  label: string
  amount: string
  isDiscount?: boolean
}) {
  return (
    <div className="flex justify-between text-sm">
      <span className="font-medium text-gray-600">{label}:</span>
      <span
        className={
          isDiscount ? 'font-bold text-red-500' : 'font-bold text-gray-800'
        }
      >
        {amount}
      </span>
    </div>
  )
}
