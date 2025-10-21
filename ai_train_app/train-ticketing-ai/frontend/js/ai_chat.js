// AI Chat interface handling

let chatSessionId = null;
let isProcessing = false;

// Initialize AI chat
document.addEventListener('DOMContentLoaded', () => {
    // Generate session ID
    chatSessionId = window.app.generateUUID();
    console.log('AI Chat session ID:', chatSessionId);

    // Setup chat input handlers
    const chatInput = document.getElementById('chatInput');
    const sendButton = document.getElementById('sendButton');

    if (chatInput && sendButton) {
        sendButton.addEventListener('click', handleSendMessage);

        // Send on Enter (Shift+Enter for new line)
        chatInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSendMessage();
            }
        });
    }
});

// Handle sending a message
async function handleSendMessage() {
    const chatInput = document.getElementById('chatInput');
    if (!chatInput) return;

    const message = chatInput.value.trim();
    if (!message || isProcessing) return;

    console.log('Sending message:', message);

    // Add user message to chat
    addMessageToChat('user', message);

    // Clear input
    chatInput.value = '';

    // Show typing indicator
    showTypingIndicator();

    isProcessing = true;

    try {
        const response = await fetch(`${window.app.API_BASE}/ai/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                session_id: chatSessionId,
                message: message
            })
        });

        if (!response.ok) {
            throw new Error('Failed to get AI response');
        }

        const result = await response.json();
        console.log('AI response:', result);

        // Hide typing indicator
        hideTypingIndicator();

        // Handle function call results with special formatting
        if (result.function_call && result.data) {
            handleFunctionCallDisplay(result.function_call, result.data);
        } else if (result.message) {
            // Check if message contains JSON - if so, don't display it
            const trimmedMessage = result.message.trim();

            // Check if the message contains a JSON object anywhere
            const containsJSON = trimmedMessage.includes('{"name":') ||
                                 trimmedMessage.includes('{name:') ||
                                 (trimmedMessage.startsWith('{') && trimmedMessage.includes('"parameters"'));

            if (containsJSON) {
                console.log('Detected JSON in message, filtering out');
                // Extract any text before the JSON
                const jsonStartIdx = trimmedMessage.indexOf('{');
                const textBeforeJSON = jsonStartIdx > 0 ? trimmedMessage.substring(0, jsonStartIdx).trim() : '';

                if (textBeforeJSON && textBeforeJSON.length > 0) {
                    // Show only the text before JSON
                    addMessageToChat('assistant', textBeforeJSON, result.data);
                } else {
                    // No text before JSON, show generic processing message
                    addMessageToChat('assistant', 'Sto elaborando la tua richiesta...');
                }
            } else {
                // Normal message, display it
                addMessageToChat('assistant', result.message, result.data);
            }
        }

    } catch (error) {
        console.error('Chat error:', error);
        hideTypingIndicator();
        addMessageToChat('assistant', 'Sorry, I encountered an error processing your request. Please try again.');
    } finally {
        isProcessing = false;
    }
}

// Add message to chat display
function addMessageToChat(role, message, data = null) {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${role}-message`;

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';

    // Format message with line breaks
    const formattedMessage = message.replace(/\n/g, '<br>');
    contentDiv.innerHTML = `<p>${formattedMessage}</p>`;

    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);

    // Scroll to bottom
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Export addMessageToChat to global scope
window.addMessageToChat = addMessageToChat;

// Show typing indicator
function showTypingIndicator() {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const typingDiv = document.createElement('div');
    typingDiv.id = 'typing-indicator';
    typingDiv.className = 'message assistant-message';
    typingDiv.innerHTML = `
        <div class="message-content">
            <div class="typing-indicator">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
            </div>
        </div>
    `;

    chatMessages.appendChild(typingDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Hide typing indicator
function hideTypingIndicator() {
    const typingIndicator = document.getElementById('typing-indicator');
    if (typingIndicator) {
        typingIndicator.remove();
    }
}

// Handle function call display with interactive cards
function handleFunctionCallDisplay(functionCall, data) {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const functionName = functionCall.name;

    if (functionName === 'search_trains' && Array.isArray(data)) {
        // Display train results as interactive cards
        displayAITrainResults(data);
    } else if (functionName === 'create_booking' && data) {
        // Display booking confirmation
        displayAIBookingConfirmation(data);
    } else if (functionName === 'book_train_direct' && data && data.booking) {
        // Display booking confirmation for direct booking
        displayAIBookingConfirmation(data.booking);
    } else if (functionName === 'get_booking_details' && data) {
        // Display booking details
        displayAIBookingDetails(data);
    }
}

// Display train results from AI
function displayAITrainResults(trains) {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    messageDiv.className = 'message assistant-message';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.style.maxWidth = '95%';

    let html = '<div style="margin-top: 10px;">';

    trains.forEach((result, index) => {
        const schedule = result.schedule;
        const train = schedule.train;

        html += `
            <div class="train-card" style="margin-bottom: 15px;">
                <div class="train-header">
                    <div class="train-number">${train.number}</div>
                    <div class="train-type">${train.type}</div>
                </div>
                <div class="train-details">
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
                        <span class="detail-label">Price:</span> ${window.app.formatCurrency(result.total_price)}
                    </div>
                </div>
                <div class="train-amenities">
                    <div class="amenity">${train.has_wifi ? '✅' : '❌'} WiFi</div>
                    <div class="amenity">${train.has_food ? '✅' : '❌'} Food</div>
                </div>
                <div style="margin-top: 10px; font-size: 0.875rem; color: var(--text-gray);">
                    Schedule ID: ${schedule.id}
                </div>
                <button class="btn btn-primary book-train-btn"
                    data-schedule-id="${schedule.id}"
                    data-origin="${schedule.origin.name}"
                    data-destination="${schedule.destination.name}"
                    data-departure="${result.departure_time}"
                    style="width: 100%; margin-top: 10px;">
                    Book This Train
                </button>
            </div>
        `;
    });

    html += '</div>';
    contentDiv.innerHTML = html;

    // Add event listeners to all book buttons
    setTimeout(() => {
        const bookButtons = contentDiv.querySelectorAll('.book-train-btn');
        bookButtons.forEach(btn => {
            btn.addEventListener('click', function() {
                const scheduleId = this.getAttribute('data-schedule-id');
                const origin = this.getAttribute('data-origin');
                const destination = this.getAttribute('data-destination');
                const departure = this.getAttribute('data-departure');
                window.bookTrainFromAI(scheduleId, origin, destination, departure);
            });
        });
    }, 0);

    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Display booking confirmation from AI
function displayAIBookingConfirmation(booking) {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    messageDiv.className = 'message assistant-message';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.style.maxWidth = '95%';

    const schedule = booking.schedule;
    const train = schedule.train;

    let html = `
        <div class="booking-confirmation" style="margin-top: 10px;">
            <h4 style="color: var(--success-color); margin-bottom: 15px;">✅ Booking Confirmed!</h4>

            <p><strong>Booking Reference:</strong></p>
            <p class="booking-ref" style="font-size: 1.5rem; color: var(--success-color); margin: 10px 0;">
                ${booking.booking_ref}
            </p>

            <div style="margin-top: 15px;">
                <p><strong>Train:</strong> ${train.number} (${train.type})</p>
                <p><strong>Route:</strong> ${schedule.origin.name} → ${schedule.destination.name}</p>
                <p><strong>Date:</strong> ${window.app.formatDate(booking.booking_date)}</p>
                <p><strong>Departure:</strong> ${window.app.formatTime(schedule.departure_time)}</p>
                <p><strong>Arrival:</strong> ${window.app.formatTime(schedule.arrival_time)}</p>
            </div>

            <div style="margin-top: 15px;">
                <p><strong>Passengers:</strong></p>
                <ul style="margin: 5px 0; padding-left: 20px;">
                    ${booking.passengers.map(p => `
                        <li>${p.name} (${p.passenger_type}) - ${p.seat_number || 'No seat'} - ${window.app.formatCurrency(p.price)}</li>
                    `).join('')}
                </ul>
            </div>

            <div style="margin-top: 15px; padding-top: 15px; border-top: 1px solid var(--border-color);">
                <p style="font-size: 1.25rem;">
                    <strong>Total Price:</strong>
                    <span style="color: var(--success-color);">${window.app.formatCurrency(booking.total_price)}</span>
                </p>
            </div>

            <div class="alert alert-info" style="margin-top: 15px; font-size: 0.9rem;">
                <p><strong>Important:</strong> Save your booking reference: <strong>${booking.booking_ref}</strong></p>
            </div>
        </div>
    `;

    contentDiv.innerHTML = html;
    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Display booking details from AI
function displayAIBookingDetails(booking) {
    const chatMessages = document.getElementById('chatMessages');
    if (!chatMessages) return;

    const messageDiv = document.createElement('div');
    messageDiv.className = 'message assistant-message';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.style.maxWidth = '95%';

    const schedule = booking.schedule;
    const train = schedule.train;
    const statusColor = booking.status === 'confirmed' ? 'var(--success-color)' : 'var(--danger-color)';

    let html = `
        <div style="margin-top: 10px; border: 2px solid var(--border-color); border-radius: 8px; padding: 15px;">
            <h4 style="margin-bottom: 15px;">📋 Booking Details</h4>

            <p><strong>Reference:</strong> ${booking.booking_ref}</p>
            <p><strong>Status:</strong> <span style="color: ${statusColor}; font-weight: bold;">${booking.status.toUpperCase()}</span></p>

            <div style="margin-top: 15px;">
                <p><strong>Train:</strong> ${train.number} (${train.type})</p>
                <p><strong>Route:</strong> ${schedule.origin.name} → ${schedule.destination.name}</p>
                <p><strong>Date:</strong> ${window.app.formatDate(booking.booking_date)}</p>
                <p><strong>Departure:</strong> ${window.app.formatTime(schedule.departure_time)}</p>
                <p><strong>Arrival:</strong> ${window.app.formatTime(schedule.arrival_time)}</p>
            </div>

            <div style="margin-top: 15px;">
                <p><strong>Passengers (${booking.passenger_count}):</strong></p>
                <ul style="margin: 5px 0; padding-left: 20px;">
                    ${booking.passengers.map(p => `
                        <li>${p.name} (${p.passenger_type}) - Seat: ${p.seat_number || 'N/A'}</li>
                    `).join('')}
                </ul>
            </div>

            <div style="margin-top: 15px; padding-top: 15px; border-top: 1px solid var(--border-color);">
                <p style="font-size: 1.1rem;">
                    <strong>Total Price:</strong> ${window.app.formatCurrency(booking.total_price)}
                </p>
            </div>
        </div>
    `;

    contentDiv.innerHTML = html;
    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Clear chat history (optional feature)
window.clearChat = function() {
    const chatMessages = document.getElementById('chatMessages');
    if (chatMessages) {
        chatMessages.innerHTML = '';
        // Re-add welcome message
        addMessageToChat('assistant',
            `Hello! I'm your AI booking assistant. I can help you find and book train tickets.\n\n` +
            `Try saying something like:\n` +
            `• "I need to go from Milan to Rome tomorrow morning"\n` +
            `• "Show me trains from Venice to Florence next week"\n` +
            `• "Book 2 tickets on the 7:30 train"`
        );
    }
    // Generate new session
    chatSessionId = window.app.generateUUID();
    console.log('New chat session:', chatSessionId);
};

// Flag to track if we're in AI booking mode
window.isAIBookingMode = false;

// Book train from AI chat - use modal form like traditional booking
window.bookTrainFromAI = function(scheduleId, origin, destination, departureTime) {
    console.log('Booking train from AI:', scheduleId);

    // Store booking context
    window.aiBookingContext = {
        scheduleId: scheduleId,
        origin: origin,
        destination: destination,
        departureTime: departureTime
    };

    // Set AI booking mode flag
    window.isAIBookingMode = true;

    // Open the booking modal with a simple form
    showAIBookingModal();
};

// Show AI booking modal
function showAIBookingModal() {
    const modal = document.getElementById('bookingModal');
    const passengerInputs = document.getElementById('passengerInputs');

    // Create simple passenger input form
    passengerInputs.innerHTML = `
        <div class="form-group">
            <label for="aiPassengerCount">Number of Passengers:</label>
            <input type="number" id="aiPassengerCount" value="1" min="1" max="9" class="form-control">
        </div>
        <div id="aiPassengerDetails"></div>
        <div class="form-group">
            <label for="aiBookingDate">Travel Date:</label>
            <input type="date" id="aiBookingDate" class="form-control" required>
        </div>
    `;

    // Set default date to tomorrow
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    document.getElementById('aiBookingDate').value = tomorrow.toISOString().split('T')[0];

    // Generate passenger fields
    updateAIPassengerFields();

    // Listen for passenger count changes
    document.getElementById('aiPassengerCount').addEventListener('change', updateAIPassengerFields);

    modal.classList.add('show');

    // Set up close handlers for AI booking
    const closeBtn = modal.querySelector('.modal-close');
    const cancelBtn = modal.querySelector('#cancelBooking');

    const closeModal = () => {
        modal.classList.remove('show');
        window.isAIBookingMode = false;
    };

    if (closeBtn) {
        closeBtn.onclick = closeModal;
    }

    if (cancelBtn) {
        cancelBtn.onclick = closeModal;
    }

    // Close on outside click
    modal.onclick = (e) => {
        if (e.target === modal) {
            closeModal();
        }
    };
}

// Update passenger input fields
function updateAIPassengerFields() {
    const count = parseInt(document.getElementById('aiPassengerCount').value) || 1;
    const container = document.getElementById('aiPassengerDetails');

    let html = '<h4 style="margin-top: 20px;">Passenger Details:</h4>';

    for (let i = 0; i < count; i++) {
        html += `
            <div class="passenger-group" style="border: 1px solid var(--border-color); padding: 15px; margin: 10px 0; border-radius: 8px;">
                <h5>Passenger ${i + 1}</h5>
                <div class="form-group">
                    <label for="aiPassengerName${i}">Full Name:</label>
                    <input type="text" id="aiPassengerName${i}" class="form-control" required placeholder="John Doe">
                </div>
                <div class="form-group">
                    <label for="aiPassengerType${i}">Type:</label>
                    <select id="aiPassengerType${i}" class="form-control">
                        <option value="adult">Adult</option>
                        <option value="senior">Senior (65+)</option>
                        <option value="child">Child (4-17)</option>
                        <option value="infant">Infant (0-3)</option>
                    </select>
                </div>
            </div>
        `;
    }

    container.innerHTML = html;
}
