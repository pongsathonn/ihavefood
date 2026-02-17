export default function Cart() {
    return (

        // Shopping Cart Section (Initially Hidden) 
        <aside id="cart-aside" className="col-span-1 hidden">
            <div className="bg-white p-6 rounded-3xl shadow-lg border border-gray-100 sticky top-40">
                <h2 className="text-2xl font-bold text-gray-800 mb-4">Your Cart</h2>
                <div id="cart-items" className="min-h-[100px] mb-4">
                    <p id="empty-cart-message" className="text-gray-500 italic text-sm">Your cart is empty.</p>
                    {/* Cart items will be injected here by JavaScript  */}
                </div>

                {/* Coupon Section  */}
                <div className="mt-4 mb-4">
                    <h3 className="text-lg font-semibold text-gray-800 mb-2">Coupons</h3>

                    <div className="flex flex-wrap gap-2 mt-2 mb-2">
                        <button className="preset-coupon-btn bg-gray-200 text-xs px-2 py-1 rounded hover:bg-gray-300"
                            data-coupon="SAVE10">SAVE10</button>
                        <button className="preset-coupon-btn bg-gray-200 text-xs px-2 py-1 rounded hover:bg-gray-300"
                            data-coupon="SAVE50">SAVE50</button>
                        <button className="preset-coupon-btn bg-gray-200 text-xs px-2 py-1 rounded hover:bg-gray-300"
                            data-coupon="FREEDELIVERY">FREEDELIVERY</button>
                    </div>

                    <div className="flex items-center space-x-2 flex-nowrap">
                        <input
                            type="text"
                            id="coupon-input"
                            placeholder="Code"
                            className="min-w-0 flex-grow p-2 rounded-lg border border-gray-300 focus:outline-none focus:border-blue-500 text-sm"
                        >
                        </input>

                        <button id="apply-coupon-btn"
                            className="flex-shrink-0 bg-blue-500 text-white font-bold py-2 px-4 rounded-lg hover:bg-blue-600 transition disabled:bg-gray-400 text-sm"
                            disabled>
                            Apply
                        </button>
                    </div>
                    <p id="coupon-status" className="text-sm mt-2"></p>
                </div>

                <div className="border-t pt-4">
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-lg font-semibold text-gray-800">Food:</span>
                        <span id="cart-food" className="text-lg font-bold text-gray-600">฿0.00</span>
                    </div>
                    <div className="flex justify-between items-center mb-2">
                        <span className="text-lg font-semibold text-gray-800">Delivery Fee:</span>
                        <span id="cart-delivery" className="text-lg font-bold text-gray-600">฿0.00</span>
                    </div>
                    <div id="discount-row" className="flex justify-between items-center mb-2 hidden">
                        <span className="text-lg font-semibold text-gray-800">Discount:</span>
                        <span id="cart-discount" className="text-lg font-bold text-red-500">-฿0.00</span>
                    </div>
                    <div className="flex justify-between items-center mb-4">
                        <span className="text-lg font-semibold text-gray-800">Total:</span>
                        <span id="cart-total" className="text-xl font-bold text-pink-600">฿0.00</span>
                    </div>
                    <button id="order-button"
                        className="w-full bg-pink-500 hover:bg-pink-600 text-white font-bold py-3 px-4 rounded-full transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-pink-400 focus:ring-opacity-50 shadow-md disabled:bg-pink-400"
                        disabled>
                        Place Order
                    </button>
                </div>
            </div>
        </aside>
    )
}


////////////////////////////////////////
// const renderCart = () => {
//     cartItemsContainer.innerHTML = '';
//     const currentCart = localCart[currentRestaurant.merchantId] || [];

//     if (currentCart.length === 0) {
//         emptyCartMessage.style.display = 'block';
//     } else {
//         emptyCartMessage.style.display = 'none';
//         currentCart.forEach(item => {
//             const itemElement = document.createElement('div');
//             itemElement.className = 'flex justify-between items-center py-2 border-b last:border-b-0 text-sm';
//             itemElement.innerHTML = `
//             <span>${item.foodName} <span class="text-gray-500">(${item.quantity})</span></span>
//             <div class="flex items-center space-x-2">
//                 <span class="font-semibold">฿${(item.price * item.quantity).toFixed(2)}</span>
//                 <button class="remove-from-cart-btn text-red-500 hover:text-red-600 transition" data-item-id="${item.itemId}">
//                     <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
//                         <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.038 21H7.962a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
//                     </svg>
//                 </button>
//             </div>
//             `;
//             cartItemsContainer.appendChild(itemElement);
//         });
//     }
//     updateCartTotals();
// };

