// Traditional booking form handling

let currentSearchResults = [];
let selectedSchedule = null;

// Initialize traditional booking
document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('searchForm');
    if (searchForm) {
        searchForm.addEventListener('submit', handleSearch);
    }

    const bookingForm = document.getElementById('bookingForm');
    if (bookingForm) {
        bookingForm.addEventListener('submit', handleBookingSubmit);
    }
});

// Handle search form submission
async function handleSearch(e) {
    e.preventDefault();

    const formData = new FormData(e.target);

    // Get passenger counts
    const adults = parseInt(formData.get('adults')) || 0;
    const seniors = parseInt(formData.get('seniors')) || 0;
    const children = parseInt(formData.get('children')) || 0;
    const passengerCount = adults + seniors + children;

    if (passengerCount === 0) {
        window.app.showAlert('Please select at least one passenger', 'error');
        return;
    }

    // Build search request
    const searchRequest = {
        origin: formData.get('origin'),
        destination: formData.get('destination'),
        date: formData.get('date'),
        time_preference: formData.get('time_preference') || 'any',
        passenger_count: passengerCount,
        filters: {
            has_wifi: document.getElementById('hasWifi').checked,
            has_food: document.getElementById('hasFood').checked
        }
    };

    console.log('Search request:', searchRequest);

    try {
        window.app.showLoading(true);

        const response = await fetch(`${window.app.API_BASE}/search`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(searchRequest)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Search failed');
        }

        const results = await response.json();
        currentSearchResults = results;

        console.log(`Found ${results.length} trains`);
        displaySearchResults(results, { adults, seniors, children });

    } catch (error) {
        console.error('Search error:', error);
        window.app.showAlert(error.message, 'error');
        displaySearchResults([]);
    } finally {
        window.app.showLoading(false);
    }
}

// Display search results
function displaySearchResults(results, passengerInfo = {}) {
    const container = document.getElementById('searchResults');
    if (!container) return;

    if (results.length === 0) {
        container.innerHTML = `
            <div class="alert alert-info">
                <p>No trains found matching your criteria. Please try different search parameters.</p>
            </div>
        `;
        return;
    }

    container.innerHTML = `
        <h3 style="margin-bottom: 20px;">Found ${results.length} trains</h3>
        ${results.map((result, index) => createTrainCard(result, index, passengerInfo)).join('')}
    `;

    // Add click handlers for booking buttons
    results.forEach((result, index) => {
        const btn = document.getElementById(`book-btn-${index}`);
        if (btn) {
            btn.addEventListener('click', () => {
                initiateBooking(result.schedule, passengerInfo);
            });
        }
    });
}

// Create train card HTML
function createTrainCard(result, index, passengerInfo) {
    const schedule = result.schedule;
    const train = schedule.train;
    const origin = schedule.origin;
    const destination = schedule.destination;

    return `
        <div class="train-card">
            <div class="train-header">
                <div class="train-number">${train.number}</div>
                <div class="train-type">${train.type}</div>
            </div>

            <div class="train-details">
                <div class="detail-item">
                    <span class="detail-label">From:</span> ${origin.name}
                </div>
                <div class="detail-item">
                    <span class="detail-label">To:</span> ${destination.name}
                </div>
                <div class="detail-item">
                    <span class="detail-label">Departure:</span> ${result.departure_time}
                </div>
                <div class="detail-item">
                    <span class="detail-label">Arrival:</span> ${result.arrival_time}
                </div>
                <div class="detail-item">
                    <span class="detail-label">Duration:</span> ${result.duration}
                </div>
                <div class="detail-item">
                    <span class="detail-label">Available Seats:</span> ${schedule.available_seats}
                </div>
            </div>

            <div class="train-amenities">
                <div class="amenity">
                    ${train.has_wifi ? '✅' : '❌'} WiFi
                </div>
                <div class="amenity">
                    ${train.has_food ? '✅' : '❌'} Food Service
                </div>
            </div>

            <div class="train-footer">
                <div>
                    <div class="price">${window.app.formatCurrency(result.total_price)}</div>
                    <div style="font-size: 0.875rem; color: var(--text-gray);">
                        ${window.app.formatCurrency(result.price_per_person)} per person
                    </div>
                </div>
                <button id="book-btn-${index}" class="btn btn-success">Book Now</button>
            </div>
        </div>
    `;
}

// Initiate booking process
function initiateBooking(schedule, passengerInfo) {
    selectedSchedule = schedule;

    console.log('Initiating booking for schedule:', schedule.id);

    // Create passenger form
    window.app.createPassengerForm(
        passengerInfo.adults || 0,
        passengerInfo.seniors || 0,
        passengerInfo.children || 0
    );

    // Show booking modal
    window.app.showModal('bookingModal');
}

