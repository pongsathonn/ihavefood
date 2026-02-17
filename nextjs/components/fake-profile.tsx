
export default function ProfileMenu() {
    return (
        < section id="user-profile" className="max-w-xl mx-auto bg-white rounded-3xl shadow-lg p-8">
            {/* Profile Picture */}
            <div className="flex flex-col items-center mb-8">
                <div className="w-24 h-24 rounded-full overflow-hidden mb-3 border-2 border-gray-200">
                    <img id="user-picture" src="https://placehold.co/128x128/94a3b8/fff?text=C"
                        alt="User Profile Picture" className="w-full h-full object-cover" />
                </div>
                <p id="customer-name" className="text-xl font-semibold text-gray-900">
                    FooCustomer
                </p>
            </div>

            {/* Customer Info */}
            <div className="space-y-6 text-sm">
                <div>

                    <div className="space-y-4">
                        <div className="flex justify-between items-center">
                            <p className="font-semibold text-gray-900">Info</p>
                            <button className="text-blue-600 text-xs hover:underline">Edit</button>
                        </div>

                        <div className="space-y-2">
                            <div className="flex items-center">
                                <span className="text-gray-500 w-24">CustomerID</span>
                                <span id="customer-id" className="text-gray-900">TODO</span>
                            </div>
                            <div className="flex items-center">
                                <span className="text-gray-500 w-24">Email</span>
                                <span id="customer-email" className="text-gray-900">TODO</span>
                            </div>

                            <div className="flex items-center gap-3">
                                <span className="text-gray-500 w-24">Phone</span>

                                <input type="text" id="customer-phone"
                                    value="TODO"
                                    className="flex-1 rounded-lg border border-gray-300 bg-gray-50 px-3 py-2 text-gray-900 transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 outline-none"
                                >
                                </input>
                            </div>
                        </div>

                        <div className="flex space-x-2 justify-end">
                            <button
                                className="px-3 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded hover:bg-gray-200">Cancel</button>
                            <button
                                className="px-3 py-1 text-xs font-medium text-white bg-blue-600 rounded hover:bg-blue-700">Save</button>
                        </div>
                    </div>

                    {/* Divider */}
                    <div className="border-t border-gray-200"></div>

                    {/* Social */}
                    <div className="space-y-2">
                        <p className="font-semibold text-gray-900 mb-3">Social</p>
                        <div className="flex">
                            <span className="text-gray-500 w-24">Facebook</span>
                            <span id="social-facebook" className="text-gray-900">TODO</span>
                        </div>
                        <div className="flex">
                            <span className="text-gray-500 w-24">Instagram</span>
                            <span id="social-instagram" className="text-gray-900">TODO</span>
                        </div>
                        <div className="flex">
                            <span className="text-gray-500 w-24">LINE</span>
                            <span id="social-line" className="text-gray-900">TODO</span>
                        </div>
                    </div>

                    {/* Divider */}
                    <div className="border-t border-gray-200"></div>
                </div>

                <div className="max-w-md mx-auto">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="font-semibold text-gray-900 text-lg">My Address</h3>
                        <button id="add-btn" className="text-sm font-medium text-blue-600 hover:text-blue-700">+ Add
                            New</button>
                    </div>
                    <div id="address-list" className="space-y-3"></div>
                </div>

                {/* Sign Out */}
                <button id="signout-btn"
                    className="mt-8 w-full bg-red-500 hover:bg-red-600 text-white font-medium py-3 rounded transition">
                    Sign Out
                </button>
            </div>
        </section>
    )
}


//////////////////////////////////////////////////////
// // TODO: render default address first
// const renderAddresses = async () => {

//     isAddingNew = false;
//     addressList.innerHTML = '';
//     const sortedAddresses = [...addresses].sort((a, b) => b.isDefault - a.isDefault);