// const updateCartTotals = () => {
//     const currentCart = localCart[currentRestaurant.merchantId] || [];
//     const food = currentCart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
//     const couponPercent = sessionStorage.getItem('coupon_percent');
//     const deliveryFee = Number(sessionStorage.getItem('delivery_fee'));
//     let discountAmount = 0;

//     cartFoodElement.textContent = `฿${food.toFixed(2)}`;

//     if (couponApplied) {
//         const couponCode = couponInput.value.trim().toUpperCase();
//         if (couponCode === "FREEDELIVERY") {

//             // cartDeliveryElement.textContent = `-฿${deliveryFee}`;
//             // cartDeliveryElement.className = 'text-lg font-bold text-red-500';

//             discountAmount = deliveryFee;
//             discountRow.classList.remove('hidden');
//             cartDiscountElement.textContent = `-฿${discountAmount}`;

//         } else if (couponPercent > 0) {

//             // Reset delivery display to normal
//             // cartDeliveryElement.textContent = `฿${deliveryFee}`;
//             // cartDeliveryElement.className = 'text-lg font-bold text-gray-600';

//             discountAmount = food * (couponPercent / 100);
//             discountRow.classList.remove('hidden');
//             cartDiscountElement.textContent = `-฿${discountAmount}`;
//         }
//     } else {
//         // No coupon applied
//         discountAmount = 0;
//         cartDeliveryElement.textContent = `฿${deliveryFee}`;
//         cartDeliveryElement.className = 'text-lg font-bold text-gray-600';
//         discountRow.classList.add('hidden');
//     }
//     sessionStorage.setItem('discount', discountAmount);

//     const total = food + deliveryFee - discountAmount;
//     cartTotalElement.textContent = `฿${total.toFixed(2)}`;

//     orderButton.disabled = food === 0 || isOrderPlaced;
// };


// const addToCart = (foodItem) => {

//     const restaurantId = currentRestaurant.merchantId;
//     if (!localCart[restaurantId]) {
//         localCart[restaurantId] = [];
//     }
//     const currentCart = localCart[restaurantId];
//     const existingItem = currentCart.find(item => item.itemId === foodItem.itemId);
//     if (existingItem) {
//         existingItem.quantity++;
//     } else {
//         currentCart.push({ ...foodItem, quantity: 1 });
//     }

//     renderCart();
// };

// const removeFromCart = (itemId) => {
//     const restaurantId = currentRestaurant.merchantId;
//     const currentCart = localCart[restaurantId] || [];
//     const itemIndex = currentCart.findIndex(item => item.itemId === itemId);
//     if (itemIndex !== -1) {
//         const item = currentCart[itemIndex];
//         if (item.quantity > 1) {
//             item.quantity--;
//         } else {
//             currentCart.splice(itemIndex, 1);
//         }
//     }
//     renderCart();
// };






///////////////////
// const buildOrderPayload = () => {
//     const restaurantId = currentRestaurant.merchantId;
//     const customerId = sessionStorage.getItem('customer_id');
//     const appliedCoupon = sessionStorage.getItem('applied_coupon');
//     const discount = Number(sessionStorage.getItem('discount'));


//     const defaultAddr = addresses.find(a => a.isDefault);
//     if (!defaultAddr) {
//         console.error('No default address found');
//         return;
//     }
//     const customerAddressId = defaultAddr.addressId;

//     const cartItems = (localCart[restaurantId] || []).map(item => ({
//         item_id: item.itemId,
//         quantity: item.quantity,
//         note: item.note || ""
//     }));

//     const orderPayload = {
//         request_id: crypto.randomUUID(),
//         customer_id: customerId,
//         merchant_id: restaurantId,
//         items: cartItems,
//         coupon_code: appliedCoupon,
//         discount: discount,
//         customer_address_id: customerAddressId,
//         payment_methods: "PAYMENT_METHOD_CREDIT_CARD"
//     };

//     return orderPayload;
// };

