// Utility functions and shared state

const API_BASE = '/api';
let stations = [];

// Initialize the application
document.addEventListener('DOMContentLoaded', async () => {
    console.log('Initializing Train Booking System...');

    // Set minimum date to today
    const dateInput = document.getElementById('date');
    if (dateInput) {
        const today = new Date().toISOString().split('T')[0];
        dateInput.min = today;

        // Set default to tomorrow
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        dateInput.value = tomorrow.toISOString().split('T')[0];
    }

    // Load stations
    await loadStations();

    // Setup modal handlers
    setupModals();
});

// Load stations from API
async function loadStations() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/stations`);

        if (!response.ok) {
            throw new Error('Failed to load stations');
        }

        stations = await response.json();
        console.log(`Loaded ${stations.length} stations`);

        // Populate dropdowns
        populateStationDropdowns();

    } catch (error) {
        console.error('Error loading stations:', error);
        showAlert('Failed to load stations. Please refresh the page.', 'error');
    } finally {
        showLoading(false);
    }
}

// Populate station dropdowns
function populateStationDropdowns() {
    const originSelect = document.getElementById('origin');
    const destinationSelect = document.getElementById('destination');

    if (!originSelect || !destinationSelect) return;

    // Clear existing options except the first one
    originSelect.innerHTML = '<option value="">Select departure station</option>';
    destinationSelect.innerHTML = '<option value="">Select arrival station</option>';

    // Add station options
    stations.forEach(station => {
        const optionText = `${station.name} (${station.code})`;

        const originOption = document.createElement('option');
        originOption.value = station.code;
        originOption.textContent = optionText;
        originSelect.appendChild(originOption);

        const destOption = document.createElement('option');
        destOption.value = station.code;
        destOption.textContent = optionText;
        destinationSelect.appendChild(destOption);
    });
}

// Show/hide loading overlay
function showLoading(show) {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        if (show) {
            overlay.classList.add('show');
        } else {
            overlay.classList.remove('show');
        }
    }
}

// Show alert message
function showAlert(message, type = 'info') {
    // For now, use console and browser alert
    // In production, implement a better alert system
    console.log(`[${type.toUpperCase()}] ${message}`);
    if (type === 'error') {
        alert(message);
    }
}

// Format currency
function formatCurrency(amount) {
    return `€${parseFloat(amount).toFixed(2)}`;
}

// Format time
function formatTime(timeString) {
    if (!timeString) return '';
    return timeString.substring(0, 5); // HH:MM
}

// Format date
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        weekday: 'short',
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

// Setup modal handlers
function setupModals() {
    // Booking modal
    const bookingModal = document.getElementById('bookingModal');
    const confirmationModal = document.getElementById('confirmationModal');

    // Close buttons
    document.querySelectorAll('.modal-close').forEach(closeBtn => {
        closeBtn.addEventListener('click', () => {
            closeAllModals();
        });
    });

    // Cancel booking button
    const cancelBtn = document.getElementById('cancelBooking');
    if (cancelBtn) {
        cancelBtn.addEventListener('click', () => {
            closeAllModals();
        });
    }

    // Close confirmation button
    const closeConfirmBtn = document.getElementById('closeConfirmation');
    if (closeConfirmBtn) {
        closeConfirmBtn.addEventListener('click', () => {
            closeAllModals();
        });
    }

    // Click outside to close
    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            closeAllModals();
        }
    });
}

// Show modal
function showModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('show');
    }
}

// Close all modals
function closeAllModals() {
    document.querySelectorAll('.modal').forEach(modal => {
        modal.classList.remove('show');
    });
}

// Create booking form for passengers
function createPassengerForm(adults = 1, seniors = 0, children = 0) {
    const container = document.getElementById('passengerInputs');
    if (!container) return;

    container.innerHTML = '';

    let passengerIndex = 1;

    // Add adult fields
    for (let i = 0; i < adults; i++) {
        container.appendChild(createPassengerField('Adult', 'adult', passengerIndex++));
    }

    // Add senior fields
    for (let i = 0; i < seniors; i++) {
        container.appendChild(createPassengerField('Senior', 'senior', passengerIndex++));
    }

    // Add children fields
    for (let i = 0; i < children; i++) {
        container.appendChild(createPassengerField('Child', 'child', passengerIndex++));
    }
}

// Create single passenger input field
function createPassengerField(label, type, index) {
    const div = document.createElement('div');
    div.className = 'form-group';

    const labelElem = document.createElement('label');
    labelElem.textContent = `${label} ${index} - Full Name:`;

    const input = document.createElement('input');
    input.type = 'text';
    input.name = `passenger_${type}_${index}`;
    input.className = 'passenger-input';
    input.dataset.type = type;
    input.required = true;
    input.placeholder = 'Enter full name';

    div.appendChild(labelElem);
    div.appendChild(input);

    return div;
}

// Get passengers from form
function getPassengersFromForm() {
    const inputs = document.querySelectorAll('.passenger-input');
    const passengers = [];

    inputs.forEach(input => {
        if (input.value.trim()) {
            passengers.push({
                name: input.value.trim(),
                passenger_type: input.dataset.type
            });
        }
    });

    return passengers;
}

// Show booking confirmation
function showBookingConfirmation(booking) {
    const container = document.getElementById('confirmationDetails');
    if (!container) return;

    const schedule = booking.schedule;
    const train = schedule.train;
    const origin = schedule.origin;
    const destination = schedule.destination;

    container.innerHTML = `
        <div class="booking-confirmation">
            <p><strong>Booking Reference:</strong></p>
            <p class="booking-ref">${booking.booking_ref}</p>

            <p><strong>Train:</strong> ${train.number} (${train.type})</p>
            <p><strong>Route:</strong> ${origin.name} → ${destination.name}</p>
            <p><strong>Date:</strong> ${formatDate(booking.booking_date)}</p>
            <p><strong>Departure:</strong> ${formatTime(schedule.departure_time)}</p>
            <p><strong>Arrival:</strong> ${formatTime(schedule.arrival_time)}</p>

            <p><strong>Passengers:</strong></p>
            <ul>
                ${booking.passengers.map(p => `
                    <li>${p.name} (${p.passenger_type}) - ${p.seat_number || 'No seat'} - ${formatCurrency(p.price)}</li>
                `).join('')}
            </ul>

            <p style="font-size: 1.25rem; margin-top: 15px;">
                <strong>Total Price:</strong>
                <span style="color: var(--success-color);">${formatCurrency(booking.total_price)}</span>
            </p>

            <div class="alert alert-success" style="margin-top: 20px;">
                <p><strong>Important:</strong> Please save your booking reference: <strong>${booking.booking_ref}</strong></p>
                <p>You will need it to check your booking or make changes.</p>
            </div>
        </div>
    `;

    showModal('confirmationModal');
}

// Generate UUID for chat session
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

// Export functions to global scope
window.app = {
    API_BASE,
    stations,
    showLoading,
    showAlert,
    formatCurrency,
    formatTime,
    formatDate,
    showModal,
    closeAllModals,
    createPassengerForm,
    getPassengersFromForm,
    showBookingConfirmation,
    generateUUID
};
