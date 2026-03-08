export function ConfirmModal() {
  return (
    // Order Confirmation Modal
    <div
      id="modal"
      className="fixed inset-0 bg-gray-900 bg-opacity-50 hidden flex items-center justify-center p-4 z-50"
    >
      <div className="bg-white p-8 rounded-3xl shadow-lg w-full max-w-sm text-center">
        <svg
          className="mx-auto h-16 w-16 text-pink-500 mb-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <h3 className="text-xl font-bold text-gray-800 mb-2">Order Placed!</h3>
        <p className="text-gray-600 mb-4">
          Your food is on its way. Thank you for your order!
        </p>
        <button
          id="track-order-button"
          className="bg-pink-500 hover:bg-pink-600 text-white font-bold py-2 px-4 rounded-full transition-colors duration-200 mt-2 mr-2"
        >
          Track Order
        </button>
        <button
          id="close-modal-button"
          className="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-full transition-colors duration-200"
        >
          Close
        </button>
      </div>
    </div>
  )
}

export function RegisterSuccessModal() {
  return (
    // Register Success Modal
    <div
      id="modal-register"
      className="fixed inset-0 bg-gray-900 bg-opacity-50 hidden flex items-center justify-center p-4 z-50"
    >
      <div className="relative bg-white p-8 rounded-3xl shadow-lg w-full max-w-sm text-center">
        {/* Close button  */}
        <button
          onclick="document.getElementById('modal-register').classNameList.add('hidden')"
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
          aria-label="Close"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>

        <svg
          className="mx-auto h-16 w-16 text-green-500 mb-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>

        <h3 className="text-xl font-bold text-gray-800 mb-2">
          Successfully registered!
        </h3>
      </div>
    </div>
  )
}
