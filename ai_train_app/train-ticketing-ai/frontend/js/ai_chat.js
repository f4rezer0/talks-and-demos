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

        // Add AI response to chat
        if (result.message) {
            addMessageToChat('assistant', result.message, result.data);
        }

        // Handle function call results with special formatting
        if (result.function_call && result.data) {
            handleFunctionCallDisplay(result.function_call, result.data);
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
            </div>
        `;
    });

    html += '</div>';
    contentDiv.innerHTML = html;

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
