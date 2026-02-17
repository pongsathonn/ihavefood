import Image from "next/image";

export default function FakeTracking() {
    return (
        // Order Tracking Section(Initially Hidden
        <section id="order-tracking-section"
            className="bg-white p-6 rounded-3xl shadow-lg border border-gray-100 col-span-1 md:col-span-3">
            <div className="flex items-center mb-4">
                <button id="tracking-back-button" className="text-gray-500 hover:text-gray-700 transition">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
                    </svg>
                </button>
                <h2 className="text-2xl font-bold text-gray-800 ml-2">Order Tracking</h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-6">
                {/* Left Column: Map and Status  */}
                <div className="space-y-6">
                    {/* Order Status  */}
                    <div className="bg-gray-50 p-6 rounded-2xl border border-gray-200 shadow-sm">
                        <h3 className="text-xl font-bold text-gray-800 mb-4">Order Status</h3>
                        <div className="flex items-center justify-between relative">
                            {/* Step 1  */}
                            <div className="tracking-step w-1/4 flex flex-col items-center relative active" id="step-1">
                                <div
                                    className="step-dot w-6 h-6 rounded-full bg-pink-500 border-4 border-white transition-all duration-[3000ms]">
                                </div>
                                <span
                                    className="step-text text-xs text-center text-gray-600 mt-2 transition-all duration-[3000ms]">Order
                                    Placed</span>
                            </div>

                            {/* Step 2  */}
                            <div className="tracking-step w-1/4 flex flex-col items-center relative" id="step-2">
                                <div
                                    className="step-dot w-6 h-6 rounded-full bg-gray-300 border-4 border-white transition-all duration-[3000ms]">
                                </div>
                                <span
                                    className="step-text text-xs text-center text-gray-600 mt-2 transition-all duration-[3000ms]">Preparing
                                    Food</span>
                            </div>
                            {/* Line1 */}
                            <div
                                className="step-line absolute h-1 bg-gray-300 top-3 left-[12.5%] right-[62.5%] transition-all duration-[3000ms]">
                            </div>

                            {/* Step 3 */}
                            <div className="tracking-step w-1/4 flex flex-col items-center relative" id="step-3">
                                <div
                                    className="step-dot w-6 h-6 rounded-full bg-gray-300 border-4 border-white transition-all duration-[3000ms]">
                                </div>
                                <span
                                    className="step-text text-xs text-center text-gray-600 mt-2 transition-all duration-[3000ms]">Rider
                                    on the Way</span>
                            </div>
                            {/* Line 2 */}
                            <div
                                className="step-line absolute h-1 bg-gray-300 top-3 left-[37.5%] right-[37.5%] transition-all duration-[3000ms]">
                            </div>


                            {/* Step 4 */}
                            <div className="tracking-step w-1/4 flex flex-col items-center relative" id="step-4">
                                <div
                                    className="step-dot w-6 h-6 rounded-full bg-gray-300 border-4 border-white transition-all duration-[3000ms]">
                                </div>
                                <span
                                    className="step-text text-xs text-center text-gray-600 mt-2 transition-all duration-[3000ms]">Delivered</span>
                            </div>
                            {/* Line 3  */}
                            <div
                                className="step-line absolute h-1 bg-gray-300 top-3 left-[62.5%] right-[12.5%] transition-all duration-[3000ms]">
                            </div>

                        </div>

                    </div>

                    {/* Map Placeholder */}
                    <div className="bg-gray-100 rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
                        <iframe
                            src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3777.8896428156843!2d98.99113197587258!3d18.78778006195882!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x30da3a7e90bb6f5d%3A0x8e4e6c5c6de1f11d!2sTha%20Phae%20Gate!5e0!3m2!1sen!2sth!4v1734000000000!5m2!1sen!2sth"
                            width="100%" height="400" style={{}} loading="lazy"
                            className="rounded-2xl">
                        </iframe>
                        <div className="p-4 bg-gray-50 text-center">
                            <p className="text-sm text-red-600">**Map feature not implemented yet</p>
                        </div>
                    </div>

                </div>

                {/* Right Column: Order & Rider Details */}
                <div className="space-y-6">
                    {/* Order Summary */}
                    <div className="bg-gray-50 p-6 rounded-2xl border border-gray-200 shadow-sm">
                        <h3 className="text-xl font-bold text-gray-800 mb-4">Order Details</h3>
                        <div className="space-y-2 text-gray-600 text-sm">
                            <div className="flex justify-between">
                                <span className="font-semibold">Order ID:</span>
                                <span id="tracking-order-id"></span>
                            </div>
                            <div className="flex justify-between">
                                <span className="font-semibold">Restaurant:</span>
                                <span id="tracking-restaurant-name"></span>
                            </div>

                            <div className="flex justify-between items-start">
                                <span className="font-semibold">Foods:</span>
                                <div id="tracking-food-items" className="flex flex-col items-end space-y-1 w-40"></div>
                            </div>
                            <div className="flex justify-between">
                                <span className="font-semibold">Total:</span>
                                <span id="tracking-order-total"></span>
                            </div>

                            {/* <div className="flex justify-between"> 
                                <span className="font-semibold">Est. Delivery:</span> 
                                <span id="tracking-eta">15-20 mins</span> 
                            </div> --> */}

                        </div>
                    </div>

                    Rider Information
                    <div className="bg-gray-50 p-6 rounded-2xl border border-gray-200 shadow-sm">
                        <h3 className="text-xl font-bold text-gray-800 mb-4">Rider Information</h3>
                        <div className="flex items-center space-x-4">

                            {/* <Image
                                alt="Rider Profile"
                                className="rounded-full object-cover"
                                width={20}
                                height={20}
                            /> */}

                            <div>
                                <p className="text-lg font-semibold text-gray-800" id="rider-name">Demo Rider</p>
                                <p className="text-sm text-gray-500" id="rider-vehicle">Honda Click 125i (กข1234)</p>
                                <a href="tel:0812345678"
                                    className="text-blue-500 hover:underline text-sm flex items-center mt-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none"
                                        viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                                            d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                                    </svg>
                                    Call Rider
                                </a>
                            </div>
                        </div>
                    </div>

                    {/* Problem Reporting */}
                    <div className="bg-gray-50 p-6 rounded-2xl border border-gray-200 shadow-sm text-center">
                        <p className="text-sm text-gray-600 mb-3">Having an issue with your order?</p>
                        <button
                            className="bg-red-500 text-white font-bold py-2 px-6 rounded-full hover:bg-red-600 transition">
                            Report a Problem
                        </button>
                    </div>
                </div>
            </div>
        </section>
    )
}