//     sortedAddresses.forEach(addr => {
//         const card = document.createElement('div');
//         card.className = 'address-card relative p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors';
//         card.innerHTML = `
//         <div class="flex justify-between items-start mb-2">
//             <div>
//                 <span class="font-bold text-gray-900">${addr.addressName}</span>
//                 ${addr.isDefault ? '<span class="ml-2 px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded-full">Default</span>' : ''}
//             </div>
//             <div class="flex space-x-3">
//                 <button data-edit="${addr.addressId}" class="text-gray-400 hover:text-blue-600"><small>Edit</small></button>
//                 <button data-delete="${addr.addressId}" class="text-gray-400 hover:text-red-600"><small>Delete</small></button>
//             </div>
//         </div>
//         <div class="text-gray-600 text-sm leading-relaxed">
//             <p>${addr.subDistrict}, ${addr.district}</p>
//             <p>${addr.province}, <span class="font-medium">${addr.postalCode}</span></p>
//         </div>
//         `;
//         addressList.appendChild(card);
//     });
// }

// async function handleSaveAddress(id, isNew, event) {
//     const addressLine = document.getElementById(`name-${id}`);
//     const subDistrict = document.getElementById(`sub-${id}`);
//     const district = document.getElementById(`district-${id}`);
//     const province = document.getElementById(`province-${id}`);
//     const postal = document.getElementById(`postal-${id}`);

//     // Reset previous error styles
//     [addressLine, subDistrict, district, province, postal].forEach(el => el.classList.remove('border-red-500'));

//     let hasError = false;

//     if (!addressLine.value) { addressLine.classList.add('border-red-500'); hasError = true; }
//     if (!subDistrict.value) { subDistrict.classList.add('border-red-500'); hasError = true; }
//     if (!district.value) { district.classList.add('border-red-500'); hasError = true; }
//     if (!province.value) { province.classList.add('border-red-500'); hasError = true; }
//     if (!postal.value) { postal.classList.add('border-red-500'); hasError = true; }

//     if (hasError) return; // stop saving

//     const button = event.target;
//     button.disabled = true;
//     button.textContent = 'Saving...';

//     try {
//         await saveAddress(id, isNew);
//     } catch (error) {
//         console.error('Error saving address:', error);
//         alert('Failed to save address. Please try again.');
//     } finally {
//         button.disabled = false;
//         button.textContent = isNew ? 'Add' : 'Save';
//     }
// }

// async function saveAddress(addrID, isNew = false) {
//     isAddingNew = false;

//     const customerID = sessionStorage.getItem('customer_id');

//     const addressLine = document.getElementById(`name-${addrID}`).value;
//     const subDistrict = document.getElementById(`sub-${addrID}`).value;
//     const district = document.getElementById(`district-${addrID}`).value;
//     const province = document.getElementById(`province-${addrID}`).value;
//     const postal = document.getElementById(`postal-${addrID}`).value;
//     let isDefault = document.getElementById(`default-${addrID}`).checked;

//     if (isNew) {
//         const res = await createAddress({
//             customer_id: customerID,
//             address: {
//                 address_name: addressLine,
//                 sub_district: subDistrict,
//                 district: district,
//                 province: province,
//                 postal_code: postal
//             }
//         });

//         if (isDefault) {
//             addresses.forEach(a => a.isDefault = false);
//         }

//         addresses.push({
//             addressId: res.addressId,
//             addressName: res.addressName,
//             subDistrict: res.subDistrict,
//             district: res.district,
//             province: res.province,
//             postalCode: res.postalCode,
//             isDefault
//         });

//     } else {

//         const res = await updateAddress({
//             customer_id: customerID,
//             address_id: addrID,
//             address: {
//                 address_name: addressLine,
//                 sub_district: subDistrict,
//                 district: district,
//                 province: province,
//                 postal_code: postal
//             }
//         });

//         if (isDefault) {
//             addresses.forEach(a => a.isDefault = false);
//         }