/////////////////////////////
// menuContainer.addEventListener('click', (event) => {
//     const targetButton = event.target.closest('.add-to-cart-btn');
//     if (targetButton) {
//         const foodItem = JSON.parse(targetButton.dataset.item);
//         addToCart(foodItem);
//     }
// });

// cartItemsContainer.addEventListener('click', (event) => {
//     const targetButton = event.target.closest('.remove-from-cart-btn');
//     if (targetButton) {
//         const itemId = targetButton.dataset.itemId;
//         removeFromCart(itemId);
//     }
// });

// couponInput.addEventListener('input', () => {
//     applyCouponBtn.disabled = couponInput.value.trim() === '';
// });

// applyCouponBtn.addEventListener('click', () => {
//     const couponCode = couponInput.value.trim().toUpperCase();

//     const coupons = {
//         'SAVE10': { percent: 10, message: 'Coupon applied! You got 10% off.' },
//         'SAVE20': { percent: 20, message: 'Coupon applied! You got 20% off.' },
//         'SAVE30': { percent: 30, message: 'Coupon applied! You got 30% off.' },
//         'SAVE40': { percent: 40, message: 'Coupon applied! You got 40% off.' },
//         'SAVE50': { percent: 50, message: 'Coupon applied! You got 50% off.' },
//         'FREEDELIVERY': { percent: 0, message: 'Coupon applied! Delivery fee waived.' }
//     };

//     const coupon = coupons[couponCode];

//     if (coupon) {
//         couponApplied = true;
//         couponStatus.textContent = coupon.message;
//         couponStatus.className = 'text-sm mt-2 text-pink-600';
//         sessionStorage.setItem("applied_coupon", couponCode);
//         sessionStorage.setItem("coupon_percent", coupon.percent);
//     } else {
//         couponApplied = false;
//         couponStatus.textContent = 'Invalid coupon code. Please try again.';
//         couponStatus.className = 'text-sm mt-2 text-red-500';
//         sessionStorage.setItem("applied_coupon", "");
//         sessionStorage.setItem("coupon_percent", 0);
//     }

//     updateCartTotals();
// });


////////////////////////////////////////
// orderButton.addEventListener('click', async () => {

//     if (!isUserAuthenticated) {
//         alert("Please sign in to place an order.");
//         return;
//     }

//     const total = Number(cartTotalElement.textContent.replace('฿', ''));
//     sessionStorage.setItem('total', total);
//     const orderPayload = buildOrderPayload()

//     try {
//         const place_order = await createPlaceOrder(orderPayload);
//         sessionStorage.setItem("place_order", JSON.stringify(place_order));

//         const restaurantId = currentRestaurant.merchantId;
//         localCart[restaurantId] = [];
//         couponApplied = false;
//         couponInput.value = '';
//         couponStatus.textContent = '';
//         applyCouponBtn.disabled = true;
//         isOrderPlaced = true;

//         renderCart();

//         showModal();

//         // Simulate status change
//         setTimeout(() => setTrackingStep(2), 2000); // Food is being prepared
//         setTimeout(() => setTrackingStep(3), 6000); // Rider is on the way
//         setTimeout(() => setTrackingStep(4), 10000); // Delivered
//     } catch (err) {
//         console.error("Order failed:", err);
//         alert("Failed to place an order. Please try again."); // or show a small popup/modal
//     }
// });


///////////////////////////////////////////////////
// trackOrderButton.addEventListener('click', () => {
//     hideModal();

//     const order = JSON.parse(sessionStorage.getItem("place_order"));
//     if (!order) return;

//     trackingOrderId.textContent = `#${order.orderId}`;
//     trackingRestaurantName.textContent = sessionStorage.getItem('selected_merchant_name');

//     const tempRestaurant = {};
//     currentRestaurant.menu.forEach(item => {
//         tempRestaurant[item.itemId] = item.foodName
//     });


//     if (order.items && Array.isArray(order.items)) {
//         order.items.forEach(item => {

//             const foodItemHtml = `
//                 <div class="flex justify-between text-sm w-full">
//                 <span>${tempRestaurant[item.itemId]}</span>
//                 <span class="flex-shrink-0">x ${item.quantity}</span>
//                 </div>
//             `;

//             trackingFoodItems.insertAdjacentHTML('beforeend', foodItemHtml);
//         });
//     }

//     const total = sessionStorage.getItem('total');
//     trackingOrderTotal.textContent = `฿${total}`;
//     showSection('order-tracking-section');
// });