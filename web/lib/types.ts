import * as z from 'zod'

export type RestaurantWithEst = Restaurant & {
  distance: number
  deliveryFee: number
  eta: number
}

export const Role = z.enum(['GUEST', 'CUSTOMER'])

const SocialSchema = z.object({
  facebook: z.string().optional(),
  instagram: z.string().optional(),
  line: z.string().optional(),
})

const TimestampSchema = z.iso.datetime()

export const AddressSchema = z.object({
  addressId: z.uuid({ version: 'v4' }),
  addressName: z.string(),
  subDistrict: z.string(),
  district: z.string(),
  province: z.string(),
  postalCode: z.string(),
})

export const CustomerSchema = z.object({
  customerId: z.uuid({ version: 'v4' }),
  username: z.string(),
  email: z.email(),
  phone: z.string().optional(),

  social: SocialSchema.optional(),
  addresses: z.array(AddressSchema).optional(),

  createTime: TimestampSchema,
})

export const MerchantStatusSchema = z.enum([
  'STORE_STATUS_OPEN',
  'STORE_STATUS_CLOSED',
])

export const PaymentMethodSchema = z.enum([
  'PAYMENT_METHOD_CASH',
  'PAYMENT_METHOD_CREDIT_CARD',
  'PAYMENT_METHOD_PROMPT_PAY',
])

export const SessionPayloadSchema = z.object({
  userId: z.uuid({ version: 'v4' }),
  accessToken: z.string(),
  role: Role,
})

export const ImageInfoSchema = z.object({
  url: z.string(),
  type: z.string(),
})

export const MenuItemSchema = z.object({
  itemId: z.uuid({ version: 'v4' }),
  foodName: z.string(),
  price: z.number().nonnegative(),
  imageInfo: ImageInfoSchema.optional(),
})

export const CouponSchema = z.object({
  code: z.string(),
  expiresAt: z.string(),
  quantityCount: z.number(),
  percentDiscount: z
    .object({
      percent: z.number().min(1),
    })
    .optional(),
  freeDelivery: z.object({}).optional(),
})
export const CouponsArraySchema = z.array(CouponSchema)

export const MerchantSchema = z
  .object({
    merchantId: z.uuid({ version: 'v4' }),
    merchantName: z.string(),
    imageInfo: ImageInfoSchema,
    status: MerchantStatusSchema,
    menu: z.array(MenuItemSchema),
    address: AddressSchema,
    phone: z.string(),
    email: z.email(),
  })
  .transform((v) => ({
    restaurantId: v.merchantId,
    restaurantName: v.merchantName,
    imageInfo: v.imageInfo,
    status: v.status,
    menu: v.menu,
    address: v.address,
    phone: v.phone,
    email: v.email,
  }))

export const PlaceOrderSchema = z
  .object({
    requestId: z.uuid({ version: 'v4' }),
    customerId: z.uuid({ version: 'v4' }),
    merchantId: z.uuid({ version: 'v4' }),
    items: z.array(MenuItemSchema),
    couponCode: z.string().optional(),
    discount: z.number().min(0),
    customerAddressId: z.string(),
    paymentMethods: PaymentMethodSchema,
  })
  .transform((v) => ({
    requestId: v.requestId,
    customerId: v.customerId,
    restaurantId: v.merchantId,
    items: v.items,
    couponCode: v.couponCode,
    discount: v.discount,
    customerAddressId: v.customerAddressId,
    paymentMethods: v.paymentMethods,
  }))

// NOTE: naming
// Merchant used in Backend to represent Restaurant, Retail Shop and other stores.
// However, in the frontend UI. Restaurant is more commonly used, so I rename it here
// from Merchant to Restaurant.
export type Coupon = z.infer<typeof CouponSchema>
export type RestaurantStatus = z.infer<typeof MerchantStatusSchema>
export type PaymentMethod = z.infer<typeof PaymentMethodSchema>
export type SessionPayload = z.infer<typeof SessionPayloadSchema>
export type ImageInfo = z.infer<typeof ImageInfoSchema>
export type Address = z.infer<typeof AddressSchema>
export type MenuItem = z.infer<typeof MenuItemSchema>
export type Restaurant = z.infer<typeof MerchantSchema>
export type Customer = z.infer<typeof CustomerSchema>
export type PlaceOrder = z.infer<typeof PlaceOrderSchema>