//         const addressToUpdate = addresses.find(a => a.addressId === addrID);
//         if (addressToUpdate) {
//             Object.assign(addressToUpdate, {
//                 addressName: res.addressName,
//                 subDistrict: res.subDistrict,
//                 district: res.district,
//                 province: res.province,
//                 postalCode: res.postalCode,
//                 isDefault
//             });
//         }
//     }

//     if (addresses.length === 1) {
//         addresses[0].isDefault = true;
//         renderRestaurants();
//     }

//     renderAddresses();
// }

// // editing form used by both '+ Add New' and 'Edit' button
// function editAddress(id, isNew = false) {

//     addressList.innerHTML = '';
//     const addr = isNew ? {} : addresses.find(a => a.addressId === id) || {};

//     const card = document.createElement('div');
//     card.className = 'address-card relative p-4 border border-blue-300 bg-blue-50 rounded-lg';

//     card.innerHTML = `
//         ${createInput('Address Line', `name-${id}`, addr.addressName || '', 'text', 'House no. / Moo / Soi / Road')}
//         <div class="grid grid-cols-2 gap-3">
//             ${createInput('Sub-district', `sub-${id}`, addr.subDistrict || '', 'text', 'e.g., Suthep')}
//             ${createInput('District', `district-${id}`, addr.district || '', 'text', 'e.g., Mueang Chiang Mai')}
//         </div>
//         <div class="grid grid-cols-2 gap-3">
//             ${createInput('Province', `province-${id}`, addr.province || '', 'text', 'e.g., Chiang Mai')}
//             ${createInput('Postal Code', `postal-${id}`, addr.postalCode || '', 'text', 'e.g., 50200')}
//         </div>
        
//         <div class="flex items-center">
//             <input type="checkbox" id="default-${id}" ${addr.isDefault ? 'checked' : ''} class="w-4 h-4 text-blue-600 rounded focus:ring-2 focus:ring-blue-500">
//             <label for="default-${id}" class="ml-2 text-sm text-gray-700">Set as default</label>
//         </div>
        
//         <div class="flex space-x-2 pt-2">
//             <button onclick="handleSaveAddress('${id}', ${isNew}, event)" class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">${isNew ? 'Add' : 'Save'}</button>
//             <button onclick="renderAddresses()" class="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors">Cancel</button>
//         </div>
//         `;
//     addressList.appendChild(card);
// }

// // Event delegation
// addressList.addEventListener('click', async (e) => {
//     const editBtn = e.target.closest('button[data-edit]');
//     const deleteBtn = e.target.closest('button[data-delete]');

//     if (editBtn) {
//         const card = editBtn.closest('.address-card');
//         editAddress(editBtn.dataset.edit, false);
//         return;
//     }

//     if (deleteBtn && confirm('Delete this address?')) {
//         const customerId = sessionStorage.getItem('customer_id');
//         const token = sessionStorage.getItem("token");
//         const addressId = deleteBtn.dataset.delete;

//         try {
//             const res = await fetch(`${serverUrl}/api/customers/${customerId}/addresses/${addressId}`, {
//                 method: "DELETE",
//                 headers: {
//                     "Authorization": `Bearer ${token}`,
//                     "Content-Type": "application/json"
//                 }
//             });

//             if (!res.ok) {
//                 throw new Error('Failed to delete');
//             }

//             addresses = addresses.filter(a => a.addressId !== addressId);
//             renderAddresses();
//         } catch (err) {
//             alert('Could not delete address. Please try again.');
//             console.error(err);
//         }
//     }

// });


// let isAddingNew = false;
// addBtn.addEventListener('click', () => {

//     const MAX_ADDR = 5;
//     if (addresses.length >= MAX_ADDR) {
//         alert(`You’ve reached the maximum saved addresses.`);
//         return;
//     }

//     // prevent multiple forms
//     if (isAddingNew) return;
//     isAddingNew = true;

//     // If a new address form already exists, do nothing
//     if (document.querySelector('.address-card.new-address')) return;

//     editAddress(null, true);
// });