// Handle booking form submission
async function handleBookingSubmit(e) {
    e.preventDefault();

    // Check if we're in AI booking mode
    if (window.isAIBookingMode) {
        // Delegate to AI booking handler
        return handleAIBookingSubmitFromTraditional(e);
    }

    if (!selectedSchedule) {
        window.app.showAlert('No schedule selected', 'error');
        return;
    }

    const passengers = window.app.getPassengersFromForm();

    if (passengers.length === 0) {
        window.app.showAlert('Please enter passenger names', 'error');
        return;
    }

    const bookingRequest = {
        schedule_id: selectedSchedule.id,
        date: document.getElementById('date').value,
        passengers: passengers
    };

    console.log('Booking request:', bookingRequest);

    try {
        window.app.showLoading(true);
        window.app.closeAllModals();

        const response = await fetch(`${window.app.API_BASE}/bookings`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(bookingRequest)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Booking failed');
        }

        const result = await response.json();

        if (!result.success) {
            throw new Error(result.message);
        }

        console.log('Booking successful:', result.booking);

        // Show confirmation
        window.app.showBookingConfirmation(result.booking);

        // Clear search results
        currentSearchResults = [];
        selectedSchedule = null;
        displaySearchResults([]);

        // Reset form
        document.getElementById('searchForm').reset();

    } catch (error) {
        console.error('Booking error:', error);
        window.app.showAlert(error.message, 'error');
    } finally {
        window.app.showLoading(false);
    }
}

// Handle AI booking submission from traditional form
async function handleAIBookingSubmitFromTraditional(e) {
    e.preventDefault();

    console.log('AI Booking: Form submitted');

    const count = parseInt(document.getElementById('aiPassengerCount')?.value) || 1;
    const date = document.getElementById('aiBookingDate')?.value;
    const context = window.aiBookingContext;

    console.log('AI Booking: Count:', count, 'Date:', date, 'Context:', context);

    if (!context) {
        console.error('AI Booking: No context found!');
        if (window.addMessageToChat) {
            window.addMessageToChat('assistant', 'Booking context not found. Please try selecting a train again.');
        }
        return;
    }

    if (!date) {
        console.error('AI Booking: No date selected!');
        if (window.addMessageToChat) {
            window.addMessageToChat('assistant', 'Please select a travel date.');
        }
        return;
    }

    // Collect passenger data
    const passengers = [];
    for (let i = 0; i < count; i++) {
        const name = document.getElementById(`aiPassengerName${i}`)?.value;
        const type = document.getElementById(`aiPassengerType${i}`)?.value;

        if (!name || !name.trim()) {
            console.error(`AI Booking: Passenger ${i+1} name is empty!`);
            if (window.addMessageToChat) {
                window.addMessageToChat('assistant', `Please enter name for passenger ${i+1}.`);
            }
            return;
        }

        passengers.push({
            name: name.trim(),
            passenger_type: type || 'adult'
        });
    }

    console.log('AI Booking: Collected passengers:', passengers);

    // Close modal
    document.getElementById('bookingModal').classList.remove('show');
    window.isAIBookingMode = false;

    // Show loading
    window.app.showLoading();

    try {
        console.log('AI Booking: Starting booking process');
        console.log('AI Booking: Context:', context);
        console.log('AI Booking: Passengers:', passengers);
        console.log('AI Booking: Date:', date);

        // Call booking API directly (not through AI)
        const response = await fetch(`${window.app.API_BASE}/bookings`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                schedule_id: parseInt(context.scheduleId),
                date: date,
                passengers: passengers
            })
        });

        console.log('AI Booking: Response status:', response.status);

        if (!response.ok) {
            const error = await response.json();
            console.error('AI Booking: Response not OK:', error);
            throw new Error(error.message || 'Booking failed');
        }

        const result = await response.json();
        console.log('AI Booking: Full result:', result);

        if (!result.success) {
            console.error('AI Booking: Success=false:', result.message);
            throw new Error(result.message);
        }

        const booking = result.booking;
        console.log('AI Booking: Extracted booking:', booking);

        window.app.showLoading(false);

        // Show confirmation using the same modal as traditional booking
        console.log('AI Booking: Calling showBookingConfirmation with:', booking);
        window.app.showBookingConfirmation(booking);

        // Also add a message to the AI chat
        if (window.addMessageToChat) {
            window.addMessageToChat('user', `Booked ${count} ticket(s) for ${date} from ${context.origin} to ${context.destination}`);
            window.addMessageToChat('assistant', `✅ Your booking has been confirmed! Your booking reference is: ${booking.booking_ref}`);
        }

    } catch (error) {
        console.error('AI Booking error:', error);
        window.app.showLoading(false);
        if (window.addMessageToChat) {
            window.addMessageToChat('assistant', `Sorry, there was an error creating your booking: ${error.message}`);
        }
    }
}
