'use client'

import {
    useState,
    Dispatch,
    SetStateAction,
} from 'react';
import { MenuItem, Restaurant } from '@/app/lib/definitions';


const renderMenu = (menu: MenuItem[]) => {
    if (!menu || menu.length === 0) {
        return <p className="text-gray-500 italic">This restaurant has no menu items.</p>;
    }

    return menu.map((item, index) => (
        <div key={index} className="flex items-center justify-between p-4 bg-white rounded-2xl border border-gray-100 shadow-sm">
            <div className="flex items-center space-x-4">
                <img src={item.imageInfo.url} alt={item.foodName} className="w-16 h-16 object-cover rounded-lg" />
                <div>
                    <h4 className="text-md font-semibold text-gray-800">{item.foodName}</h4>
                    <span className="text-sm text-gray-500">฿{item.price.toFixed(2)}</span>
                </div>
            </div>
            <button
                className="add-to-cart-btn bg-gray-900 text-white px-4 py-2 rounded-xl text-sm font-bold hover:bg-black transition"
                data-item={JSON.stringify(item)}
            >
                Add
            </button>
        </div>
    ));
};

export default function FoodCategory({
    restaurants,
    currentRestaurant,
    setCurrentRestaurant,
}: {
    restaurants?: Restaurant[],
    currentRestaurant?: Restaurant,
    setCurrentRestaurant: Dispatch<SetStateAction<Restaurant | undefined>>,
}) {

    if (restaurants == undefined || restaurants.length == 0) {
        return <p className="text-gray-200 italic">Restaurants are undefined</p>;
    }

    // const [customerAddress, setCustomerAddress] = useState(() => {
    //     const saved = localStorage.getItem('customer_address');
    //     return saved ? JSON.parse(saved) : undefined;
    // });
    // const updatedAddress = customerAddress !== undefined;

    const renderRestaurants = ({
        restaurants,
    }: {
        restaurants?: Restaurant[],
    }) => {

        if (!restaurants || restaurants.length === 0) {
            return <p className="text-gray-500 italic">No restaurants available.</p>;
        }

        return restaurants.map((restaurant, index) => {
            const isClosed = restaurant.status !== "STORE_STATUS_OPEN";

            return (
                <div
                    key={restaurant.restaurantId || index}
                    onClick={() => !isClosed && setCurrentRestaurant(restaurant)}
                    className={`
                flex flex-col h-full 
                bg-gray-50 rounded-2xl shadow-sm overflow-hidden cursor-pointer hover:shadow-lg transition-all 
                ${isClosed ? 'opacity-50 pointer-events-none' : ''
                        }
            `}
                >
                    <img
                        src={restaurant.imageInfo.url}
                        alt={restaurant.restaurantName}
                        className="w-full aspect-[16/9] md:aspect-[4/3] object-cover shrink-0"
                    />

                    <div className="p-4 flex flex-col flex-grow">
                        <h3 className="text-lg font-semibold text-gray-800 line-clamp-1 mb-1">
                            {restaurant.restaurantName}
                        </h3>

                        <div className="flex-grow">
                            {!isClosed && (
                                <p className="text-sm text-gray-500 flex items-center gap-2 whitespace-nowrap">
                                    <span>Delivery: <span className="text-red-700">฿20.00</span></span>
                                    <span className="text-gray-300">|</span>
                                    <span className="truncate">10.8 km (6hr 30min)</span>
                                </p>
                            )}
                        </div>

                        <p className="text-sm text-gray-500 pt-2">
                            Status: {isClosed ? 'CLOSED' : 'OPEN'}
                        </p>
                    </div>
                </div>
            );
        });
    }

    return (
        <div >
            {(currentRestaurant == undefined) ? (
                <section id="restaurant-list" className="bg-white p-6 rounded-3xl shadow-lg border border-gray-100">
                    <h2 className="text-2xl font-bold text-gray-800 mb-4">Restaurants near you</h2>
                    <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-6 items-stretch">
                        {renderRestaurants({ restaurants })}
                    </div>
                </section>
            ) : (
                <section id="food-menu" className="bg-gray-50 p-6 rounded-3xl shadow-lg border border-gray-100">
                    <div className="flex items-center mb-4">
                        <button onClick={() => setCurrentRestaurant(undefined)} className="text-gray-500 hover:text-gray-700 transition">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
                            </svg>
                        </button>
                        <h2 id="menu-title" className="text-2xl font-bold text-gray-800 ml-2">
                            {currentRestaurant && currentRestaurant.restaurantName}
                        </h2>
                    </div>
                    <div id="menu-container" className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 rounded-2xl">
                        {(currentRestaurant != undefined) && renderMenu(currentRestaurant.menu)}
                    </div>
                </section>
            )}
        </div>
    );
}



//     let restaurants = cachedData.restaurants;
//     restaurantsContainer.innerHTML = '';

//         const url =
//             `${serverUrl}/api/deliveries/fee` +
//             `?customer_id=${customerId}` +
//             `&customer_address_id=${customerAddrId}` +
//             `&restaurant_id=${restaurantId}`;
//         const res = await fetch(url, {
//             method: "GET",
//             headers: {
//                 "Accept": "application/json",
//                 "Authorization": `Bearer ${token}`
//             }
//         });
//         const { fee = 10 } = await res.json(); // Default to 10 instead of 0
//         deliveryFees[restaurantId] = fee;

//         // Reverse calculate distance from fee with some randomness
//         // Formula: distance = (fee - 10) / 1.6
//         const baseDistance = (fee - 10) / 1.6;
//         const randomVariation = (Math.random() * 2) - 1; // ±1 km variation
//         const distance = Math.max(0, Math.min(25, baseDistance + randomVariation)).toFixed(1);

//         // Calculate fake ETA based on distance (1-60 minutes)
//         const distanceNum = parseFloat(distance);
//         const baseEta = Math.round((distanceNum / 30) * 60); // Convert to minutes
//         const etaVariation = Math.floor(Math.random() * 10) - 5; // ±5 minutes
//         const eta = Math.max(1, Math.min(60, baseEta + etaVariation)); // Clamp between 1-60

//         let imageUrl = '';
//         try {
//             imageUrl = await getDownloadURL(
//                 ref(storage, restaurant.imageInfo.url)
//             );
//         } catch (err) {
//             console.error(err);
//         }




