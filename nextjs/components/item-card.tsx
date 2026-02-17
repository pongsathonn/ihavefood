
export default function ItemCard() {
    return (
        <>

        </>
    )
}


////////////
// const renderMenu = async (menu) => {
//     menuContainer.innerHTML = '';
//     if (!menu || menu.length === 0) {
//         menuContainer.innerHTML = '<p class="text-gray-500 italic">This restaurant has no menu items.</p>';
//         return;
//     }

//     for (const item of menu) {

//         const storage = getStorage();
//         let imageUrl = '';
//         try {
//             imageUrl = await getDownloadURL(
//                 ref(storage, item.imageInfo.url)
//             );
//         } catch (err) {
//             console.error(err);
//         }
//         const card = document.createElement('div');
//         card.className = 'food-item-card ...';

//         card.innerHTML = `
//         <div class="flex items-center space-x-4">
//             <img src="${imageUrl}" alt="${item.foodName}" class="w-16 h-16 object-cover rounded-lg">
//             <div>
//                 <h4 class="text-md font-semibold text-gray-800">${item.foodName}</h4>
//                 <span class="text-sm text-gray-500">฿${(item.price).toFixed(2)}</span>
//             </div>
//         </div>
//         <button class="add-to-cart-btn ..." data-item='${JSON.stringify(item)}'>Add</button>
//         `;
//         menuContainer.appendChild(card);
//     }
// };