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
  addressId: z.uuidv4(),
  addressName: z.string(),
  subDistrict: z.string(),
  district: z.string(),
  province: z.string(),
  postalCode: z.string(),
})

export const CustomerSchema = z.object({
  customerId: z.uuidv4(),
  username: z.string(),
  email: z.email(),
  phone: z.string().optional(),
  defaultAddressId: z.uuidv4().optional(),
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
  userId: z.uuidv4(),
  accessToken: z.string(),
  role: Role,
})

export const ImageInfoSchema = z.object({
  url: z.string(),
  type: z.string(),
})

export const MenuItemSchema = z.object({
  itemId: z.uuidv4(),
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
  freeDelivery: z.object().optional(),
})
export const CouponsArraySchema = z.array(CouponSchema)

export const MerchantSchema = z
  .object({
    merchantId: z.uuidv4(),
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

export const PaymentStatusSchema = z.enum({
  PAYMENT_STATUS_UNSPECIFIED: 0,
  PAYMENT_STATUS_PENDING: 1,
  PAYMENT_STATUS_PAID: 2,
});

export const OrderStatusSchema = z.enum({
  ORDER_STATUS_UNSPECIFIED: 0,
  ORDER_STATUS_PENDING: 1,
  ORDER_STATUS_PREPARING_ORDER: 2,
  ORDER_STATUS_FINDING_RIDER: 3,
  ORDER_STATUS_WAIT_FOR_PICKUP: 4,
  ORDER_STATUS_ONGOING: 5,
  ORDER_STATUS_DELIVERED: 6,
  ORDER_STATUS_CANCELLED: 7,
});

export const DeliveryStatusSchema = z.enum({
  DELIVERY_STATUS_UNSPECIFIED: 0,
  DELIVERY_STATUS_RIDER_PENDING: 1,
  DELIVERY_STATUS_RIDER_ACCEPTED: 2,
  DELIVERY_STATUS_RIDER_PICKED_UP: 3,
  DELIVERY_STATUS_RIDER_DELIVERED: 4,
});


export const OrderItem = z.object({
  itemId: z.uuidv4(),
  quantity: z.number().int().min(1),
  note: z.string().optional().or(z.literal('')),
});

export const PlaceOrderSchema = z
  .object({
    requestId: z.uuidv4(),
    orderId: z.uuidv4(),
    customerId: z.uuidv4(),
    merchantId: z.uuidv4(),
    items: z.array(OrderItem),
    couponCode: z.string().optional().nullable(),
    couponDiscount: z.number().optional(),
    deliveryFee: z.number().min(0),
    total: z.number().min(0),
    customerAddress: AddressSchema,
    merchantAddress: AddressSchema,
    customerPhone: z.string().optional(),
    paymentMethods: PaymentMethodSchema.or(z.string()),
    paymentStatus: PaymentStatusSchema.default(0),
    orderStatus: OrderStatusSchema.default(0),
  })
  .transform((v) => ({
    requestId: v.requestId,
    orderId: v.orderId,
    customerId: v.customerId,
    restaurantId: v.merchantId,
    items: v.items,
    couponCode: v.couponCode,
    discount: v.couponDiscount,
    deliveryFee: v.deliveryFee,
    total: v.total,
    customerAddress: v.customerAddress,
    merchantAddress: v.merchantAddress,
    customerPhone: v.customerPhone,
    paymentMethods: v.paymentMethods,
    paymentStatus: v.paymentStatus,
    orderStatus: v.orderStatus,
  }));


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
export type DeliveryStatus = z.infer<typeof DeliveryStatusSchema>
export type OrderStatus = z.infer<typeof OrderStatusSchema>
export type PlaceOrder = z.infer<typeof PlaceOrderSchema>
